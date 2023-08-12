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

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

//go:generate mockgen -destination=mocks/code_repository.go -package=mocks github.com/go-sigma/sigma/pkg/dal/dao CodeRepositoryService
//go:generate mockgen -destination=mocks/code_repository_factory.go -package=mocks github.com/go-sigma/sigma/pkg/dal/dao CodeRepositoryServiceFactory

// CodeRepositoryService is the interface that provides the code repository service methods.
type CodeRepositoryService interface {
	// Create creates new code repository record in the database
	CreateInBatches(ctx context.Context, codeRepositories []*models.CodeRepository) error
	// CreateOwnersInBatches creates new code repository owner records in the database
	CreateOwnersInBatches(ctx context.Context, codeRepositoryOwners []*models.CodeRepositoryOwner) error
	// DeleteInBatches deletes code repository records in the database
	DeleteInBatches(ctx context.Context, ids []int64) error
	// DeleteOwnerInBatches deletes code repository owner records in the database
	DeleteOwnerInBatches(ctx context.Context, ids []int64) error
	// ListAll lists all code repository records in the database
	ListAll(ctx context.Context, user3rdPartyID int64) ([]*models.CodeRepository, error)
	// ListOwnersAll lists all code repository owners records in the database
	ListOwnersAll(ctx context.Context, user3rdPartyID int64) ([]*models.CodeRepositoryOwner, error)
	// ListWithPagination list code repositories with pagination
	ListWithPagination(ctx context.Context, userID int64, provider enums.Provider, owner, name *string, pagination types.Pagination, sort types.Sortable) ([]*models.CodeRepository, int64, error)
	// ListOwnerWithPagination list code repositories with pagination
	ListOwnerWithPagination(ctx context.Context, userID int64, provider enums.Provider, owner *string, pagination types.Pagination, sort types.Sortable) ([]*models.CodeRepositoryOwner, int64, error)
}

type codeRepositoryService struct {
	tx *query.Query
}

// CodeRepositoryServiceFactory is the interface that provides the code repository service factory methods.
type CodeRepositoryServiceFactory interface {
	New(txs ...*query.Query) CodeRepositoryService
}

type codeRepositoryServiceFactory struct{}

// NewCodeRepositoryServiceFactory creates a new code repository service factory
func NewCodeRepositoryServiceFactory() CodeRepositoryServiceFactory {
	return &codeRepositoryServiceFactory{}
}

func (f *codeRepositoryServiceFactory) New(txs ...*query.Query) CodeRepositoryService {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &codeRepositoryService{
		tx: tx,
	}
}

// Create creates new code repository record in the database
func (s *codeRepositoryService) CreateInBatches(ctx context.Context, codeRepositories []*models.CodeRepository) error {
	return s.tx.CodeRepository.WithContext(ctx).CreateInBatches(codeRepositories, consts.InsertBatchSize)
}

// CreateOwnersInBatches creates new code repository owner records in the database
func (s *codeRepositoryService) CreateOwnersInBatches(ctx context.Context, codeRepositoryOwners []*models.CodeRepositoryOwner) error {
	return s.tx.CodeRepositoryOwner.WithContext(ctx).CreateInBatches(codeRepositoryOwners, consts.InsertBatchSize)
}

// DeleteInBatches deletes code repository records in the database
func (s *codeRepositoryService) DeleteInBatches(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := s.tx.CodeRepository.WithContext(ctx).Where(s.tx.CodeRepository.ID.In(ids...)).Delete()
	if err != nil {
		return err
	}
	return nil
}

// DeleteOwnerInBatches deletes code repository owner records in the database
func (s *codeRepositoryService) DeleteOwnerInBatches(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := s.tx.CodeRepositoryOwner.WithContext(ctx).Where(s.tx.CodeRepositoryOwner.ID.In(ids...)).Delete()
	if err != nil {
		return err
	}
	return nil
}

