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

package user

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

//go:generate go tool mockgen -destination=user_mocks.go -package=user github.com/go-sigma/sigma/pkg/dal/repository/user UserRepository

// UserRepository defines user repository operations
type UserRepository interface {
	// Get gets a user by ID
	Get(ctx context.Context, id string) (*models.User, error)
	// GetByUsername gets a user by username
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	// Create creates a new user
	Create(ctx context.Context, user *models.User) error
	// CreateUser3rdParty creates a third-party user binding
	CreateUser3rdParty(ctx context.Context, user3rdParty *models.User3rdParty) error
	// UpdateUser3rdParty updates a third-party user binding
	UpdateUser3rdParty(ctx context.Context, id string, updates map[string]any) error
	// List lists users with filtering, pagination, and sorting
	List(ctx context.Context, name *string, pagination api.Pagination, sort api.Sortable) ([]*models.User, int64, error)
	// ListWithoutUsername lists users excluding usernames and optionally admin users
	ListWithoutUsername(ctx context.Context, except []string, withoutAdmin bool, name *string, pagination api.Pagination, sort api.Sortable) ([]*models.User, int64, error)
	// UpdateByID updates the user with the specified user ID
	UpdateByID(ctx context.Context, id string, updates map[string]any) error
	// Count gets the total number of users
	Count(ctx context.Context) (int64, error)
	// GetUser3rdPartyByAccountID gets a third-party user binding by provider and account ID
	GetUser3rdPartyByAccountID(ctx context.Context, provider enums.Provider, accountID string) (*models.User3rdParty, error)
	// GetUser3rdPartyByProvider gets a third-party user binding by user ID and provider
	GetUser3rdPartyByProvider(ctx context.Context, userID string, provider enums.Provider) (*models.User3rdParty, error)
	// GetUser3rdParty gets a third-party user binding by ID
	GetUser3rdParty(ctx context.Context, user3rdPartyID string) (*models.User3rdParty, error)
	// ListUser3rdParty lists third-party user bindings for a user
	ListUser3rdParty(ctx context.Context, userID string) ([]*models.User3rdParty, error)
	// GetRecoverCodeByUserID gets the recover code with the specified user ID
	GetRecoverCodeByUserID(ctx context.Context, userID string) (*models.UserRecoverCode, error)
	// GetByRecoverCode gets the user with the specified recover code
	GetByRecoverCode(ctx context.Context, code string) (*models.User, error)
	// CreateRecoverCode creates a new recover code
	CreateRecoverCode(ctx context.Context, recoverCode *models.UserRecoverCode) error
	// DeleteRecoverCode deletes the recover code with the specified user ID
	DeleteRecoverCode(ctx context.Context, userID string) error
}

type userRepository struct {
	tx *query.Query
}

// NewUserRepository creates a new user repository with the optional query transaction
func NewUserRepository(txs ...*query.Query) UserRepository {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &userRepository{
		tx: tx,
	}
}

// Get gets a user by ID
func (s *userRepository) Get(ctx context.Context, id string) (*models.User, error) {
	return s.tx.User.WithContext(ctx).Where(s.tx.User.ID.Eq(id)).First()
}

// GetByUsername gets a user by username
func (s *userRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	return s.tx.User.WithContext(ctx).Where(s.tx.User.Username.Eq(username)).First()
}

// Create creates a new user
func (s *userRepository) Create(ctx context.Context, user *models.User) error {
	return s.tx.User.WithContext(ctx).Create(user)
}

// CreateUser3rdParty creates a third-party user binding
func (s *userRepository) CreateUser3rdParty(ctx context.Context, user3rdParty *models.User3rdParty) error {
	return s.tx.User3rdParty.WithContext(ctx).Create(user3rdParty)
}

