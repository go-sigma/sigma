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

package dao

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

//go:generate mockgen -destination=mocks/user.go -package=mocks github.com/go-sigma/sigma/pkg/dal/dao UserService
//go:generate mockgen -destination=mocks/user_factory.go -package=mocks github.com/go-sigma/sigma/pkg/dal/dao UserServiceFactory

// UserService is the interface that provides the user service methods.
type UserService interface {
	// Get get user by id.
	Get(ctx context.Context, id int64) (*models.User, error)
	// GetByUsername gets the user with the specified user name.
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	// Create creates a new user.
	Create(ctx context.Context, user *models.User) error
	// CreateUser3rdParty create a new 3rdparty user.
	CreateUser3rdParty(ctx context.Context, user3rdParty *models.User3rdParty) error
	// UpdateUser3rdParty update 3rdParty user
	UpdateUser3rdParty(ctx context.Context, id int64, updates map[string]any) error
	// List all users with pagination
	List(ctx context.Context, name *string, pagination types.Pagination, sort types.Sortable) ([]*models.User, int64, error)
	// UpdateByID updates the namespace with the specified namespace ID.
	UpdateByID(ctx context.Context, id int64, updates map[string]interface{}) error
	// Count gets the total number of users.
	Count(ctx context.Context) (int64, error)
	// GetByProvider gets the user with the specified oauth2 provider.
	GetByProvider(ctx context.Context, provider enums.Provider, accountID string) (*models.User3rdParty, error)
	// GetUser3rdParty gets the user 3rdparty with the specified 3rdparty userid
	GetUser3rdParty(ctx context.Context, userID int64) (*models.User3rdParty, error)
	// ListUser3rdParty gets the user 3rdparty with the specified 3rdparty userid
	ListUser3rdParty(ctx context.Context, userID int64) ([]*models.User3rdParty, error)
	// GetRecoverCodeByUserID gets the recover code with the specified user id.
	GetRecoverCodeByUserID(ctx context.Context, userID int64) (*models.UserRecoverCode, error)
	// GetByRecoverCode gets the user with the specified recover code.
	GetByRecoverCode(ctx context.Context, code string) (*models.User, error)
	// CreateRecoverCode creates a new recover code.
	CreateRecoverCode(ctx context.Context, recoverCode *models.UserRecoverCode) error
	// DeleteRecoverCode deletes the recover code with the specified user id.
	DeleteRecoverCode(ctx context.Context, userID int64) error
}

type userService struct {
	tx *query.Query
}

// UserServiceFactory is the interface that provides the user service factory methods.
type UserServiceFactory interface {
	New(txs ...*query.Query) UserService
}

type userServiceFactory struct{}

// NewUserServiceFactory creates a new user service factory.
func NewUserServiceFactory() UserServiceFactory {
	return &userServiceFactory{}
}

// New creates a new user service.
func (f *userServiceFactory) New(txs ...*query.Query) UserService {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &userService{
		tx: tx,
	}
}

var _ UserServiceFactory = &userServiceFactory{}

// Get get user by id.
func (s *userService) Get(ctx context.Context, id int64) (*models.User, error) {
	return s.tx.User.WithContext(ctx).Where(s.tx.User.ID.Eq(id)).First()
}

// GetByUsername gets the user with the specified user name.
func (s *userService) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	return s.tx.User.WithContext(ctx).Where(s.tx.User.Username.Eq(username)).First()
}

// Create creates a new user.
func (s *userService) Create(ctx context.Context, user *models.User) error {
	return s.tx.User.WithContext(ctx).Create(user)
}

// CreateUser3rdParty create a new 3rdparty user.
func (s *userService) CreateUser3rdParty(ctx context.Context, user3rdParty *models.User3rdParty) error {
	return s.tx.User3rdParty.WithContext(ctx).Create(user3rdParty)
}

