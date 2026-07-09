// Copyright 2026 sigma
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

package registry

import (
	"context"
	"errors"
	"log/slog"

	"go.uber.org/dig"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/dal/models"
	dalredis "github.com/go-sigma/sigma/pkg/dal/redis"
	cacher "github.com/go-sigma/sigma/pkg/infra/cache"
)

const (
	repositoryCacheByIDPrefix   = "repository:id"
	repositoryCacheByNamePrefix = "repository:name"
)

// RepositoryCacheInvalidator invalidates repository metadata cache entries.
type RepositoryCacheInvalidator interface {
	InvalidateRepository(ctx context.Context, repositoryObj *models.Repository) error
	InvalidateRepositoryID(ctx context.Context, id string) error
}

// RepositoryCacheSeeder seeds repository metadata cache entries.
type RepositoryCacheSeeder interface {
	SeedRepository(ctx context.Context, repositoryObj *models.Repository) error
}

type cachedRepositoryRepository struct {
	next   RepositoryRepository
	byID   cacher.Cacher[*models.Repository]
	byName cacher.Cacher[*models.Repository]
}

type CachedRepositoryRepositoryParams struct {
	dig.In

	Config             *config.Configuration
	RedisClientFactory dalredis.ClientFactory `optional:"true"`
}

// NewCachedRepositoryRepository creates a cached registry repository for DI.
func NewCachedRepositoryRepository(params CachedRepositoryRepositoryParams) (RepositoryRepository, error) {
	next := NewRepositoryRepository()
	options := cacher.Options{
		IsNotFound: func(err error) bool {
			return errors.Is(err, gorm.ErrRecordNotFound)
		},
	}
	byID, err := cacher.NewWithOptions(cacher.Params{
		Config:             params.Config,
		RedisClientFactory: params.RedisClientFactory,
	}, repositoryCacheByIDPrefix, next.Get, options)
	if err != nil {
		return nil, err
	}
	byName, err := cacher.NewWithOptions(cacher.Params{
		Config:             params.Config,
		RedisClientFactory: params.RedisClientFactory,
	}, repositoryCacheByNamePrefix, next.GetByName, options)
	if err != nil {
		return nil, err
	}
	return &cachedRepositoryRepository{
		next:   next,
		byID:   byID,
		byName: byName,
	}, nil
}

func (r *cachedRepositoryRepository) Create(ctx context.Context, repositoryObj *models.Repository) error {
	if err := r.next.Create(ctx, repositoryObj); err != nil {
		return err
	}
	return r.SeedRepository(ctx, repositoryObj)
}

func (r *cachedRepositoryRepository) FindAll(ctx context.Context, namespaceID string, limit int64, last string) ([]*models.Repository, error) {
	repositories, err := r.next.FindAll(ctx, namespaceID, limit, last)
	if err != nil {
		return nil, err
	}
	r.seedRepositories(ctx, repositories)
	return repositories, nil
}

func (r *cachedRepositoryRepository) FindDeletableWithCursor(ctx context.Context, namespaceID string, before int64, last string, limit int64) ([]*models.Repository, error) {
	repositories, err := r.next.FindDeletableWithCursor(ctx, namespaceID, before, last, limit)
	if err != nil {
		return nil, err
	}
	r.seedRepositories(ctx, repositories)
	return repositories, nil
}

func (r *cachedRepositoryRepository) FindRecentlyUpdated(ctx context.Context, limit int) ([]*models.Repository, error) {
	repositories, err := r.next.FindRecentlyUpdated(ctx, limit)
	if err != nil {
		return nil, err
	}
	r.seedRepositories(ctx, repositories)
	return repositories, nil
}

