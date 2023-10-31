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

	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

//go:generate mockgen -destination=mocks/daemon.go -package=mocks github.com/go-sigma/sigma/pkg/dal/dao DaemonService
//go:generate mockgen -destination=mocks/daemon_factory.go -package=mocks github.com/go-sigma/sigma/pkg/dal/dao DaemonServiceFactory

// DaemonService is the interface that provides methods to operate on daemon model
type DaemonService interface {
	// Create creates a new daemon log record in the database
	Create(ctx context.Context, daemonLog *models.DaemonLog) error
	// CreateMany creates many new daemon log records in the database
	CreateMany(ctx context.Context, daemonLogs []*models.DaemonLog) error
	// Delete delete a daemon log record with specific id
	Delete(ctx context.Context, id int64) error
	// List lists all daemon log
	List(ctx context.Context, pagination types.Pagination, sort types.Sortable) ([]*models.DaemonLog, int64, error)

	// GetGcRepositoryRunner ...
	GetGcRepositoryRunner(ctx context.Context, runnerID int64) (*models.DaemonGcRepositoryRunner, error)
	// GetLastGcRepositoryRunner ...
	GetLastGcRepositoryRunner(ctx context.Context, namespaceID *int64) (*models.DaemonGcRepositoryRunner, error)
	// CreateGcRepositoryRunner ...
	CreateGcRepositoryRunner(ctx context.Context, runnerObj *models.DaemonGcRepositoryRunner) error
	// ListGcRepositoryRunners ...
	ListGcRepositoryRunners(ctx context.Context, namespaceID *int64, pagination types.Pagination, sort types.Sortable) ([]*models.DaemonGcRepositoryRunner, int64, error)
	// CreateGcRepositoryRecords ...
	CreateGcRepositoryRecords(ctx context.Context, records []*models.DaemonGcRepositoryRecord) error
	// UpdateGcRepositoryRunner ...
	UpdateGcRepositoryRunner(ctx context.Context, runnerID int64, updates map[string]interface{}) error
	// ListGcRepositoryRecords lists all gc repository records.
	ListGcRepositoryRecords(ctx context.Context, runnerID int64, pagination types.Pagination, sort types.Sortable) ([]*models.DaemonGcRepositoryRecord, int64, error)

	// GetGcBlobRunner ...
	GetGcBlobRunner(ctx context.Context, runnerID int64) (*models.DaemonGcBlobRunner, error)
	// GetLastGcBlobRunner ...
	GetLastGcBlobRunner(ctx context.Context) (*models.DaemonGcBlobRunner, error)
	// ListLastGcBlobRunner ...
	ListLastGcBlobRunner(ctx context.Context, pagination types.Pagination, sort types.Sortable) ([]*models.DaemonGcBlobRunner, int64, error)
	// CreateGcBlobRunner ...
	CreateGcBlobRunner(ctx context.Context, runnerObj *models.DaemonGcBlobRunner) error
	// CreateGcBlobRecords ...
	CreateGcBlobRecords(ctx context.Context, records []*models.DaemonGcBlobRecord) error
	// UpdateGcBlobRunner ...
	UpdateGcBlobRunner(ctx context.Context, runnerID int64, updates map[string]interface{}) error
	// ListGcBlobRecords ...
	ListGcBlobRecords(ctx context.Context, runnerID int64, pagination types.Pagination, sort types.Sortable) ([]*models.DaemonGcBlobRecord, int64, error)

	// GetGcArtifactRunner ...
	GetGcArtifactRunner(ctx context.Context, runnerID int64) (*models.DaemonGcArtifactRunner, error)
	// GetLastGcArtifactRunner ...
	GetLastGcArtifactRunner(ctx context.Context, namespaceID *int64) (*models.DaemonGcArtifactRunner, error)
	// ListLastGcArtifactRunner ...
	ListLastGcArtifactRunner(ctx context.Context, namespaceID *int64, pagination types.Pagination, sort types.Sortable) ([]*models.DaemonGcArtifactRunner, int64, error)
	// CreateGcArtifactRunner ...
	CreateGcArtifactRunner(ctx context.Context, runnerObj *models.DaemonGcArtifactRunner) error
	// CreateGcArtifactRecords ...
	CreateGcArtifactRecords(ctx context.Context, records []*models.DaemonGcArtifactRecord) error
	// UpdateGcArtifactRunner ...
	UpdateGcArtifactRunner(ctx context.Context, runnerID int64, updates map[string]interface{}) error
	// ListGcArtifactRecords ...
	ListGcArtifactRecords(ctx context.Context, runnerID int64, pagination types.Pagination, sort types.Sortable) ([]*models.DaemonGcArtifactRecord, int64, error)
}

