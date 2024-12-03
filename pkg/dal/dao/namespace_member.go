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

	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

//go:generate mockgen -destination=mocks/namespace_member.go -package=mocks github.com/go-sigma/sigma/pkg/dal/dao NamespaceMemberService
//go:generate mockgen -destination=mocks/namespace_member_factory.go -package=mocks github.com/go-sigma/sigma/pkg/dal/dao NamespaceMemberServiceFactory

// NamespaceMemberService is the interface that provides methods to operate on role model
type NamespaceMemberService interface {
	// AddNamespaceMember ...
	AddNamespaceMember(ctx context.Context, userID int64, namespaceObj models.Namespace, role enums.NamespaceRole) (*models.NamespaceMember, error)
	// UpdateNamespaceMember ...
	UpdateNamespaceMember(ctx context.Context, userID int64, namespaceObj models.Namespace, role enums.NamespaceRole) error
	// DeleteNamespaceMember ...
	DeleteNamespaceMember(ctx context.Context, userID int64, namespaceObj models.Namespace) error
	// ListNamespaceMembers ...
	ListNamespaceMembers(ctx context.Context, namespaceID int64, name *string, pagination types.Pagination, sort types.Sortable) ([]*models.NamespaceMember, int64, error)
	// GetNamespaceMember ...
	GetNamespaceMember(ctx context.Context, namespaceID int64, userID int64) (*models.NamespaceMember, error)
	// GetNamespacesMember ...
	GetNamespacesMember(ctx context.Context, namespaceIDs []int64, userID int64) ([]*models.NamespaceMember, error)
	// CountNamespaceMember ...
	CountNamespaceMember(ctx context.Context, userID int64, namespaceID int64) (int64, error)
}

var _ NamespaceMemberService = &namespaceMemberService{}

type namespaceMemberService struct {
	tx *query.Query
}

// NamespaceMemberServiceFactory is the interface that provides the namespace member service factory methods
type NamespaceMemberServiceFactory interface {
	New(txs ...*query.Query) NamespaceMemberService
}

type namespaceMemberServiceFactory struct{}

// NewNamespaceMemberServiceFactory creates a new namespace member service factory
func NewNamespaceMemberServiceFactory() NamespaceMemberServiceFactory {
	return &namespaceMemberServiceFactory{}
}

// New creates a new namespace member service
func (s *namespaceMemberServiceFactory) New(txs ...*query.Query) NamespaceMemberService {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &namespaceMemberService{
		tx: tx,
	}
}

// AddNamespaceMember ...
func (s namespaceMemberService) AddNamespaceMember(ctx context.Context, userID int64, namespaceObj models.Namespace, role enums.NamespaceRole) (*models.NamespaceMember, error) {
	err := s.tx.CasbinRule.WithContext(ctx).Create(&models.CasbinRule{
		PType: ptr.Of("g"),
		V0:    ptr.Of(fmt.Sprintf("%d", userID)),
		V1:    ptr.Of(role.String()),
		V2:    ptr.Of(namespaceObj.Name),
		V3:    ptr.Of(""),
		V4:    ptr.Of(""),
		V5:    ptr.Of(""),
	})
	if err != nil {
		return nil, err
	}
	namespaceMember := &models.NamespaceMember{UserID: userID, NamespaceID: namespaceObj.ID, Role: role}
	err = s.tx.NamespaceMember.WithContext(ctx).Create(namespaceMember)
	if err != nil {
		return nil, err
	}
	return namespaceMember, nil
}

// UpdateNamespaceMember ...
func (s namespaceMemberService) UpdateNamespaceMember(ctx context.Context, userID int64, namespaceObj models.Namespace, role enums.NamespaceRole) error {
	_, err := s.tx.CasbinRule.WithContext(ctx).Where(
		s.tx.CasbinRule.V0.Eq(fmt.Sprintf("%d", userID)),
		s.tx.CasbinRule.V2.Eq(namespaceObj.Name),
	).Updates(map[string]any{
		query.CasbinRule.V1.ColumnName().String(): role,
	})
	if err != nil {
		return err
	}
	_, err = s.tx.NamespaceMember.WithContext(ctx).Where(
		s.tx.NamespaceMember.UserID.Eq(userID),
		s.tx.NamespaceMember.NamespaceID.Eq(namespaceObj.ID),
	).Updates(map[string]any{
		query.NamespaceMember.Role.ColumnName().String(): role,
	})
	return err
}

// DeleteNamespaceMember ...
func (s namespaceMemberService) DeleteNamespaceMember(ctx context.Context, userID int64, namespaceObj models.Namespace) error {
	_, err := s.tx.CasbinRule.WithContext(ctx).Where(
		s.tx.CasbinRule.V0.Eq(fmt.Sprintf("%d", userID)),
		s.tx.CasbinRule.V2.Eq(namespaceObj.Name),
	).Delete()
	if err != nil {
		return err
	}
	_, err = s.tx.NamespaceMember.WithContext(ctx).Where(
		s.tx.NamespaceMember.UserID.Eq(userID),
		s.tx.NamespaceMember.NamespaceID.Eq(namespaceObj.ID),
	).Delete()
	return err
}

// ListNamespaceMembers ...
func (s namespaceMemberService) ListNamespaceMembers(ctx context.Context, namespaceID int64, name *string, pagination types.Pagination, sort types.Sortable) ([]*models.NamespaceMember, int64, error) {
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

// GetNamespaceMember ...
func (s namespaceMemberService) GetNamespaceMember(ctx context.Context, namespaceID int64, userID int64) (*models.NamespaceMember, error) {
	return s.tx.NamespaceMember.WithContext(ctx).Where(
		s.tx.NamespaceMember.UserID.Eq(userID),
		s.tx.NamespaceMember.NamespaceID.Eq(namespaceID),
	).First()
}

// GetNamespacesMember ...
func (s namespaceMemberService) GetNamespacesMember(ctx context.Context, namespaceIDs []int64, userID int64) ([]*models.NamespaceMember, error) {
	if len(namespaceIDs) == 0 {
		return nil, nil
	}
	return s.tx.NamespaceMember.WithContext(ctx).Where(
		s.tx.NamespaceMember.UserID.Eq(userID),
		s.tx.NamespaceMember.NamespaceID.In(namespaceIDs...),
	).Find()
}

// CountNamespaceMember ...
func (s namespaceMemberService) CountNamespaceMember(ctx context.Context, userID int64, namespaceID int64) (int64, error) {
	return s.tx.NamespaceMember.WithContext(ctx).Where(
		s.tx.NamespaceMember.UserID.Eq(userID),
		s.tx.NamespaceMember.NamespaceID.Eq(namespaceID),
	).Count()
}