func (r *cachedRepositoryRepository) Get(ctx context.Context, repositoryID string) (*models.Repository, error) {
	repositoryObj, err := r.byID.Get(ctx, repositoryID)
	if err != nil {
		if errors.Is(err, cacher.ErrNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	if err := r.SeedRepository(ctx, repositoryObj); err != nil {
		slog.Warn("seed repository cache failed", "err", err, "repository_id", repositoryObj.ID)
	}
	return repositoryObj, nil
}

func (r *cachedRepositoryRepository) GetByName(ctx context.Context, name string) (*models.Repository, error) {
	repositoryObj, err := r.byName.Get(ctx, name)
	if err != nil {
		if errors.Is(err, cacher.ErrNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	if err := r.SeedRepository(ctx, repositoryObj); err != nil {
		slog.Warn("seed repository cache failed", "err", err, "repository_id", repositoryObj.ID)
	}
	return repositoryObj, nil
}

func (r *cachedRepositoryRepository) ListByDtPagination(ctx context.Context, limit int, lastID ...string) ([]*models.Repository, error) {
	repositories, err := r.next.ListByDtPagination(ctx, limit, lastID...)
	if err != nil {
		return nil, err
	}
	r.seedRepositories(ctx, repositories)
	return repositories, nil
}

func (r *cachedRepositoryRepository) ListWithScrollable(ctx context.Context, namespaceID, userID string, name *string, limit int, lastID string) ([]*models.Repository, error) {
	repositories, err := r.next.ListWithScrollable(ctx, namespaceID, userID, name, limit, lastID)
	if err != nil {
		return nil, err
	}
	r.seedRepositories(ctx, repositories)
	return repositories, nil
}

func (r *cachedRepositoryRepository) ListRepository(ctx context.Context, namespaceID string, name *string, pagination api.Pagination, sort api.Sortable) ([]*models.Repository, int64, error) {
	repositories, total, err := r.next.ListRepository(ctx, namespaceID, name, pagination, sort)
	if err != nil {
		return nil, 0, err
	}
	r.seedRepositories(ctx, repositories)
	return repositories, total, nil
}

func (r *cachedRepositoryRepository) ListRepositoryWithAuth(ctx context.Context, namespaceID, userID string, name *string, pagination api.Pagination, sort api.Sortable) ([]*models.Repository, int64, error) {
	repositories, total, err := r.next.ListRepositoryWithAuth(ctx, namespaceID, userID, name, pagination, sort)
	if err != nil {
		return nil, 0, err
	}
	r.seedRepositories(ctx, repositories)
	return repositories, total, nil
}

func (r *cachedRepositoryRepository) CountRepository(ctx context.Context, namespaceID string, name *string) (int64, error) {
	return r.next.CountRepository(ctx, namespaceID, name)
}

func (r *cachedRepositoryRepository) UpdateRepository(ctx context.Context, id string, updates map[string]any) error {
	repositoryObj := r.getBeforeChange(ctx, id)
	if err := r.next.UpdateRepository(ctx, id, updates); err != nil {
		return err
	}
	return r.invalidateKnown(ctx, repositoryObj, id)
}

func (r *cachedRepositoryRepository) CountByNamespace(ctx context.Context, namespaceIDs []string) (map[string]int64, error) {
	return r.next.CountByNamespace(ctx, namespaceIDs)
}

func (r *cachedRepositoryRepository) DeleteByID(ctx context.Context, id string) error {
	repositoryObj := r.getBeforeChange(ctx, id)
	if err := r.next.DeleteByID(ctx, id); err != nil {
		return err
	}
	return r.invalidateKnown(ctx, repositoryObj, id)
}

func (r *cachedRepositoryRepository) DeleteEmpty(ctx context.Context, namespaceID *string) ([]string, error) {
	names, err := r.next.DeleteEmpty(ctx, namespaceID)
	if err != nil {
		return nil, err
	}
	for _, name := range names {
		if err := r.byName.Del(ctx, name); err != nil {
			slog.Warn("invalidate repository cache by name failed", "err", err, "repository", name)
		}
	}
	return names, nil
}

func (r *cachedRepositoryRepository) IncrementSize(ctx context.Context, repositoryID string, delta int64) error {
	repositoryObj := r.getBeforeChange(ctx, repositoryID)
	if err := r.next.IncrementSize(ctx, repositoryID, delta); err != nil {
		return err
	}
	return r.invalidateKnown(ctx, repositoryObj, repositoryID)
}

func (r *cachedRepositoryRepository) DecrementSize(ctx context.Context, repositoryID string, delta int64) error {
	repositoryObj := r.getBeforeChange(ctx, repositoryID)
	if err := r.next.DecrementSize(ctx, repositoryID, delta); err != nil {
		return err
	}
	return r.invalidateKnown(ctx, repositoryObj, repositoryID)
}

func (r *cachedRepositoryRepository) FindWithCursorDirty(ctx context.Context, limit int64, last string) ([]*models.Repository, error) {
	repositories, err := r.next.FindWithCursorDirty(ctx, limit, last)
	if err != nil {
		return nil, err
	}
	r.seedRepositories(ctx, repositories)
	return repositories, nil
}

func (r *cachedRepositoryRepository) UpdateSizeAndClearDirty(ctx context.Context, repositoryID string, size int64) error {
	repositoryObj := r.getBeforeChange(ctx, repositoryID)
	if err := r.next.UpdateSizeAndClearDirty(ctx, repositoryID, size); err != nil {
		return err
	}
	return r.invalidateKnown(ctx, repositoryObj, repositoryID)
}

func (r *cachedRepositoryRepository) ClearSizeDirty(ctx context.Context, repositoryID string) error {
	repositoryObj := r.getBeforeChange(ctx, repositoryID)
	if err := r.next.ClearSizeDirty(ctx, repositoryID); err != nil {
		return err
	}
	return r.invalidateKnown(ctx, repositoryObj, repositoryID)
}

func (r *cachedRepositoryRepository) SeedRepository(ctx context.Context, repositoryObj *models.Repository) error {
	if repositoryObj == nil {
		return nil
	}
	if err := r.byID.Set(ctx, repositoryObj.ID, repositoryObj); err != nil {
		return err
	}
	return r.byName.Set(ctx, repositoryObj.Name, repositoryObj)
}

func (r *cachedRepositoryRepository) InvalidateRepository(ctx context.Context, repositoryObj *models.Repository) error {
	if repositoryObj == nil {
		return nil
	}
	if err := r.byID.Del(ctx, repositoryObj.ID); err != nil {
		return err
	}
	return r.byName.Del(ctx, repositoryObj.Name)
}

func (r *cachedRepositoryRepository) InvalidateRepositoryID(ctx context.Context, id string) error {
	return r.byID.Del(ctx, id)
}

func (r *cachedRepositoryRepository) seedRepositories(ctx context.Context, repositories []*models.Repository) {
	for _, repositoryObj := range repositories {
		if err := r.SeedRepository(ctx, repositoryObj); err != nil {
			slog.Warn("seed repository cache failed", "err", err, "repository_id", repositoryObj.ID)
		}
	}
}

func (r *cachedRepositoryRepository) getBeforeChange(ctx context.Context, id string) *models.Repository {
	repositoryObj, err := r.next.Get(ctx, id)
	if err != nil {
		return nil
	}
	return repositoryObj
}

func (r *cachedRepositoryRepository) invalidateKnown(ctx context.Context, repositoryObj *models.Repository, id string) error {
	if repositoryObj != nil {
		return r.InvalidateRepository(ctx, repositoryObj)
	}
	return r.InvalidateRepositoryID(ctx, id)
}