type daemonService struct {
	tx *query.Query
}

// DaemonServiceFactory is the interface that provides the daemon service factory methods.
type DaemonServiceFactory interface {
	New(txs ...*query.Query) DaemonService
}

type daemonServiceFactory struct{}

// NewDaemonServiceFactory creates a new audit service factory.
func NewDaemonServiceFactory() DaemonServiceFactory {
	return &daemonServiceFactory{}
}

func (f *daemonServiceFactory) New(txs ...*query.Query) DaemonService {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &daemonService{
		tx: tx,
	}
}

// Create creates a new daemon record in the database
func (s *daemonService) Create(ctx context.Context, daemonLog *models.DaemonLog) error {
	return s.tx.DaemonLog.WithContext(ctx).Create(daemonLog)
}

// CreateMany creates many new daemon log records in the database
func (s *daemonService) CreateMany(ctx context.Context, daemonLogs []*models.DaemonLog) error {
	return s.tx.DaemonLog.WithContext(ctx).CreateInBatches(daemonLogs, 100)
}

// Delete delete a daemon log record with specific id
func (s *daemonService) Delete(ctx context.Context, id int64) error {
	matched, err := s.tx.DaemonLog.WithContext(ctx).Where(s.tx.DaemonLog.ID.Eq(id)).Delete()
	if err != nil {
		return err
	}
	if matched.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// List lists all daemon log
func (s *daemonService) List(ctx context.Context, pagination types.Pagination, sort types.Sortable) ([]*models.DaemonLog, int64, error) {
	pagination = utils.NormalizePagination(pagination)
	query := s.tx.DaemonLog.WithContext(ctx)
	field, ok := s.tx.DaemonLog.GetFieldByName(ptr.To(sort.Sort))
	if ok {
		switch ptr.To(sort.Method) {
		case enums.SortMethodDesc:
			query = query.Order(field.Desc())
		case enums.SortMethodAsc:
			query = query.Order(field)
		default:
			query = query.Order(s.tx.DaemonLog.UpdatedAt.Desc())
		}
	} else {
		query = query.Order(s.tx.DaemonLog.UpdatedAt.Desc())
	}
	return query.FindByPage(ptr.To(pagination.Limit)*(ptr.To(pagination.Page)-1), ptr.To(pagination.Limit))
}

// GetGcRepositoryRunner ...
func (s *daemonService) GetGcRepositoryRunner(ctx context.Context, runnerID int64) (*models.DaemonGcRepositoryRunner, error) {
	return s.tx.DaemonGcRepositoryRunner.WithContext(ctx).Where(s.tx.DaemonGcRepositoryRunner.ID.Eq(runnerID)).First()
}

// CreateGcRepositoryRunner ...
func (s *daemonService) CreateGcRepositoryRunner(ctx context.Context, runnerObj *models.DaemonGcRepositoryRunner) error {
	return s.tx.DaemonGcRepositoryRunner.WithContext(ctx).Create(runnerObj)
}

// GetLastGcRepositoryRunner ...
func (s *daemonService) GetLastGcRepositoryRunner(ctx context.Context, namespaceID *int64) (*models.DaemonGcRepositoryRunner, error) {
	query := s.tx.DaemonGcRepositoryRunner.WithContext(ctx)
	if namespaceID == nil {
		query = query.Where(s.tx.DaemonGcRepositoryRunner.NamespaceID.IsNull())
	} else {
		query = query.Where(s.tx.DaemonGcRepositoryRunner.NamespaceID.Eq(ptr.To(namespaceID)))
	}
	return query.Order(s.tx.DaemonGcRepositoryRunner.CreatedAt.Desc()).First()
}

// CreateGcRepositoryRecord ...
func (s *daemonService) CreateGcRepositoryRecords(ctx context.Context, records []*models.DaemonGcRepositoryRecord) error {
	return s.tx.DaemonGcRepositoryRecord.WithContext(ctx).CreateInBatches(records, consts.InsertBatchSize)
}

// UpdateGcRepositoryRunner ...
func (s *daemonService) UpdateGcRepositoryRunner(ctx context.Context, runnerID int64, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}
	matched, err := s.tx.DaemonGcRepositoryRunner.WithContext(ctx).Where(s.tx.DaemonGcRepositoryRunner.ID.Eq(runnerID)).Updates(updates)
	if err != nil {
		return err
	}
	if matched.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ListLastGcArtifactRunner ...
func (s *daemonService) ListLastGcArtifactRunner(ctx context.Context, namespaceID *int64, pagination types.Pagination, sort types.Sortable) ([]*models.DaemonGcArtifactRunner, int64, error) {
	pagination = utils.NormalizePagination(pagination)
	query := s.tx.DaemonGcArtifactRunner.WithContext(ctx)
	if namespaceID == nil {
		query = query.Where(s.tx.DaemonGcRepositoryRunner.NamespaceID.IsNull())
	} else {
		query = query.Where(s.tx.DaemonGcRepositoryRunner.NamespaceID.Eq(ptr.To(namespaceID)))
	}
	field, ok := s.tx.DaemonGcArtifactRunner.GetFieldByName(ptr.To(sort.Sort))
	if ok {
		switch ptr.To(sort.Method) {
		case enums.SortMethodDesc:
			query = query.Order(field.Desc())
		case enums.SortMethodAsc:
			query = query.Order(field)
		default:
			query = query.Order(s.tx.DaemonGcArtifactRunner.UpdatedAt.Desc())
		}
	} else {
		query = query.Order(s.tx.DaemonGcArtifactRunner.UpdatedAt.Desc())
	}
	return query.FindByPage(ptr.To(pagination.Limit)*(ptr.To(pagination.Page)-1), ptr.To(pagination.Limit))
}

// ListLastGcBlobRunner ...
func (s *daemonService) ListLastGcBlobRunner(ctx context.Context, pagination types.Pagination, sort types.Sortable) ([]*models.DaemonGcBlobRunner, int64, error) {
	pagination = utils.NormalizePagination(pagination)
	query := s.tx.DaemonGcBlobRunner.WithContext(ctx)
	field, ok := s.tx.DaemonGcBlobRunner.GetFieldByName(ptr.To(sort.Sort))
	if ok {
		switch ptr.To(sort.Method) {
		case enums.SortMethodDesc:
			query = query.Order(field.Desc())
		case enums.SortMethodAsc:
			query = query.Order(field)
		default:
			query = query.Order(s.tx.DaemonGcBlobRunner.UpdatedAt.Desc())
		}
	} else {
		query = query.Order(s.tx.DaemonGcBlobRunner.UpdatedAt.Desc())
	}
	return query.FindByPage(ptr.To(pagination.Limit)*(ptr.To(pagination.Page)-1), ptr.To(pagination.Limit))
}

// ListGcRepositoryRunners ...
func (s *daemonService) ListGcRepositoryRunners(ctx context.Context, namespaceID *int64, pagination types.Pagination, sort types.Sortable) ([]*models.DaemonGcRepositoryRunner, int64, error) {
	pagination = utils.NormalizePagination(pagination)
	query := s.tx.DaemonGcRepositoryRunner.WithContext(ctx)
	if namespaceID == nil {
		query = query.Where(s.tx.DaemonGcRepositoryRunner.NamespaceID.IsNull())
	} else {
		query = query.Where(s.tx.DaemonGcRepositoryRunner.NamespaceID.Eq(ptr.To(namespaceID)))
	}
	field, ok := s.tx.DaemonGcRepositoryRunner.GetFieldByName(ptr.To(sort.Sort))
	if ok {
		switch ptr.To(sort.Method) {
		case enums.SortMethodDesc:
			query = query.Order(field.Desc())
		case enums.SortMethodAsc:
			query = query.Order(field)
		default:
			query = query.Order(s.tx.DaemonGcRepositoryRunner.UpdatedAt.Desc())
		}
	} else {
		query = query.Order(s.tx.DaemonGcRepositoryRunner.UpdatedAt.Desc())
	}
	return query.FindByPage(ptr.To(pagination.Limit)*(ptr.To(pagination.Page)-1), ptr.To(pagination.Limit))
}

// ListGcRepositoryRecords lists all gc repository records.
func (s *daemonService) ListGcRepositoryRecords(ctx context.Context, runnerID int64, pagination types.Pagination, sort types.Sortable) ([]*models.DaemonGcRepositoryRecord, int64, error) {
	pagination = utils.NormalizePagination(pagination)
	query := s.tx.DaemonGcRepositoryRecord.WithContext(ctx).Where(s.tx.DaemonGcRepositoryRecord.RunnerID.Eq(runnerID))
	field, ok := s.tx.DaemonGcRepositoryRecord.GetFieldByName(ptr.To(sort.Sort))
	if ok {
		switch ptr.To(sort.Method) {
		case enums.SortMethodDesc:
			query = query.Order(field.Desc())
		case enums.SortMethodAsc:
			query = query.Order(field)
		default:
			query = query.Order(s.tx.DaemonGcRepositoryRecord.UpdatedAt.Desc())
		}
	} else {
		query = query.Order(s.tx.DaemonGcRepositoryRecord.UpdatedAt.Desc())
	}
	return query.FindByPage(ptr.To(pagination.Limit)*(ptr.To(pagination.Page)-1), ptr.To(pagination.Limit))
}

// ListGcBlobRecords ...
func (s *daemonService) ListGcBlobRecords(ctx context.Context, runnerID int64, pagination types.Pagination, sort types.Sortable) ([]*models.DaemonGcBlobRecord, int64, error) {
	pagination = utils.NormalizePagination(pagination)
	query := s.tx.DaemonGcBlobRecord.WithContext(ctx).Where(s.tx.DaemonGcBlobRecord.RunnerID.Eq(runnerID))
	field, ok := s.tx.DaemonGcBlobRecord.GetFieldByName(ptr.To(sort.Sort))
	if ok {
		switch ptr.To(sort.Method) {
		case enums.SortMethodDesc:
			query = query.Order(field.Desc())
		case enums.SortMethodAsc:
			query = query.Order(field)
		default:
			query = query.Order(s.tx.DaemonGcBlobRecord.UpdatedAt.Desc())
		}
	} else {
		query = query.Order(s.tx.DaemonGcBlobRecord.UpdatedAt.Desc())
	}
	return query.FindByPage(ptr.To(pagination.Limit)*(ptr.To(pagination.Page)-1), ptr.To(pagination.Limit))
}

// ListGcArtifactRecords ...
func (s *daemonService) ListGcArtifactRecords(ctx context.Context, runnerID int64, pagination types.Pagination, sort types.Sortable) ([]*models.DaemonGcArtifactRecord, int64, error) {
	pagination = utils.NormalizePagination(pagination)
	query := s.tx.DaemonGcArtifactRecord.WithContext(ctx).Where(s.tx.DaemonGcArtifactRecord.RunnerID.Eq(runnerID))
	field, ok := s.tx.DaemonGcArtifactRecord.GetFieldByName(ptr.To(sort.Sort))
	if ok {
		switch ptr.To(sort.Method) {
		case enums.SortMethodDesc:
			query = query.Order(field.Desc())
		case enums.SortMethodAsc:
			query = query.Order(field)
		default:
			query = query.Order(s.tx.DaemonGcArtifactRecord.UpdatedAt.Desc())
		}
	} else {
		query = query.Order(s.tx.DaemonGcArtifactRecord.UpdatedAt.Desc())
	}
	return query.FindByPage(ptr.To(pagination.Limit)*(ptr.To(pagination.Page)-1), ptr.To(pagination.Limit))
}

// GetGcBlobRunner ...
func (s *daemonService) GetGcBlobRunner(ctx context.Context, runnerID int64) (*models.DaemonGcBlobRunner, error) {
	return s.tx.DaemonGcBlobRunner.WithContext(ctx).Where(s.tx.DaemonGcBlobRunner.ID.Eq(runnerID)).First()
}

// GetLastGcBlobRunner ...
func (s *daemonService) GetLastGcBlobRunner(ctx context.Context) (*models.DaemonGcBlobRunner, error) {
	return s.tx.DaemonGcBlobRunner.WithContext(ctx).Order(s.tx.DaemonGcBlobRunner.CreatedAt.Desc()).First()
}

// CreateGcBlobRunner ...
func (s *daemonService) CreateGcBlobRunner(ctx context.Context, runnerObj *models.DaemonGcBlobRunner) error {
	return s.tx.DaemonGcBlobRunner.WithContext(ctx).Create(runnerObj)
}

// CreateGcBlobRecords ...
func (s *daemonService) CreateGcBlobRecords(ctx context.Context, records []*models.DaemonGcBlobRecord) error {
	return s.tx.DaemonGcBlobRecord.WithContext(ctx).CreateInBatches(records, consts.InsertBatchSize)
}

// UpdateGcBlobRunner ...
func (s *daemonService) UpdateGcBlobRunner(ctx context.Context, runnerID int64, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}
	matched, err := s.tx.DaemonGcRepositoryRunner.WithContext(ctx).Where(s.tx.DaemonGcBlobRunner.ID.Eq(runnerID)).Updates(updates)
	if err != nil {
		return err
	}
	if matched.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// GetGcArtifactRunner ...
func (s *daemonService) GetGcArtifactRunner(ctx context.Context, runnerID int64) (*models.DaemonGcArtifactRunner, error) {
	return s.tx.DaemonGcArtifactRunner.WithContext(ctx).Where(s.tx.DaemonGcArtifactRunner.ID.Eq(runnerID)).First()
}

// GetLastGcArtifactRunner ...
func (s *daemonService) GetLastGcArtifactRunner(ctx context.Context, namespaceID *int64) (*models.DaemonGcArtifactRunner, error) {
	query := s.tx.DaemonGcArtifactRunner.WithContext(ctx)
	if namespaceID == nil {
		query = query.Where(s.tx.DaemonGcArtifactRunner.NamespaceID.IsNull())
	} else {
		query = query.Where(s.tx.DaemonGcArtifactRunner.NamespaceID.Eq(ptr.To(namespaceID)))
	}
	return query.Order(s.tx.DaemonGcArtifactRunner.CreatedAt.Desc()).First()
}

// CreateGcArtifactRunner ...
func (s *daemonService) CreateGcArtifactRunner(ctx context.Context, runnerObj *models.DaemonGcArtifactRunner) error {
	return s.tx.DaemonGcArtifactRunner.WithContext(ctx).Create(runnerObj)
}

// CreateGcArtifactRecords ...
func (s *daemonService) CreateGcArtifactRecords(ctx context.Context, records []*models.DaemonGcArtifactRecord) error {
	return s.tx.DaemonGcArtifactRecord.WithContext(ctx).CreateInBatches(records, consts.InsertBatchSize)
}

// UpdateGcArtifactRunner ...
func (s *daemonService) UpdateGcArtifactRunner(ctx context.Context, runnerID int64, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}
	matched, err := s.tx.DaemonGcArtifactRunner.WithContext(ctx).Where(s.tx.DaemonGcArtifactRunner.ID.Eq(runnerID)).Updates(updates)
	if err != nil {
		return err
	}
	if matched.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