// UpdateUser3rdParty updates a third-party user binding
func (s *userRepository) UpdateUser3rdParty(ctx context.Context, id string, updates map[string]any) error {
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

// ListWithoutUsername lists users excluding usernames and optionally admin users
func (s *userRepository) ListWithoutUsername(ctx context.Context, except []string, withoutAdmin bool, name *string, pagination api.Pagination, sort api.Sortable) ([]*models.User, int64, error) {
	pagination = utils.NormalizePagination(pagination)
	q := s.tx.User.WithContext(ctx)
	if len(except) > 0 {
		q = q.Where(s.tx.User.Username.NotIn(except...))
	}
	if withoutAdmin {
		q = q.Where(s.tx.User.Role.Neq(enums.UserRoleAdmin), s.tx.User.Role.Neq(enums.UserRoleRoot))
	}
	if name != nil {
		q = q.Where(s.tx.User.Username.Like(fmt.Sprintf("%s%%", ptr.To(name))))
	}
	field, ok := s.tx.User.GetFieldByName(ptr.To(sort.Sort))
	if ok {
		switch ptr.To(sort.Method) {
		case enums.SortMethodDesc:
			q = q.Order(field.Desc())
		case enums.SortMethodAsc:
			q = q.Order(field)
		default:
			q = q.Order(s.tx.User.UpdatedAt.Desc())
		}
	} else {
		q = q.Order(s.tx.User.UpdatedAt.Desc())
	}
	return q.FindByPage(ptr.To(pagination.Limit)*(ptr.To(pagination.Page)-1), ptr.To(pagination.Limit))
}

// List lists users with filtering, pagination, and sorting
func (s *userRepository) List(ctx context.Context, name *string, pagination api.Pagination, sort api.Sortable) ([]*models.User, int64, error) {
	pagination = utils.NormalizePagination(pagination)
	q := s.tx.User.WithContext(ctx)
	if name != nil {
		q = q.Where(s.tx.User.Username.Like(fmt.Sprintf("%%%s%%", ptr.To(name))))
	}
	field, ok := s.tx.User.GetFieldByName(ptr.To(sort.Sort))
	if ok {
		switch ptr.To(sort.Method) {
		case enums.SortMethodDesc:
			q = q.Order(field.Desc())
		case enums.SortMethodAsc:
			q = q.Order(field)
		default:
			q = q.Order(s.tx.User.UpdatedAt.Desc())
		}
	} else {
		q = q.Order(s.tx.User.UpdatedAt.Desc())
	}
	return q.FindByPage(ptr.To(pagination.Limit)*(ptr.To(pagination.Page)-1), ptr.To(pagination.Limit))
}

// Count gets the total number of users
func (s *userRepository) Count(ctx context.Context) (int64, error) {
	return s.tx.User.WithContext(ctx).Count()
}

// GetUser3rdPartyByAccountID gets a third-party user binding by provider and account ID
func (s *userRepository) GetUser3rdPartyByAccountID(ctx context.Context, provider enums.Provider, accountID string) (*models.User3rdParty, error) {
	return s.tx.User3rdParty.WithContext(ctx).
		Where(s.tx.User3rdParty.Provider.Eq(provider), s.tx.User3rdParty.AccountID.Eq(accountID)).
		Preload(s.tx.User3rdParty.User).First()
}

// GetUser3rdPartyByProvider gets a third-party user binding by user ID and provider
func (s *userRepository) GetUser3rdPartyByProvider(ctx context.Context, userID string, provider enums.Provider) (*models.User3rdParty, error) {
	return s.tx.User3rdParty.WithContext(ctx).Where(
		s.tx.User3rdParty.UserID.Eq(userID), s.tx.User3rdParty.Provider.Eq(provider)).
		Preload(s.tx.User3rdParty.User).First()
}

// GetUser3rdParty gets a third-party user binding by ID
func (s *userRepository) GetUser3rdParty(ctx context.Context, user3rdPartyID string) (*models.User3rdParty, error) {
	return s.tx.User3rdParty.WithContext(ctx).Where(s.tx.User3rdParty.ID.Eq(user3rdPartyID)).First()
}

// ListUser3rdParty lists third-party user bindings for a user
func (s *userRepository) ListUser3rdParty(ctx context.Context, userID string) ([]*models.User3rdParty, error) {
	return s.tx.User3rdParty.WithContext(ctx).Where(s.tx.User3rdParty.UserID.Eq(userID)).Find()
}

// GetRecoverCodeByUserID gets the recover code with the specified user ID
func (s *userRepository) GetRecoverCodeByUserID(ctx context.Context, userID string) (*models.UserRecoverCode, error) {
	return s.tx.UserRecoverCode.WithContext(ctx).Where(s.tx.UserRecoverCode.UserID.Eq(userID)).First()
}

// CreateRecoverCode creates a new recover code
func (s *userRepository) CreateRecoverCode(ctx context.Context, recoverCode *models.UserRecoverCode) error {
	return s.tx.UserRecoverCode.WithContext(ctx).Create(recoverCode)
}

// DeleteRecoverCode deletes the recover code with the specified user ID
func (s *userRepository) DeleteRecoverCode(ctx context.Context, userID string) error {
	_, err := s.tx.UserRecoverCode.WithContext(ctx).Where(s.tx.UserRecoverCode.UserID.Eq(userID)).Delete()
	return err
}

// GetByRecoverCode gets the user with the specified recover code
func (s *userRepository) GetByRecoverCode(ctx context.Context, code string) (*models.User, error) {
	recoverCode, err := s.tx.UserRecoverCode.WithContext(ctx).Where(s.tx.UserRecoverCode.Code.Eq(code)).First()
	if err != nil {
		return nil, err
	}
	return s.tx.User.WithContext(ctx).Where(s.tx.User.ID.Eq(recoverCode.UserID)).First()
}

// UpdateByID updates the user with the specified user ID
func (s *userRepository) UpdateByID(ctx context.Context, id string, updates map[string]any) error {
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