// UpdateUser3rdParty update 3rdParty user
func (s *userService) UpdateUser3rdParty(ctx context.Context, id int64, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	matched, err := s.tx.User3rdParty.WithContext(ctx).Where(s.tx.User3rdParty.ID.Eq(id)).Updates(updates)
	if err != nil {
		return err
	}
	if matched.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// List all users with pagination
func (s *userService) List(ctx context.Context, name *string, pagination types.Pagination, sort types.Sortable) ([]*models.User, int64, error) {
	pagination = utils.NormalizePagination(pagination)
	query := s.tx.User.WithContext(ctx)
	if name != nil {
		query = query.Where(s.tx.User.Username.Like(fmt.Sprintf("%%%s%%", ptr.To(name))))
	}
	field, ok := s.tx.User.GetFieldByName(ptr.To(sort.Sort))
	if ok {
		switch ptr.To(sort.Method) {
		case enums.SortMethodDesc:
			query = query.Order(field.Desc())
		case enums.SortMethodAsc:
			query = query.Order(field)
		default:
			query = query.Order(s.tx.User.UpdatedAt.Desc())
		}
	} else {
		query = query.Order(s.tx.User.UpdatedAt.Desc())
	}
	return query.FindByPage(ptr.To(pagination.Limit)*(ptr.To(pagination.Page)-1), ptr.To(pagination.Limit))
}

// Count gets the total number of users.
func (s *userService) Count(ctx context.Context) (int64, error) {
	return s.tx.User.WithContext(ctx).Count()
}

// GetByProvider gets the user with the specified oauth2 provider.
func (s *userService) GetByProvider(ctx context.Context, provider enums.Provider, accountID string) (*models.User3rdParty, error) {
	return s.tx.User3rdParty.WithContext(ctx).
		Where(s.tx.User3rdParty.Provider.Eq(provider), s.tx.User3rdParty.AccountID.Eq(accountID)).
		Preload(s.tx.User3rdParty.User).First()
}

// GetUser3rdParty gets the user 3rdparty with the specified 3rdparty userid
func (s *userService) GetUser3rdParty(ctx context.Context, userID int64) (*models.User3rdParty, error) {
	return s.tx.User3rdParty.WithContext(ctx).Where(s.tx.User3rdParty.ID.Eq(userID)).First()
}

// ListUser3rdParty gets the user 3rdparty with the specified 3rdparty userid
func (s *userService) ListUser3rdParty(ctx context.Context, userID int64) ([]*models.User3rdParty, error) {
	return s.tx.User3rdParty.WithContext(ctx).Where(s.tx.User3rdParty.UserID.Eq(userID)).Find()
}

// GetRecoverCodeByUserID gets the recover code with the specified user id.
func (s *userService) GetRecoverCodeByUserID(ctx context.Context, userID int64) (*models.UserRecoverCode, error) {
	return s.tx.UserRecoverCode.WithContext(ctx).Where(s.tx.UserRecoverCode.UserID.Eq(userID)).First()
}

// CreateRecoverCode creates a new recover code.
func (s *userService) CreateRecoverCode(ctx context.Context, recoverCode *models.UserRecoverCode) error {
	return s.tx.UserRecoverCode.WithContext(ctx).Create(recoverCode)
}

// DeleteRecoverCode deletes the recover code with the specified user id.
func (s *userService) DeleteRecoverCode(ctx context.Context, userID int64) error {
	_, err := s.tx.UserRecoverCode.WithContext(ctx).Where(s.tx.UserRecoverCode.UserID.Eq(userID)).Delete()
	return err
}

// GetByRecoverCode gets the user with the specified recover code.
func (s *userService) GetByRecoverCode(ctx context.Context, code string) (*models.User, error) {
	recoverCode, err := s.tx.UserRecoverCode.WithContext(ctx).Where(s.tx.UserRecoverCode.Code.Eq(code)).First()
	if err != nil {
		return nil, err
	}
	return s.tx.User.WithContext(ctx).Where(s.tx.User.ID.Eq(recoverCode.UserID)).First()
}

// UpdateByID updates the namespace with the specified namespace ID.
func (s *userService) UpdateByID(ctx context.Context, id int64, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}
	matched, err := s.tx.User.WithContext(ctx).Where(s.tx.User.ID.Eq(id)).Updates(updates)
	if err != nil {
		return err
	}
	if matched.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
