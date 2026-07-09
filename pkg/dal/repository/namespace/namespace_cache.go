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

package namespace

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
	namespaceCacheByIDPrefix   = "namespace:id"
	namespaceCacheByNamePrefix = "namespace:name"
)

// NamespaceCacheInvalidator invalidates namespace metadata cache entries.
type NamespaceCacheInvalidator interface {
	InvalidateNamespace(ctx context.Context, namespaceObj *models.Namespace) error
	InvalidateNamespaceID(ctx context.Context, id string) error
}

// NamespaceCacheSeeder seeds namespace metadata cache entries.
type NamespaceCacheSeeder interface {
	SeedNamespace(ctx context.Context, namespaceObj *models.Namespace) error
}

type cachedNamespaceRepository struct {
	next   NamespaceRepository
	byID   cacher.Cacher[*models.Namespace]
	byName cacher.Cacher[*models.Namespace]
}

type CachedNamespaceRepositoryParams struct {
	dig.In

	Config             *config.Configuration
	RedisClientFactory dalredis.ClientFactory `optional:"true"`
}

// NewCachedNamespaceRepository creates a cached namespace repository for DI.
func NewCachedNamespaceRepository(params CachedNamespaceRepositoryParams) (NamespaceRepository, error) {
	next := NewNamespaceRepository()
	options := cacher.Options{
		IsNotFound: func(err error) bool {
			return errors.Is(err, gorm.ErrRecordNotFound)
		},
	}
	byID, err := cacher.NewWithOptions(cacher.Params{
		Config:             params.Config,
		RedisClientFactory: params.RedisClientFactory,
	}, namespaceCacheByIDPrefix, next.Get, options)
	if err != nil {
		return nil, err
	}
	byName, err := cacher.NewWithOptions(cacher.Params{
		Config:             params.Config,
		RedisClientFactory: params.RedisClientFactory,
	}, namespaceCacheByNamePrefix, next.GetByName, options)
	if err != nil {
		return nil, err
	}
	return &cachedNamespaceRepository{
		next:   next,
		byID:   byID,
		byName: byName,
	}, nil
}

func (r *cachedNamespaceRepository) Create(ctx context.Context, namespaceObj *models.Namespace) error {
	if err := r.next.Create(ctx, namespaceObj); err != nil {
		return err
	}
	return r.SeedNamespace(ctx, namespaceObj)
}

func (r *cachedNamespaceRepository) FindAll(ctx context.Context) ([]*models.Namespace, error) {
	namespaces, err := r.next.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	r.seedNamespaces(ctx, namespaces)
	return namespaces, nil
}

func (r *cachedNamespaceRepository) FindWithCursor(ctx context.Context, limit int64, last string) ([]*models.Namespace, error) {
	namespaces, err := r.next.FindWithCursor(ctx, limit, last)
	if err != nil {
		return nil, err
	}
	r.seedNamespaces(ctx, namespaces)
	return namespaces, nil
}

func (r *cachedNamespaceRepository) FindRecentlyUpdated(ctx context.Context, limit int) ([]*models.Namespace, error) {
	namespaces, err := r.next.FindRecentlyUpdated(ctx, limit)
	if err != nil {
		return nil, err
	}
	r.seedNamespaces(ctx, namespaces)
	return namespaces, nil
}

func (r *cachedNamespaceRepository) UpdateQuota(ctx context.Context, namespaceID string, limit int64) error {
	namespaceObj := r.getBeforeChange(ctx, namespaceID)
	if err := r.next.UpdateQuota(ctx, namespaceID, limit); err != nil {
		return err
	}
	return r.invalidateKnown(ctx, namespaceObj, namespaceID)
}