// ListAll lists all code repository records in the database
func (s *codeRepositoryService) ListAll(ctx context.Context, user3rdPartyID int64) ([]*models.CodeRepository, error) {
	return s.tx.CodeRepository.WithContext(ctx).Where(s.tx.CodeRepository.User3rdPartyID.Eq(user3rdPartyID)).Find()
}

// ListOwnersAll lists all code repository owners records in the database
func (s *codeRepositoryService) ListOwnersAll(ctx context.Context, user3rdPartyID int64) ([]*models.CodeRepositoryOwner, error) {
	return s.tx.CodeRepositoryOwner.WithContext(ctx).Where(s.tx.CodeRepositoryOwner.User3rdPartyID.Eq(user3rdPartyID)).Find()
}

// ListWithPagination list code repositories with pagination
func (s *codeRepositoryService) ListWithPagination(ctx context.Context, userID int64, provider enums.Provider, owner, name *string, pagination types.Pagination, sort types.Sortable) ([]*models.CodeRepository, int64, error) {
	user3rdPartyObj, err := s.tx.User3rdParty.WithContext(ctx).Where(s.tx.User3rdParty.UserID.Eq(userID), s.tx.User3rdParty.Provider.Eq(provider)).First()
	if err != nil {
		return nil, 0, err
	}

	pagination = utils.NormalizePagination(pagination)
	query := s.tx.CodeRepository.WithContext(ctx).Where(s.tx.CodeRepository.User3rdPartyID.Eq(user3rdPartyObj.ID))
	if owner != nil {
		query = query.Where(s.tx.CodeRepository.Owner.Eq(ptr.To(owner)))
	}
	if name != nil {
		query = query.Where(s.tx.CodeRepository.Name.Like(fmt.Sprintf("%%%s%%", ptr.To(name))))
	}
	field, ok := s.tx.CodeRepository.GetFieldByName(ptr.To(sort.Sort))
	if ok {
		switch ptr.To(sort.Method) {
		case enums.SortMethodDesc:
			query = query.Order(field.Desc())
		case enums.SortMethodAsc:
			query = query.Order(field)
		default:
			query = query.Order(s.tx.CodeRepository.UpdatedAt.Desc())
		}
	} else {
		query = query.Order(s.tx.CodeRepository.UpdatedAt.Desc())
	}
	return query.FindByPage(ptr.To(pagination.Limit)*(ptr.To(pagination.Page)-1), ptr.To(pagination.Limit))
}

// ListOwnerWithPagination list code repositories with pagination
func (s *codeRepositoryService) ListOwnerWithPagination(ctx context.Context, userID int64, provider enums.Provider, owner *string, pagination types.Pagination, sort types.Sortable) ([]*models.CodeRepositoryOwner, int64, error) {
	user3rdPartyObj, err := s.tx.User3rdParty.WithContext(ctx).Where(s.tx.User3rdParty.UserID.Eq(userID), s.tx.User3rdParty.Provider.Eq(provider)).First()
	if err != nil {
		return nil, 0, err
	}

	pagination = utils.NormalizePagination(pagination)
	query := s.tx.CodeRepositoryOwner.WithContext(ctx).Where(s.tx.CodeRepositoryOwner.User3rdPartyID.Eq(user3rdPartyObj.ID))
	if owner != nil {
		query = query.Where(s.tx.CodeRepositoryOwner.Name.Like(fmt.Sprintf("%%%s%%", ptr.To(owner))))
	}
	field, ok := s.tx.CodeRepositoryOwner.GetFieldByName(ptr.To(sort.Sort))
	if ok {
		switch ptr.To(sort.Method) {
		case enums.SortMethodDesc:
			query = query.Order(field.Desc())
		case enums.SortMethodAsc:
			query = query.Order(field)
		default:
			query = query.Order(s.tx.CodeRepositoryOwner.UpdatedAt.Desc())
		}
	} else {
		query = query.Order(s.tx.CodeRepositoryOwner.UpdatedAt.Desc())
	}
	return query.FindByPage(ptr.To(pagination.Limit)*(ptr.To(pagination.Page)-1), ptr.To(pagination.Limit))
}