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

package builder

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

//go:generate mockgen -destination=builder_mocks.go -package=builder github.com/go-sigma/sigma/pkg/dal/repository/builder BuilderRepository

// BuilderRepository defines builder repository operations
type BuilderRepository interface {
	// Create creates a new builder
	Create(ctx context.Context, builder *models.Builder) error
	// Update updates the builder with the specified ID
	Update(ctx context.Context, id string, updates map[string]any) error
	// Get gets the builder with the specified repository ID
	Get(ctx context.Context, repositoryID string) (*models.Builder, error)
	// GetByRepositoryIDs gets builders grouped by repository ID
	GetByRepositoryIDs(ctx context.Context, repositoryIDs []string) (map[string]*models.Builder, error)
	// GetByRepositoryID gets the builder with the specified repository ID
	GetByRepositoryID(ctx context.Context, repositoryID string) (*models.Builder, error)
	// CreateRunner creates a new builder runner
	CreateRunner(ctx context.Context, runner *models.BuilderRunner) error
	// GetRunner gets a builder runner by ID
	GetRunner(ctx context.Context, id string) (*models.BuilderRunner, error)
	// ListRunners lists builder runners with pagination and sorting
	ListRunners(ctx context.Context, id string, pagination api.Pagination, sort api.Sortable) ([]*models.BuilderRunner, int64, error)
	// UpdateRunner updates a builder runner
	UpdateRunner(ctx context.Context, builderID, runnerID string, updates map[string]any) error
	// GetByNextTrigger gets builders whose next trigger is before the given time
	GetByNextTrigger(ctx context.Context, now time.Time, limit int) ([]*models.Builder, error)
	// UpdateNextTrigger updates a builder's next trigger time
	UpdateNextTrigger(ctx context.Context, id string, next time.Time) error
	// ClaimNextTrigger advances a due builder's next trigger time
	ClaimNextTrigger(ctx context.Context, id string, dueBefore, next time.Time) (bool, error)
}

type builderRepository struct {
	tx *query.Query
}

// NewBuilderRepository creates a new builder repository with the optional query transaction
func NewBuilderRepository(txs ...*query.Query) BuilderRepository {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &builderRepository{
		tx: tx,
	}
}

// Create creates a new builder
func (s builderRepository) Create(ctx context.Context, builder *models.Builder) error {
	return s.tx.Builder.WithContext(ctx).Create(builder)
}

// Update updates the builder with the specified ID
func (s builderRepository) Update(ctx context.Context, id string, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	result, err := s.tx.Builder.WithContext(ctx).Where(s.tx.Builder.ID.Eq(id)).UpdateColumns(updates)
	if err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Get gets the builder with the specified repository ID
func (s builderRepository) Get(ctx context.Context, repositoryID string) (*models.Builder, error) {
	return s.tx.Builder.WithContext(ctx).Where(s.tx.Builder.RepositoryID.Eq(repositoryID)).First()
}

// GetByRepositoryIDs gets builders grouped by repository ID
func (s builderRepository) GetByRepositoryIDs(ctx context.Context, repositoryIDs []string) (map[string]*models.Builder, error) {
	if len(repositoryIDs) == 0 {
		return nil, nil
	}
	builderObjs, err := s.tx.Builder.WithContext(ctx).Where(s.tx.Builder.RepositoryID.In(repositoryIDs...)).Find()
	if err != nil {
		return nil, err
	}
	result := make(map[string]*models.Builder, len(builderObjs))
	for _, builderObj := range builderObjs {
		result[builderObj.RepositoryID] = builderObj
	}
	return result, nil
}

// GetByRepositoryID gets the builder with the specified repository ID
func (s builderRepository) GetByRepositoryID(ctx context.Context, repositoryID string) (*models.Builder, error) {
	query := s.tx.Builder.WithContext(ctx).Where(s.tx.Builder.RepositoryID.Eq(repositoryID)).Preload(s.tx.Builder.CodeRepository)
	query.UnderlyingDB().Preload("CodeRepository.User3rdParty")
	return query.First()
}

// CreateRunner creates a new builder runner
func (s builderRepository) CreateRunner(ctx context.Context, runner *models.BuilderRunner) error {
	return s.tx.BuilderRunner.WithContext(ctx).Create(runner)
}

// GetRunner gets a builder runner by ID
func (s builderRepository) GetRunner(ctx context.Context, id string) (*models.BuilderRunner, error) {
	return s.tx.BuilderRunner.WithContext(ctx).Where(s.tx.BuilderRunner.ID.Eq(id)).Preload(s.tx.BuilderRunner.Builder).First()
}

// ListRunners lists builder runners with pagination and sorting
func (s builderRepository) ListRunners(ctx context.Context, id string, pagination api.Pagination, sort api.Sortable) ([]*models.BuilderRunner, int64, error) {
	pagination = utils.NormalizePagination(pagination)
	query := s.tx.BuilderRunner.WithContext(ctx).Where(s.tx.BuilderRunner.BuilderID.Eq(id))
	field, ok := s.tx.BuilderRunner.GetFieldByName(ptr.To(sort.Sort))
	if ok {
		switch ptr.To(sort.Method) {
		case enums.SortMethodDesc:
			query = query.Order(field.Desc())
		case enums.SortMethodAsc:
			query = query.Order(field)
		default:
			query = query.Order(s.tx.BuilderRunner.CreatedAt.Desc())
		}
	} else {
		query = query.Order(s.tx.BuilderRunner.CreatedAt.Desc())
	}
	return query.FindByPage(ptr.To(pagination.Limit)*(ptr.To(pagination.Page)-1), ptr.To(pagination.Limit))
}

// UpdateRunner updates a builder runner
func (s builderRepository) UpdateRunner(ctx context.Context, builderID, runnerID string, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	updates[query.BuilderRunner.ID.ColumnName().String()] = runnerID
	matched, err := s.tx.BuilderRunner.WithContext(ctx).Where(
		s.tx.BuilderRunner.BuilderID.Eq(builderID),
		s.tx.BuilderRunner.ID.Eq(runnerID)).Updates(updates)
	if err != nil {
		return err
	}
	if matched.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// GetByNextTrigger gets builders whose next trigger is before the given time
func (s builderRepository) GetByNextTrigger(ctx context.Context, now time.Time, limit int) ([]*models.Builder, error) {
	return s.tx.Builder.WithContext(ctx).Where(s.tx.Builder.CronNextTrigger.Lt(now.UnixMilli())).Limit(limit).Find()
}

// UpdateNextTrigger updates a builder's next trigger time
func (s builderRepository) UpdateNextTrigger(ctx context.Context, id string, next time.Time) error {
	matched, err := s.tx.Builder.WithContext(ctx).Where(s.tx.Builder.ID.Eq(id)).
		Update(s.tx.Builder.CronNextTrigger, next.UnixMilli())
	if err != nil {
		return err
	}
	if matched.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ClaimNextTrigger advances a due builder's next trigger time.
func (s builderRepository) ClaimNextTrigger(ctx context.Context, id string, dueBefore, next time.Time) (bool, error) {
	matched, err := s.tx.Builder.WithContext(ctx).
		Where(s.tx.Builder.ID.Eq(id)).
		Where(s.tx.Builder.CronNextTrigger.Lt(dueBefore.UnixMilli())).
		Update(s.tx.Builder.CronNextTrigger, next.UnixMilli())
	if err != nil {
		return false, err
	}
	return matched.RowsAffected == 1, nil
}