func (r *cachedNamespaceRepository) Get(ctx context.Context, id string) (*models.Namespace, error) {
	namespaceObj, err := r.byID.Get(ctx, id)
	if err != nil {
		if errors.Is(err, cacher.ErrNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	if err := r.SeedNamespace(ctx, namespaceObj); err != nil {
		slog.Warn("seed namespace cache failed", "err", err, "namespace_id", namespaceObj.ID)
	}
	return namespaceObj, nil
}

func (r *cachedNamespaceRepository) GetByName(ctx context.Context, name string) (*models.Namespace, error) {
	namespaceObj, err := r.byName.Get(ctx, name)
	if err != nil {
		if errors.Is(err, cacher.ErrNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	if err := r.SeedNamespace(ctx, namespaceObj); err != nil {
		slog.Warn("seed namespace cache failed", "err", err, "namespace_id", namespaceObj.ID)
	}
	return namespaceObj, nil
}

func (r *cachedNamespaceRepository) ListNamespace(ctx context.Context, name *string, pagination api.Pagination, sort api.Sortable) ([]*models.Namespace, int64, error) {
	namespaces, total, err := r.next.ListNamespace(ctx, name, pagination, sort)
	if err != nil {
		return nil, 0, err
	}
	r.seedNamespaces(ctx, namespaces)
	return namespaces, total, nil
}

func (r *cachedNamespaceRepository) ListNamespaceWithAuth(ctx context.Context, userID string, name *string, pagination api.Pagination, sort api.Sortable) ([]*models.Namespace, int64, error) {
	namespaces, total, err := r.next.ListNamespaceWithAuth(ctx, userID, name, pagination, sort)
	if err != nil {
		return nil, 0, err
	}
	r.seedNamespaces(ctx, namespaces)
	return namespaces, total, nil
}

func (r *cachedNamespaceRepository) CountNamespace(ctx context.Context, name *string) (int64, error) {
	return r.next.CountNamespace(ctx, name)
}

func (r *cachedNamespaceRepository) DeleteByID(ctx context.Context, id string) error {
	namespaceObj := r.getBeforeChange(ctx, id)
	if err := r.next.DeleteByID(ctx, id); err != nil {
		return err
	}
	return r.invalidateKnown(ctx, namespaceObj, id)
}

func (r *cachedNamespaceRepository) UpdateByID(ctx context.Context, id string, updates map[string]any) error {
	namespaceObj := r.getBeforeChange(ctx, id)
	if err := r.next.UpdateByID(ctx, id, updates); err != nil {
		return err
	}
	return r.invalidateKnown(ctx, namespaceObj, id)
}

func (r *cachedNamespaceRepository) IncrementSize(ctx context.Context, namespaceID string, delta int64) error {
	namespaceObj := r.getBeforeChange(ctx, namespaceID)
	if err := r.next.IncrementSize(ctx, namespaceID, delta); err != nil {
		return err
	}
	return r.invalidateKnown(ctx, namespaceObj, namespaceID)
}

func (r *cachedNamespaceRepository) DecrementSize(ctx context.Context, namespaceID string, delta int64) error {
	namespaceObj := r.getBeforeChange(ctx, namespaceID)
	if err := r.next.DecrementSize(ctx, namespaceID, delta); err != nil {
		return err
	}
	return r.invalidateKnown(ctx, namespaceObj, namespaceID)
}

func (r *cachedNamespaceRepository) FindWithCursorDirty(ctx context.Context, limit int64, last string) ([]*models.Namespace, error) {
	namespaces, err := r.next.FindWithCursorDirty(ctx, limit, last)
	if err != nil {
		return nil, err
	}
	r.seedNamespaces(ctx, namespaces)
	return namespaces, nil
}

func (r *cachedNamespaceRepository) UpdateSizeAndClearDirty(ctx context.Context, namespaceID string, size int64) error {
	namespaceObj := r.getBeforeChange(ctx, namespaceID)
	if err := r.next.UpdateSizeAndClearDirty(ctx, namespaceID, size); err != nil {
		return err
	}
	return r.invalidateKnown(ctx, namespaceObj, namespaceID)
}

func (r *cachedNamespaceRepository) ClearSizeDirty(ctx context.Context, namespaceID string) error {
	namespaceObj := r.getBeforeChange(ctx, namespaceID)
	if err := r.next.ClearSizeDirty(ctx, namespaceID); err != nil {
		return err
	}
	return r.invalidateKnown(ctx, namespaceObj, namespaceID)
}

func (r *cachedNamespaceRepository) SeedNamespace(ctx context.Context, namespaceObj *models.Namespace) error {
	if namespaceObj == nil {
		return nil
	}
	if err := r.byID.Set(ctx, namespaceObj.ID, namespaceObj); err != nil {
		return err
	}
	return r.byName.Set(ctx, namespaceObj.Name, namespaceObj)
}

func (r *cachedNamespaceRepository) InvalidateNamespace(ctx context.Context, namespaceObj *models.Namespace) error {
	if namespaceObj == nil {
		return nil
	}
	if err := r.byID.Del(ctx, namespaceObj.ID); err != nil {
		return err
	}
	return r.byName.Del(ctx, namespaceObj.Name)
}

func (r *cachedNamespaceRepository) InvalidateNamespaceID(ctx context.Context, id string) error {
	return r.byID.Del(ctx, id)
}

func (r *cachedNamespaceRepository) seedNamespaces(ctx context.Context, namespaces []*models.Namespace) {
	for _, namespaceObj := range namespaces {
		if err := r.SeedNamespace(ctx, namespaceObj); err != nil {
			slog.Warn("seed namespace cache failed", "err", err, "namespace_id", namespaceObj.ID)
		}
	}
}

func (r *cachedNamespaceRepository) getBeforeChange(ctx context.Context, id string) *models.Namespace {
	namespaceObj, err := r.next.Get(ctx, id)
	if err != nil {
		return nil
	}
	return namespaceObj
}

func (r *cachedNamespaceRepository) invalidateKnown(ctx context.Context, namespaceObj *models.Namespace, id string) error {
	if namespaceObj != nil {
		return r.InvalidateNamespace(ctx, namespaceObj)
	}
	return r.InvalidateNamespaceID(ctx, id)
}
