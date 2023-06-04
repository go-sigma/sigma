// Copyright 2023 XImager
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

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/ximager/ximager/pkg/dal/models"
	"github.com/ximager/ximager/pkg/dal/query"
	"github.com/ximager/ximager/pkg/types"
	"github.com/ximager/ximager/pkg/utils/imagerefs"
)

//go:generate mockgen -destination=mocks/repository.go -package=mocks github.com/ximager/ximager/pkg/dal/dao RepositoryService
//go:generate mockgen -destination=mocks/repository_factory.go -package=mocks github.com/ximager/ximager/pkg/dal/dao RepositoryServiceFactory

// RepositoryService is the interface that provides the repository service methods.
type RepositoryService interface {
	// Create creates a new repository.
	Create(context.Context, *models.Repository) error
	// Save saves the repository.
	Save(context.Context, *models.Repository) error
	// Get gets the repository with the specified repository ID.
	Get(context.Context, uint64) (*models.Repository, error)
	// GetByName gets the repository with the specified repository name.
	GetByName(context.Context, string) (*models.Repository, error)
	// ListByDtPagination lists the repositories by the pagination.
	ListByDtPagination(ctx context.Context, limit int, lastID ...uint64) ([]*models.Repository, error)
	// ListRepository lists all repositories.
	ListRepository(ctx context.Context, req types.ListRepositoryRequest) ([]*models.Repository, error)
	// CountRepository counts all repositories.
	CountRepository(ctx context.Context, req types.ListRepositoryRequest) (int64, error)
	// DeleteByID deletes the repository with the specified repository ID.
	DeleteByID(ctx context.Context, id uint64) error
}

type repositoryService struct {
	tx *query.Query
}

// RepositoryServiceFactory is the interface that provides the repository service factory methods.
type RepositoryServiceFactory interface {
	New(txs ...*query.Query) RepositoryService
}

type repositoryServiceFactory struct{}

// NewRepositoryServiceFactory creates a new repository service factory.
func NewRepositoryServiceFactory() RepositoryServiceFactory {
	return &repositoryServiceFactory{}
}

func (f *repositoryServiceFactory) New(txs ...*query.Query) RepositoryService {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &repositoryService{
		tx: tx,
	}
}

// Create creates a new repository.
func (s *repositoryService) Create(ctx context.Context, repository *models.Repository) error {
	return s.tx.Repository.WithContext(ctx).Create(repository)
}

// Save saves the repository.
func (s *repositoryService) Save(ctx context.Context, repository *models.Repository) error {
	_, ns, _, _, err := imagerefs.Parse(repository.Name)
	if err != nil {
		return err
	}
	nsObj, err := s.tx.Namespace.WithContext(ctx).Where(s.tx.Namespace.Name.Eq(ns)).First()
	if err != nil {
		return err
	}
	repository.NamespaceID = nsObj.ID
	err = s.tx.Repository.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(repository)
	if err != nil {
		return err
	}
	rRepository, err := s.tx.Repository.WithContext(ctx).Where(s.tx.Repository.NamespaceID.Eq(nsObj.ID),
		s.tx.Repository.Name.Eq(repository.Name)).First()
	if err != nil {
		return err
	}
	return copier.Copy(repository, rRepository)
}

// Get gets the repository with the specified repository ID.
func (s *repositoryService) Get(ctx context.Context, id uint64) (*models.Repository, error) {
	repo, err := s.tx.Repository.WithContext(ctx).Where(s.tx.Repository.ID.Eq(id)).First()
	if err != nil {
		return nil, err
	}
	return repo, nil
}

// GetByName gets the repository with the specified repository name.
func (s *repositoryService) GetByName(ctx context.Context, name string) (*models.Repository, error) {
	repo, err := s.tx.Repository.WithContext(ctx).Where(s.tx.Repository.Name.Eq(name)).First()
	if err != nil {
		return nil, err
	}
	return repo, nil
}

// ListByDtPagination lists the repositories by the pagination.
func (s *repositoryService) ListByDtPagination(ctx context.Context, limit int, lastID ...uint64) ([]*models.Repository, error) {
	do := s.tx.Repository.WithContext(ctx)
	if len(lastID) > 0 {
		do = do.Where(s.tx.Tag.ID.Gt(lastID[0]))
	}
	repositories, err := do.Order(s.tx.Repository.ID).Limit(limit).Find()
	if err != nil {
		return nil, err
	}
	return repositories, nil
}

// ListRepository lists all repositories.
func (s *repositoryService) ListRepository(ctx context.Context, req types.ListRepositoryRequest) ([]*models.Repository, error) {
	query := s.tx.Repository.WithContext(ctx).
		LeftJoin(s.tx.Namespace, s.tx.Namespace.ID.EqCol(s.tx.Repository.NamespaceID)).
		Where(s.tx.Namespace.Name.Eq(req.Namespace)).
		Offset(req.PageSize * (req.PageNum - 1)).Limit(req.PageSize)
	return query.Find()
}

// CountRepository counts all repositories.
func (s *repositoryService) CountRepository(ctx context.Context, req types.ListRepositoryRequest) (int64, error) {
	return s.tx.Repository.WithContext(ctx).Count()
}

// DeleteByID deletes the repository with the specified repository ID.
func (s *repositoryService) DeleteByID(ctx context.Context, id uint64) error {
	matched, err := s.tx.Repository.WithContext(ctx).Where(s.tx.Repository.ID.Eq(id)).Delete()
	if err != nil {
		return err
	}
	if matched.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
