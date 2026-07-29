// Copyright 2023 sigma
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package users

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/dig"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	repouser "github.com/go-sigma/sigma/pkg/dal/repository/user"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/service/password"
	"github.com/go-sigma/sigma/pkg/service/token"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

//go:generate mockgen -destination=users_mocks.go -package=users github.com/go-sigma/sigma/pkg/service/users UserService

// UserService encapsulates user-related business logic.
type UserService interface {
	// ListUsers lists users with pagination, excluding specified usernames.
	ListUsers(ctx context.Context, exceptUsernames []string, withoutAdmin bool, name *string, pagination api.Pagination, sort api.Sortable) ([]*models.User, int64, error)
	// Login updates the user's lastLogin timestamp and creates new access/refresh tokens.
	Login(ctx context.Context, userID string, ttl, refreshTTL time.Duration) (accessToken, refreshToken string, err error)
	// Logout validates and revokes the provided tokens along with the current jti.
	Logout(ctx context.Context, tokens []string, jti string) error
	// CreateUser creates a new user (admin operation, manages platform role).
	CreateUser(ctx context.Context, req api.PostUserRequest) error
	// UpdateUser updates a user (admin operation, manages platform role).
	UpdateUser(ctx context.Context, userID string, req api.PutUserRequest) error
	// Signup registers a new user and returns the new user along with tokens.
	Signup(ctx context.Context, req api.PostUserSignupRequest, ttl, refreshTTL time.Duration) (*models.User, string, string, error)
	// RecoverPassword creates a recover code for the user matching the username/email.
	RecoverPassword(ctx context.Context, req api.PostUserRecoverPasswordRequest) error
	// RecoverPasswordReset resets the user's password using a recover code.
	RecoverPasswordReset(ctx context.Context, req api.PostUserRecoverResetPasswordRequest) error
	// ResetPassword resets a user's password (admin operation).
	ResetPassword(ctx context.Context, userID string, password string) error
	// SelfUpdate updates the current user's profile.
	SelfUpdate(ctx context.Context, userID string, req api.PutUserSelfRequest) error
	// SelfResetPassword resets the current user's password.
	SelfResetPassword(ctx context.Context, userID string, password string) error
}

type userService struct {
	userRepository repouser.UserRepository
	passwordSvc    password.Service
	tokenSvc       token.Service
}

type ServiceParams struct {
	dig.In

	UserRepository repouser.UserRepository
	PasswordSvc    password.Service
	TokenSvc       token.Service
}

func NewService(digCon *dig.Container) error {
	return digCon.Provide(func(params ServiceParams) UserService {
		return &userService{
			userRepository: params.UserRepository,
			passwordSvc:    params.PasswordSvc,
			tokenSvc:       params.TokenSvc,
		}
	})
}

func (s *userService) ListUsers(ctx context.Context, exceptUsernames []string, withoutAdmin bool, name *string, pagination api.Pagination, sort api.Sortable) ([]*models.User, int64, error) {
	userRepository := s.userRepository
	return userRepository.ListWithoutUsername(ctx, exceptUsernames, withoutAdmin, name, pagination, sort)
}

func (s *userService) Login(ctx context.Context, userID string, ttl, refreshTTL time.Duration) (string, string, error) {
	userRepository := s.userRepository
	err := userRepository.UpdateByID(ctx, userID, map[string]any{
		query.User.LastLogin.ColumnName().String(): time.Now().UnixMilli(),
	})
	if err != nil {
		slog.Error("update user last login failed", "err", err)
		return "", "", errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Update user last login failed: %v", err))
	}

	refreshToken, err := s.tokenSvc.New(userID, refreshTTL)
	if err != nil {
		slog.Error("create refresh token failed", "err", err)
		return "", "", errcode.HTTPErrCodeInternalError.Detail(err.Error())
	}

	accessToken, err := s.tokenSvc.New(userID, ttl)
	if err != nil {
		slog.Error("create token failed", "err", err)
		return "", "", errcode.HTTPErrCodeInternalError.Detail(err.Error())
	}

	return accessToken, refreshToken, nil
}

func (s *userService) Logout(ctx context.Context, tokens []string, jti string) error {
	ids := sets.New[string]()
	for _, t := range tokens {
		if t == "" {
			continue
		}
		id, _, err := s.tokenSvc.Validate(ctx, t)
		if err != nil {
			if errors.Is(err, token.ErrRevoked) || errors.Is(err, jwt.ErrTokenExpired) {
				continue
			}
			slog.Error("revoke token failed", "err", err, "token", t)
			return errcode.HTTPErrCodeInternalError.Detail(err.Error())
		}
		ids.Insert(id)
	}

	if jti == "" {
		return errcode.HTTPErrCodeUnauthorized.Detail("Get jti failed")
	}
	ids.Insert(jti)

	for {
		id, ok := ids.PopAny()
		if !ok {
			break
		}
		err := s.tokenSvc.Revoke(ctx, id)
		if err != nil {
			slog.Error("revoke token failed", "err", err)
			return errcode.HTTPErrCodeInternalError.Detail(err.Error())
		}
	}
	return nil
}

