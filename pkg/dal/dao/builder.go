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
	"time"

	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

//go:generate mockgen -destination=mocks/builder.go -package=mocks github.com/go-sigma/sigma/pkg/dal/dao BuilderService
//go:generate mockgen -destination=mocks/builder_factory.go -package=mocks github.com/go-sigma/sigma/pkg/dal/dao BuilderServiceFactory

// BuilderService is the interface that provides methods to operate on Builder model
type BuilderService interface {
	// Create creates a new builder record in the database
	Create(ctx context.Context, builder *models.Builder) error
	// Update update the builder by id
	Update(ctx context.Context, id int64, updates map[string]interface{}) error
	// Get get builder by repository id
	Get(ctx context.Context, repositoryID int64) (*models.Builder, error)
	// GetByRepositoryIDs get builders by repository ids
	GetByRepositoryIDs(ctx context.Context, repositoryIDs []int64) (map[int64]*models.Builder, error)
	// Get get builder by repository id
	GetByRepositoryID(ctx context.Context, repositoryID int64) (*models.Builder, error)
	// CreateRunner creates a new builder runner record in the database
	CreateRunner(ctx context.Context, runner *models.BuilderRunner) error
	// GetRunner get runner from object storage or database
	GetRunner(ctx context.Context, id int64) (*models.BuilderRunner, error)
	// ListRunners list builder runners
	ListRunners(ctx context.Context, id int64, pagination types.Pagination, sort types.Sortable) ([]*models.BuilderRunner, int64, error)
	// UpdateRunner update builder runner
	UpdateRunner(ctx context.Context, builderID, runnerID int64, updates map[string]interface{}) error
	// GetByNextTrigger get by next trigger
	GetByNextTrigger(ctx context.Context, now time.Time, limit int) ([]*models.Builder, error)
	// UpdateNextTrigger update next trigger
	UpdateNextTrigger(ctx context.Context, id int64, next time.Time) error
}

type builderService struct {
	tx *query.Query
}

// BuilderServiceFactory is the interface that provides the builder service factory methods.
type BuilderServiceFactory interface {
	New(txs ...*query.Query) BuilderService
}

type builderServiceFactory struct{}

// NewBuilderServiceFactory creates a new builder service factory.
func NewBuilderServiceFactory() BuilderServiceFactory {
	return &builderServiceFactory{}
}

func (f *builderServiceFactory) New(txs ...*query.Query) BuilderService {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &builderService{
		tx: tx,
	}
}

// Create creates a new builder record in the database
func (s builderService) Create(ctx context.Context, builder *models.Builder) error {
	return s.tx.Builder.WithContext(ctx).Create(builder)
}

// Update update the builder by id
func (s builderService) Update(ctx context.Context, id int64, updates map[string]interface{}) error {
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

// Get get builder by id
func (s builderService) Get(ctx context.Context, repositoryID int64) (*models.Builder, error) {
	return s.tx.Builder.WithContext(ctx).Where(s.tx.Builder.RepositoryID.Eq(repositoryID)).First()
}

// GetByRepositoryIDs get builders by repository ids
func (s builderService) GetByRepositoryIDs(ctx context.Context, repositoryIDs []int64) (map[int64]*models.Builder, error) {
	if len(repositoryIDs) == 0 {
		return nil, nil
	}
	builderObjs, err := s.tx.Builder.WithContext(ctx).Where(s.tx.Builder.RepositoryID.In(repositoryIDs...)).Find()
	if err != nil {
		return nil, err
	}
	result := make(map[int64]*models.Builder, len(builderObjs))
	for _, builderObj := range builderObjs {
		result[builderObj.RepositoryID] = builderObj
	}
	return result, nil
}

// Get get builder by repository id
func (s builderService) GetByRepositoryID(ctx context.Context, repositoryID int64) (*models.Builder, error) {
	return s.tx.Builder.WithContext(ctx).Where(s.tx.Builder.RepositoryID.Eq(repositoryID)).First()
}

// CreateRunner creates a new builder runner record in the database
func (s builderService) CreateRunner(ctx context.Context, runner *models.BuilderRunner) error {
	return s.tx.BuilderRunner.WithContext(ctx).Create(runner)
}

// GetRunner get runner from object storage or database
func (s builderService) GetRunner(ctx context.Context, id int64) (*models.BuilderRunner, error) {
	return s.tx.BuilderRunner.WithContext(ctx).Where(s.tx.BuilderRunner.ID.Eq(id)).First()
}

// ListRunners list builder runners
func (s builderService) ListRunners(ctx context.Context, id int64, pagination types.Pagination, sort types.Sortable) ([]*models.BuilderRunner, int64, error) {
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
			query = query.Order(s.tx.BuilderRunner.UpdatedAt.Desc())
		}
	} else {
		query = query.Order(s.tx.BuilderRunner.UpdatedAt.Desc())
	}
	return query.FindByPage(ptr.To(pagination.Limit)*(ptr.To(pagination.Page)-1), ptr.To(pagination.Limit))
}

// UpdateRunner update builder runner
func (s builderService) UpdateRunner(ctx context.Context, builderID, runnerID int64, updates map[string]interface{}) error {
	matched, err := s.tx.BuilderRunner.WithContext(ctx).Where(s.tx.BuilderRunner.BuilderID.Eq(builderID), s.tx.BuilderRunner.ID.Eq(runnerID)).Updates(updates)
	if err != nil {
		return err
	}
	if matched.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// GetByNextTrigger get by next trigger
func (s builderService) GetByNextTrigger(ctx context.Context, now time.Time, limit int) ([]*models.Builder, error) {
	return s.tx.Builder.WithContext(ctx).Where(s.tx.Builder.CronNextTrigger.Lt(now)).Limit(limit).Find()
}

// UpdateNextTrigger update next trigger
func (s builderService) UpdateNextTrigger(ctx context.Context, id int64, next time.Time) error {
	matched, err := s.tx.Builder.WithContext(ctx).Where(s.tx.Builder.ID.Eq(id)).Update(s.tx.Builder.CronNextTrigger, next)
	if err != nil {
		return err
	}
	if matched.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
