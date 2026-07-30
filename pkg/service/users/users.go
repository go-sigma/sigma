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

//go:generate mockgen -mock_names Service=MockUserService -destination=users_mocks.go -package=users github.com/go-sigma/sigma/pkg/service/users Service

// Service encapsulates user-related business logic.
type Service interface {
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

type service struct {
	dig.In

	RepoUser    repouser.UserRepository
	SvcPassword password.Service
	SvcToken    token.Service
}

func NewService(params service) Service {
	return &params
}

func (s *service) ListUsers(ctx context.Context, exceptUsernames []string, withoutAdmin bool, name *string, pagination api.Pagination, sort api.Sortable) ([]*models.User, int64, error) {
	return s.RepoUser.ListWithoutUsername(ctx, exceptUsernames, withoutAdmin, name, pagination, sort)
}

func (s *service) Login(ctx context.Context, userID string, ttl, refreshTTL time.Duration) (string, string, error) {
	err := s.RepoUser.UpdateByID(ctx, userID, map[string]any{
		query.User.LastLogin.ColumnName().String(): time.Now().UnixMilli(),
	})
	if err != nil {
		slog.Error("update user last login failed", "err", err)
		return "", "", errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Update user last login failed: %v", err))
	}

	refreshToken, err := s.SvcToken.New(userID, refreshTTL)
	if err != nil {
		slog.Error("create refresh token failed", "err", err)
		return "", "", errcode.HTTPErrCodeInternalError.Detail(err.Error())
	}

	accessToken, err := s.SvcToken.New(userID, ttl)
	if err != nil {
		slog.Error("create token failed", "err", err)
		return "", "", errcode.HTTPErrCodeInternalError.Detail(err.Error())
	}

	return accessToken, refreshToken, nil
}

func (s *service) Logout(ctx context.Context, tokens []string, jti string) error {
	ids := sets.New[string]()
	for _, t := range tokens {
		if t == "" {
			continue
		}
		id, _, err := s.SvcToken.Validate(ctx, t)
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
		err := s.SvcToken.Revoke(ctx, id)
		if err != nil {
			slog.Error("revoke token failed", "err", err)
			return errcode.HTTPErrCodeInternalError.Detail(err.Error())
		}
	}
	return nil
}

func (s *service) CreateUser(ctx context.Context, req api.PostUserRequest) error {
	pwdHash, err := s.SvcPassword.Hash(req.Password)
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

func (s *service) UpdateUser(ctx context.Context, userID string, req api.PutUserRequest) error {
	userObj, err := s.RepoUser.Get(ctx, userID)
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
		pwdHash, err := s.SvcPassword.Hash(ptr.To(req.Password))
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
		repoUser := repouser.NewUserRepository(tx)
		err = repoUser.UpdateByID(ctx, userObj.ID, updates)
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

func (s *service) Signup(ctx context.Context, req api.PostUserSignupRequest, ttl, refreshTTL time.Duration) (*models.User, string, string, error) {
	pwdHash, err := s.SvcPassword.Hash(req.Password)
	if err != nil {
		slog.Error("hash password failed", "err", err)
		return nil, "", "", errcode.HTTPErrCodeInternalError.Detail(err.Error())
	}

	_, err = s.RepoUser.GetByUsername(ctx, req.Username)
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
	err = s.RepoUser.Create(ctx, userObj)
	if err != nil {
		slog.Error("create user failed", "err", err)
		return nil, "", "", errcode.HTTPErrCodeInternalError.Detail(err.Error())
	}

	refreshToken, err := s.SvcToken.New(userObj.ID, refreshTTL)
	if err != nil {
		slog.Error("create refresh token failed", "err", err)
		return nil, "", "", errcode.HTTPErrCodeInternalError.Detail(err.Error())
	}

	accessToken, err := s.SvcToken.New(userObj.ID, ttl)
	if err != nil {
		slog.Error("create token failed", "err", err)
		return nil, "", "", errcode.HTTPErrCodeInternalError.Detail(err.Error())
	}

	return userObj, accessToken, refreshToken, nil
}

func (s *service) RecoverPassword(ctx context.Context, req api.PostUserRecoverPasswordRequest) error {
	user, err := s.RepoUser.GetByUsername(ctx, req.Username)
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

	_, err = s.RepoUser.GetRecoverCodeByUserID(ctx, user.ID)
	if err == nil {
		slog.Error("recover code already exists", "err", err, "username", req.Username)
		return errcode.HTTPErrCodeConflict.Detail("Recover code already exists")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.Error("get recover code failed", "err", err, "username", req.Username)
		return errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get recover code failed: %v", err))
	}

	return query.Q.Transaction(func(tx *query.Query) error {
		repoUser := repouser.NewUserRepository(tx)
		err = repoUser.CreateRecoverCode(ctx, &models.UserRecoverCode{
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

func (s *service) RecoverPasswordReset(ctx context.Context, req api.PostUserRecoverResetPasswordRequest) error {
	userObj, err := s.RepoUser.GetByRecoverCode(ctx, req.Code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("recover code not found", "err", err, "code", req.Code)
			return errcode.HTTPErrCodeNotFound.Detail("Recover code not found")
		}
		slog.Error("get recover code failed", "err", err, "code", req.Code)
		return errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get recover code failed: %v", err))
	}

	return query.Q.Transaction(func(tx *query.Query) error {
		repoUser := repouser.NewUserRepository(tx)
		err = repoUser.DeleteRecoverCode(ctx, userObj.ID)
		if err != nil {
			slog.Error("delete recover code failed", "err", err, "code", req.Code)
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Delete recover code failed: %v", err))
		}
		pwdHash, err := s.SvcPassword.Hash(req.Password)
		if err != nil {
			slog.Error("hash password failed", "err", err)
			return errcode.HTTPErrCodeInternalError.Detail(err.Error())
		}
		err = repoUser.UpdateByID(ctx, userObj.ID, map[string]any{
			query.User.Password.ColumnName().String(): pwdHash,
		})
		if err != nil {
			slog.Error("update user failed", "err", err)
			return errcode.HTTPErrCodeInternalError.Detail(err.Error())
		}
		return nil
	})
}

func (s *service) ResetPassword(ctx context.Context, userID string, password string) error {
	pwdHash, err := s.SvcPassword.Hash(password)
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

func (s *service) SelfUpdate(ctx context.Context, userID string, req api.PutUserSelfRequest) error {
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

func (s *service) SelfResetPassword(ctx context.Context, userID string, password string) error {
	return query.Q.Transaction(func(tx *query.Query) error {
		userRepository := repouser.NewUserRepository(tx)
		pwdHash, err := s.SvcPassword.Hash(password)
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
