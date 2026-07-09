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

package coderepo

import (
	"context"
	"fmt"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

//go:generate mockgen -destination=code_repository_mocks.go -package=coderepo github.com/go-sigma/sigma/pkg/dal/repository/coderepo CodeRepositoryRepository

// CodeRepositoryRepository defines code repository operations
type CodeRepositoryRepository interface {
	// CreateInBatches creates code repository records in batches
	CreateInBatches(ctx context.Context, codeRepositories []*models.CodeRepository) error
	// CreateOwnersInBatches creates code repository owner records in batches
	CreateOwnersInBatches(ctx context.Context, codeRepositoryOwners []*models.CodeRepositoryOwner) error
	// CreateBranchesInBatches creates code repository branch records in batches
	CreateBranchesInBatches(ctx context.Context, branches []*models.CodeRepositoryBranch) error
	// UpdateInBatches updates code repository records in batches
	UpdateInBatches(ctx context.Context, codeRepositories []*models.CodeRepository) error
	// UpdateOwnersInBatches updates code repository owner records in batches
	UpdateOwnersInBatches(ctx context.Context, codeRepositoryOwners []*models.CodeRepositoryOwner) error
	// DeleteInBatches deletes code repository records in batches
	DeleteInBatches(ctx context.Context, ids []string) error
	// DeleteOwnerInBatches deletes code repository owner records in batches
	DeleteOwnerInBatches(ctx context.Context, ids []string) error
	// DeleteBranchesInBatches deletes code repository branch records in batches
	DeleteBranchesInBatches(ctx context.Context, ids []string) error
	// ListAll lists code repositories for a third-party user
	ListAll(ctx context.Context, user3rdPartyID string) ([]*models.CodeRepository, error)
	// Get gets a code repository by ID
	Get(ctx context.Context, id string) (*models.CodeRepository, error)
	// ListOwnersAll lists code repository owners for a third-party user
	ListOwnersAll(ctx context.Context, user3rdPartyID string) ([]*models.CodeRepositoryOwner, error)
	// ListWithPagination lists code repositories with filtering, pagination, and sorting
	ListWithPagination(ctx context.Context, userID string, provider enums.Provider, owner, name *string, pagination api.Pagination, sort api.Sortable) ([]*models.CodeRepository, int64, error)
	// ListOwnerWithoutPagination lists code repository owners without pagination
	ListOwnerWithoutPagination(ctx context.Context, userID string, provider enums.Provider, owner *string) ([]*models.CodeRepositoryOwner, int64, error)
	// ListBranchesWithoutPagination lists branches for a code repository without pagination
	ListBranchesWithoutPagination(ctx context.Context, codeRepositoryID string) ([]*models.CodeRepositoryBranch, int64, error)
	// GetBranchByName gets a branch by code repository ID and branch name
	GetBranchByName(ctx context.Context, codeRepositoryID string, branch string) (*models.CodeRepositoryBranch, error)
	// GetCloneCredential gets clone credentials for a third-party user
	GetCloneCredential(ctx context.Context, user3rdPartyID string) (*models.CodeRepositoryCloneCredential, error)
}

type codeRepositoryRepository struct {
	tx *query.Query
}

// NewCodeRepositoryRepository creates a new code repository repository with the optional query transaction
func NewCodeRepositoryRepository(txs ...*query.Query) CodeRepositoryRepository {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &codeRepositoryRepository{
		tx: tx,
	}
}

// CreateInBatches creates code repository records in batches
func (s *codeRepositoryRepository) CreateInBatches(ctx context.Context, codeRepositories []*models.CodeRepository) error {
	return s.tx.CodeRepository.WithContext(ctx).CreateInBatches(codeRepositories, consts.InsertBatchSize)
}

// CreateOwnersInBatches creates code repository owner records in batches
func (s *codeRepositoryRepository) CreateOwnersInBatches(ctx context.Context, codeRepositoryOwners []*models.CodeRepositoryOwner) error {
	return s.tx.CodeRepositoryOwner.WithContext(ctx).CreateInBatches(codeRepositoryOwners, consts.InsertBatchSize)
}

// CreateBranchesInBatches creates code repository branch records in batches
func (s *codeRepositoryRepository) CreateBranchesInBatches(ctx context.Context, branches []*models.CodeRepositoryBranch) error {
	return s.tx.CodeRepositoryBranch.WithContext(ctx).CreateInBatches(branches, consts.InsertBatchSize)
}

// UpdateInBatches updates code repository records in batches
func (s *codeRepositoryRepository) UpdateInBatches(ctx context.Context, codeRepositories []*models.CodeRepository) error {
	for _, cr := range codeRepositories {
		_, err := s.tx.CodeRepository.WithContext(ctx).Where(
			s.tx.CodeRepository.User3rdPartyID.Eq(cr.User3rdPartyID),
			s.tx.CodeRepository.RepositoryID.Eq(cr.RepositoryID)).Updates(map[string]any{
			query.CodeRepository.Owner.ColumnName().String():    cr.Owner,
			query.CodeRepository.Name.ColumnName().String():     cr.Name,
			query.CodeRepository.SshUrl.ColumnName().String():   cr.SshUrl,
			query.CodeRepository.CloneUrl.ColumnName().String(): cr.CloneUrl,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// UpdateOwnersInBatches updates code repository owner records in batches
func (s *codeRepositoryRepository) UpdateOwnersInBatches(ctx context.Context, codeRepositoryOwners []*models.CodeRepositoryOwner) error {
	for _, cro := range codeRepositoryOwners {
		_, err := s.tx.CodeRepositoryOwner.WithContext(ctx).Where(
			s.tx.CodeRepositoryOwner.User3rdPartyID.Eq(cro.User3rdPartyID),
			s.tx.CodeRepositoryOwner.OwnerID.Eq(cro.OwnerID)).Updates(map[string]any{
			query.CodeRepositoryOwner.Owner.ColumnName().String(): cro.Owner,
			query.CodeRepositoryOwner.IsOrg.ColumnName().String(): cro.IsOrg,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteInBatches deletes code repository records in batches
func (s *codeRepositoryRepository) DeleteInBatches(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := s.tx.CodeRepository.WithContext(ctx).Where(s.tx.CodeRepository.ID.In(ids...)).Delete()
	if err != nil {
		return err
	}
	return nil
}

// DeleteOwnerInBatches deletes code repository owner records in batches
func (s *codeRepositoryRepository) DeleteOwnerInBatches(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := s.tx.CodeRepositoryOwner.WithContext(ctx).Where(s.tx.CodeRepositoryOwner.ID.In(ids...)).Delete()
	if err != nil {
		return err
	}
	return nil
}

// DeleteBranchesInBatches deletes code repository branch records in batches
func (s *codeRepositoryRepository) DeleteBranchesInBatches(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := s.tx.CodeRepositoryBranch.WithContext(ctx).Where(s.tx.CodeRepositoryBranch.ID.In(ids...)).Delete()
	if err != nil {
		return err
	}
	return nil
}

// ListAll lists code repositories for a third-party user
func (s *codeRepositoryRepository) ListAll(ctx context.Context, user3rdPartyID string) ([]*models.CodeRepository, error) {
	return s.tx.CodeRepository.WithContext(ctx).Where(s.tx.CodeRepository.User3rdPartyID.Eq(user3rdPartyID)).Find()
}

// Get gets a code repository by ID
func (s *codeRepositoryRepository) Get(ctx context.Context, id string) (*models.CodeRepository, error) {
	return s.tx.CodeRepository.WithContext(ctx).Where(s.tx.CodeRepository.ID.Eq(id)).Preload(s.tx.CodeRepository.User3rdParty).First()
}

// ListOwnersAll lists code repository owners for a third-party user
func (s *codeRepositoryRepository) ListOwnersAll(ctx context.Context, user3rdPartyID string) ([]*models.CodeRepositoryOwner, error) {
	return s.tx.CodeRepositoryOwner.WithContext(ctx).Where(s.tx.CodeRepositoryOwner.User3rdPartyID.Eq(user3rdPartyID)).Find()
}

// ListWithPagination lists code repositories with filtering, pagination, and sorting
func (s *codeRepositoryRepository) ListWithPagination(ctx context.Context, userID string, provider enums.Provider, owner, name *string, pagination api.Pagination, sort api.Sortable) ([]*models.CodeRepository, int64, error) {
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

// ListOwnerWithoutPagination lists code repository owners without pagination
func (s *codeRepositoryRepository) ListOwnerWithoutPagination(ctx context.Context, userID string, provider enums.Provider, owner *string) ([]*models.CodeRepositoryOwner, int64, error) {
	user3rdPartyObj, err := s.tx.User3rdParty.WithContext(ctx).Where(s.tx.User3rdParty.UserID.Eq(userID), s.tx.User3rdParty.Provider.Eq(provider)).First()
	if err != nil {
		return nil, 0, err
	}

	query := s.tx.CodeRepositoryOwner.WithContext(ctx).Where(s.tx.CodeRepositoryOwner.User3rdPartyID.Eq(user3rdPartyObj.ID))
	if owner != nil {
		query = query.Where(s.tx.CodeRepositoryOwner.Owner.Like(fmt.Sprintf("%%%s%%", ptr.To(owner))))
	}

	return query.FindByPage(-1, -1)
}

// ListBranchesWithoutPagination lists branches for a code repository without pagination
func (s *codeRepositoryRepository) ListBranchesWithoutPagination(ctx context.Context, codeRepositoryID string) ([]*models.CodeRepositoryBranch, int64, error) {
	return s.tx.CodeRepositoryBranch.WithContext(ctx).Where(s.tx.CodeRepositoryBranch.CodeRepositoryID.Eq(codeRepositoryID)).FindByPage(-1, -1)
}

// GetBranchByName gets a branch by code repository ID and branch name
func (s *codeRepositoryRepository) GetBranchByName(ctx context.Context, codeRepositoryID string, branch string) (*models.CodeRepositoryBranch, error) {
	return s.tx.CodeRepositoryBranch.WithContext(ctx).Where(s.tx.CodeRepositoryBranch.CodeRepositoryID.Eq(codeRepositoryID), s.tx.CodeRepositoryBranch.Name.Eq(branch)).First()
}

// GetCloneCredential gets clone credentials for a third-party user
func (s *codeRepositoryRepository) GetCloneCredential(ctx context.Context, user3rdPartyID string) (*models.CodeRepositoryCloneCredential, error) {
	return s.tx.CodeRepositoryCloneCredential.WithContext(ctx).Where(s.tx.CodeRepositoryCloneCredential.User3rdPartyID.Eq(user3rdPartyID)).First()
}