func (s *userService) CreateUser(ctx context.Context, req api.PostUserRequest) error {
	pwdHash, err := s.passwordSvc.Hash(req.Password)
	if err != nil {
		slog.Error("hash password failed", "err", err)
		return errcode.HTTPErrCodeInternalError.Detail(err.Error())
	}

	userObj := models.User{
		ID:             uuid.NewV7String(),
		Username:       req.Username,
		Password:       new(pwdHash),
		Email:          new(req.Email),
		Role:           req.Role,
		NamespaceLimit: ptr.To(req.NamespaceLimit),
	}

	err = query.Q.Transaction(func(tx *query.Query) error {
		userRepository := repouser.NewUserRepository(tx)
		if userObj.Role == enums.UserRoleAdmin {
			userObj.NamespaceLimit = 0
		}
		err = userRepository.Create(ctx, &userObj)
		if err != nil {
			slog.Error("create user failed", "err", err)
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create user failed: %v", err))
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *userService) UpdateUser(ctx context.Context, userID string, req api.PutUserRequest) error {
	userRepository := s.userRepository
	userObj, err := userRepository.Get(ctx, userID)
	if err != nil {
		slog.Error("get user failed", "err", err, "id", userID)
		return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get user failed: %v", err))
	}

	updates := make(map[string]any, 5)
	if req.Username != nil {
		updates[query.User.Username.ColumnName().String()] = ptr.To(req.Username)
	}
	if req.Email != nil {
		updates[query.User.Email.ColumnName().String()] = ptr.To(req.Email)
	}
	if req.Status != nil {
		updates[query.User.Status.ColumnName().String()] = req.Status
	}
	if req.Password != nil {
		pwdHash, err := s.passwordSvc.Hash(ptr.To(req.Password))
		if err != nil {
			slog.Error("hash password failed", "err", err)
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Hash password failed: %v", err))
		}
		updates[query.User.Password.ColumnName().String()] = pwdHash
	}
	if req.NamespaceLimit != nil {
		updates[query.User.NamespaceLimit.ColumnName().String()] = req.NamespaceLimit
	}
	err = query.Q.Transaction(func(tx *query.Query) error {
		userRepository := repouser.NewUserRepository(tx)
		err = userRepository.UpdateByID(ctx, userObj.ID, updates)
		if err != nil {
			slog.Error("update user failed", "err", err)
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Update user failed: %v", err))
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *userService) Signup(ctx context.Context, req api.PostUserSignupRequest, ttl, refreshTTL time.Duration) (*models.User, string, string, error) {
	pwdHash, err := s.passwordSvc.Hash(req.Password)
	if err != nil {
		slog.Error("hash password failed", "err", err)
		return nil, "", "", errcode.HTTPErrCodeInternalError.Detail(err.Error())
	}

	userRepository := s.userRepository
	_, err = userRepository.GetByUsername(ctx, req.Username)
	if err == nil {
		slog.Error("username already exists")
		return nil, "", "", errcode.HTTPErrCodeConflict.Detail("username already exists")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.Error("get user by username failed", "err", err)
		return nil, "", "", errcode.HTTPErrCodeInternalError.Detail(err.Error())
	}

	userObj := &models.User{
		ID:       uuid.NewV7String(),
		Username: req.Username,
		Password: new(pwdHash),
		Email:    new(req.Email),
	}
	err = userRepository.Create(ctx, userObj)
	if err != nil {
		slog.Error("create user failed", "err", err)
		return nil, "", "", errcode.HTTPErrCodeInternalError.Detail(err.Error())
	}

	refreshToken, err := s.tokenSvc.New(userObj.ID, refreshTTL)
	if err != nil {
		slog.Error("create refresh token failed", "err", err)
		return nil, "", "", errcode.HTTPErrCodeInternalError.Detail(err.Error())
	}

	accessToken, err := s.tokenSvc.New(userObj.ID, ttl)
	if err != nil {
		slog.Error("create token failed", "err", err)
		return nil, "", "", errcode.HTTPErrCodeInternalError.Detail(err.Error())
	}

	return userObj, accessToken, refreshToken, nil
}

func (s *userService) RecoverPassword(ctx context.Context, req api.PostUserRecoverPasswordRequest) error {
	userRepository := s.userRepository
	user, err := userRepository.GetByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("username not found", "err", err, "username", req.Username)
			return errcode.HTTPErrCodeNotFound.Detail("User or email not found")
		}
		slog.Error("username find failed", "err", err, "username", req.Username)
		return errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Username find failed: %v", err))
	}
	if ptr.To(user.Email) != req.Email {
		slog.Error("email not equal to real", "err", err, "username", req.Username, "realEmail", ptr.To(user.Email), "email", req.Email)
		return errcode.HTTPErrCodeNotFound.Detail("User or email not found")
	}

	_, err = userRepository.GetRecoverCodeByUserID(ctx, user.ID)
	if err == nil {
		slog.Error("recover code already exists", "err", err, "username", req.Username)
		return errcode.HTTPErrCodeConflict.Detail("Recover code already exists")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.Error("get recover code failed", "err", err, "username", req.Username)
		return errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get recover code failed: %v", err))
	}

	return query.Q.Transaction(func(tx *query.Query) error {
		userRepository := repouser.NewUserRepository(tx)
		err = userRepository.CreateRecoverCode(ctx, &models.UserRecoverCode{
			ID:     uuid.NewV7String(),
			UserID: user.ID,
			Code:   uuid.NewV7String(),
		})
		if err != nil {
			slog.Error("create recover code failed", "err", err, "username", req.Username)
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create recover code failed: %v", err))
		}
		return nil
	})
}

func (s *userService) RecoverPasswordReset(ctx context.Context, req api.PostUserRecoverResetPasswordRequest) error {
	userRepository := s.userRepository
	userObj, err := userRepository.GetByRecoverCode(ctx, req.Code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("recover code not found", "err", err, "code", req.Code)
			return errcode.HTTPErrCodeNotFound.Detail("Recover code not found")
		}
		slog.Error("get recover code failed", "err", err, "code", req.Code)
		return errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get recover code failed: %v", err))
	}

	return query.Q.Transaction(func(tx *query.Query) error {
		userRepository := repouser.NewUserRepository(tx)
		err = userRepository.DeleteRecoverCode(ctx, userObj.ID)
		if err != nil {
			slog.Error("delete recover code failed", "err", err, "code", req.Code)
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Delete recover code failed: %v", err))
		}
		pwdHash, err := s.passwordSvc.Hash(req.Password)
		if err != nil {
			slog.Error("hash password failed", "err", err)
			return errcode.HTTPErrCodeInternalError.Detail(err.Error())
		}
		err = userRepository.UpdateByID(ctx, userObj.ID, map[string]any{
			query.User.Password.ColumnName().String(): pwdHash,
		})
		if err != nil {
			slog.Error("update user failed", "err", err)
			return errcode.HTTPErrCodeInternalError.Detail(err.Error())
		}
		return nil
	})
}

func (s *userService) ResetPassword(ctx context.Context, userID string, password string) error {
	pwdHash, err := s.passwordSvc.Hash(password)
	if err != nil {
		slog.Error("hash password failed", "err", err)
		return errcode.HTTPErrCodeInternalError.Detail(err.Error())
	}
	return query.Q.Transaction(func(tx *query.Query) error {
		userRepository := repouser.NewUserRepository(tx)
		err = userRepository.UpdateByID(ctx, userID, map[string]any{
			query.User.Password.ColumnName().String(): pwdHash,
		})
		if err != nil {
			slog.Error("update user failed", "err", err)
			return errcode.HTTPErrCodeInternalError.Detail(err.Error())
		}
		return nil
	})
}

func (s *userService) SelfUpdate(ctx context.Context, userID string, req api.PutUserSelfRequest) error {
	updates := make(map[string]any, 5)
	if req.Username != nil {
		updates[query.User.Username.ColumnName().String()] = ptr.To(req.Username)
	}
	if req.Email != nil {
		updates[query.User.Email.ColumnName().String()] = ptr.To(req.Email)
	}
	return query.Q.Transaction(func(tx *query.Query) error {
		userRepository := repouser.NewUserRepository(tx)
		err := userRepository.UpdateByID(ctx, userID, updates)
		if err != nil {
			slog.Error("update user failed", "err", err)
			return errcode.HTTPErrCodeInternalError.Detail(err.Error())
		}
		return nil
	})
}

func (s *userService) SelfResetPassword(ctx context.Context, userID string, password string) error {
	return query.Q.Transaction(func(tx *query.Query) error {
		userRepository := repouser.NewUserRepository(tx)
		pwdHash, err := s.passwordSvc.Hash(password)
		if err != nil {
			slog.Error("hash password failed", "err", err)
			return errcode.HTTPErrCodeInternalError.Detail(err.Error())
		}
		err = userRepository.UpdateByID(ctx, userID, map[string]any{
			query.User.Password.ColumnName().String(): pwdHash,
		})
		if err != nil {
			slog.Error("update user failed", "err", err)
			return errcode.HTTPErrCodeInternalError.Detail(err.Error())
		}
		return nil
	})
}
