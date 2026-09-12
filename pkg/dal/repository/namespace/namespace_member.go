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

package namespace

import (
	"context"
	"fmt"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

//go:generate go tool mockgen -destination=namespace_member_mocks.go -package=namespace github.com/go-sigma/sigma/pkg/dal/repository/namespace NamespaceMemberRepository

// NamespaceMemberRepository defines namespace member repository operations
type NamespaceMemberRepository interface {
	// AddNamespaceMember adds a user to a namespace with the specified role
	AddNamespaceMember(ctx context.Context, userID string, namespaceObj models.Namespace, role enums.NamespaceRole) (*models.NamespaceMember, error)
	// UpdateNamespaceMember updates a user's role in a namespace
	UpdateNamespaceMember(ctx context.Context, userID string, namespaceObj models.Namespace, role enums.NamespaceRole) error
	// DeleteNamespaceMember removes a user from a namespace
	DeleteNamespaceMember(ctx context.Context, userID string, namespaceObj models.Namespace) error
	// ListNamespaceMembers lists namespace members with filtering, pagination, and sorting
	ListNamespaceMembers(ctx context.Context, namespaceID string, name *string, pagination api.Pagination, sort api.Sortable) ([]*models.NamespaceMember, int64, error)
	// GetNamespaceMember gets a namespace member by namespace and user ID
	GetNamespaceMember(ctx context.Context, namespaceID string, userID string) (*models.NamespaceMember, error)
	// GetNamespacesMember gets namespace memberships for the specified namespaces and user
	GetNamespacesMember(ctx context.Context, namespaceIDs []string, userID string) ([]*models.NamespaceMember, error)
	// CountNamespaceMember counts matching namespace memberships for a user
	CountNamespaceMember(ctx context.Context, userID string, namespaceID string) (int64, error)
}

var _ NamespaceMemberRepository = &namespaceMemberRepository{}

type namespaceMemberRepository struct {
	tx *query.Query
}

// NewNamespaceMemberRepository creates a new namespace member repository with the optional query transaction
func NewNamespaceMemberRepository(txs ...*query.Query) NamespaceMemberRepository {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &namespaceMemberRepository{
		tx: tx,
	}
}

// AddNamespaceMember adds a user to a namespace with the specified role
func (s namespaceMemberRepository) AddNamespaceMember(ctx context.Context, userID string, namespaceObj models.Namespace, role enums.NamespaceRole) (*models.NamespaceMember, error) {
	namespaceMember := &models.NamespaceMember{ID: uuid.NewV7String(), UserID: userID, NamespaceID: namespaceObj.ID, Role: role}
	err := s.tx.NamespaceMember.WithContext(ctx).Create(namespaceMember)
	if err != nil {
		return nil, err
	}
	return namespaceMember, nil
}

// UpdateNamespaceMember updates a user's role in a namespace
func (s namespaceMemberRepository) UpdateNamespaceMember(ctx context.Context, userID string, namespaceObj models.Namespace, role enums.NamespaceRole) error {
	_, err := s.tx.NamespaceMember.WithContext(ctx).Where(
		s.tx.NamespaceMember.UserID.Eq(userID),
		s.tx.NamespaceMember.NamespaceID.Eq(namespaceObj.ID),
	).Updates(map[string]any{
		query.NamespaceMember.Role.ColumnName().String(): role,
	})
	return err
}

// DeleteNamespaceMember removes a user from a namespace
func (s namespaceMemberRepository) DeleteNamespaceMember(ctx context.Context, userID string, namespaceObj models.Namespace) error {
	_, err := s.tx.NamespaceMember.WithContext(ctx).Where(
		s.tx.NamespaceMember.UserID.Eq(userID),
		s.tx.NamespaceMember.NamespaceID.Eq(namespaceObj.ID),
	).Delete()
	return err
}

// ListNamespaceMembers lists namespace members with filtering, pagination, and sorting
func (s namespaceMemberRepository) ListNamespaceMembers(ctx context.Context, namespaceID string, name *string, pagination api.Pagination, sort api.Sortable) ([]*models.NamespaceMember, int64, error) {
	pagination = utils.NormalizePagination(pagination)
	q := s.tx.NamespaceMember.WithContext(ctx).Where(s.tx.NamespaceMember.NamespaceID.Eq(namespaceID))
	if name != nil {
		q = q.RightJoin(s.tx.User, s.tx.NamespaceMember.UserID.EqCol(s.tx.User.ID), s.tx.User.Username.Like(fmt.Sprintf("%s%%", ptr.To(name))))
	}
	q = q.Preload(s.tx.NamespaceMember.User)
	field, ok := s.tx.NamespaceMember.GetFieldByName(ptr.To(sort.Sort))
	if ok {
		switch ptr.To(sort.Method) {
		case enums.SortMethodDesc:
			q = q.Order(field.Desc())
		case enums.SortMethodAsc:
			q = q.Order(field)
		default:
			q = q.Order(s.tx.NamespaceMember.UpdatedAt.Desc())
		}
	} else {
		q = q.Order(s.tx.NamespaceMember.UpdatedAt.Desc())
	}
	return q.FindByPage(ptr.To(pagination.Limit)*(ptr.To(pagination.Page)-1), ptr.To(pagination.Limit))
}

// GetNamespaceMember gets a namespace member by namespace and user ID
func (s namespaceMemberRepository) GetNamespaceMember(ctx context.Context, namespaceID string, userID string) (*models.NamespaceMember, error) {
	return s.tx.NamespaceMember.WithContext(ctx).Where(
		s.tx.NamespaceMember.UserID.Eq(userID),
		s.tx.NamespaceMember.NamespaceID.Eq(namespaceID),
	).First()
}

// GetNamespacesMember gets namespace memberships for the specified namespaces and user
func (s namespaceMemberRepository) GetNamespacesMember(ctx context.Context, namespaceIDs []string, userID string) ([]*models.NamespaceMember, error) {
	if len(namespaceIDs) == 0 {
		return nil, nil
	}
	return s.tx.NamespaceMember.WithContext(ctx).Where(
		s.tx.NamespaceMember.UserID.Eq(userID),
		s.tx.NamespaceMember.NamespaceID.In(namespaceIDs...),
	).Find()
}

// CountNamespaceMember counts matching namespace memberships for a user
func (s namespaceMemberRepository) CountNamespaceMember(ctx context.Context, userID string, namespaceID string) (int64, error) {
	return s.tx.NamespaceMember.WithContext(ctx).Where(
		s.tx.NamespaceMember.UserID.Eq(userID),
		s.tx.NamespaceMember.NamespaceID.Eq(namespaceID),
	).Count()
}
