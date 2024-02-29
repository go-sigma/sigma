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
	// GetGcTagRule ...
	GetGcTagRule(ctx context.Context, namespaceID *int64) (*models.DaemonGcTagRule, error)
	// CreateGcTagRule ...
	CreateGcTagRule(ctx context.Context, ruleObj *models.DaemonGcTagRule) error
	// UpdateGcTagRule ...
	UpdateGcTagRule(ctx context.Context, ruleID int64, updates map[string]any) error
	// GetGcTagLatestRunner ...
	GetGcTagLatestRunner(ctx context.Context, ruleID int64) (*models.DaemonGcTagRunner, error)
	// GetGcTagRunner ...
	GetGcTagRunner(ctx context.Context, runnerID int64) (*models.DaemonGcTagRunner, error)
	// ListGcTagRunners ...
	ListGcTagRunners(ctx context.Context, ruleID int64, pagination types.Pagination, sort types.Sortable) ([]*models.DaemonGcTagRunner, int64, error)
	// CreateGcTagRunner ...
	CreateGcTagRunner(ctx context.Context, runnerObj *models.DaemonGcTagRunner) error
	// UpdateGcTagRunner ...
	UpdateGcTagRunner(ctx context.Context, runnerID int64, updates map[string]any) error
	// CreateGcTagRecords ...
	CreateGcTagRecords(ctx context.Context, recordObjs []*models.DaemonGcTagRecord) error
	// ListGcTagRecords ...
	ListGcTagRecords(ctx context.Context, runnerID int64, pagination types.Pagination, sort types.Sortable) ([]*models.DaemonGcTagRecord, int64, error)
	// GetGcTagRecord ...
	GetGcTagRecord(ctx context.Context, recordID int64) (*models.DaemonGcTagRecord, error)

	// GetGcRepositoryRule ...
	GetGcRepositoryRule(ctx context.Context, namespaceID *int64) (*models.DaemonGcRepositoryRule, error)
	// CreateGcRepositoryRule ...
	CreateGcRepositoryRule(ctx context.Context, ruleObj *models.DaemonGcRepositoryRule) error
	// UpdateGcRepositoryRule ...
	UpdateGcRepositoryRule(ctx context.Context, ruleID int64, updates map[string]any) error
	// GetGcRepositoryLatestRunner ...
	GetGcRepositoryLatestRunner(ctx context.Context, ruleID int64) (*models.DaemonGcRepositoryRunner, error)
	// GetGcRepositoryRunner ...
	GetGcRepositoryRunner(ctx context.Context, runnerID int64) (*models.DaemonGcRepositoryRunner, error)
	// ListGcRepositoryRunners ...
	ListGcRepositoryRunners(ctx context.Context, ruleID int64, pagination types.Pagination, sort types.Sortable) ([]*models.DaemonGcRepositoryRunner, int64, error)
	// CreateGcRepositoryRunner ...
	CreateGcRepositoryRunner(ctx context.Context, runnerObj *models.DaemonGcRepositoryRunner) error
	// UpdateGcRepositoryRunner ...
	UpdateGcRepositoryRunner(ctx context.Context, runnerID int64, updates map[string]any) error
	// CreateGcRepositoryRecords ...
	CreateGcRepositoryRecords(ctx context.Context, records []*models.DaemonGcRepositoryRecord) error
	// ListGcRepositoryRecords lists all gc repository records.
	ListGcRepositoryRecords(ctx context.Context, runnerID int64, pagination types.Pagination, sort types.Sortable) ([]*models.DaemonGcRepositoryRecord, int64, error)
	// GetGcRepositoryRecord ...
	GetGcRepositoryRecord(ctx context.Context, recordID int64) (*models.DaemonGcRepositoryRecord, error)

	// GetGcArtifactRule ...
	GetGcArtifactRule(ctx context.Context, namespaceID *int64) (*models.DaemonGcArtifactRule, error)
	// CreateGcArtifactRule ...
	CreateGcArtifactRule(ctx context.Context, ruleObj *models.DaemonGcArtifactRule) error
	// UpdateGcArtifactRule ...
	UpdateGcArtifactRule(ctx context.Context, ruleID int64, updates map[string]any) error
	// GetGcArtifactLatestRunner ...
	GetGcArtifactLatestRunner(ctx context.Context, ruleID int64) (*models.DaemonGcArtifactRunner, error)
	// GetGcArtifactRunner ...
	GetGcArtifactRunner(ctx context.Context, runnerID int64) (*models.DaemonGcArtifactRunner, error)
	// ListGcArtifactRunners ...
	ListGcArtifactRunners(ctx context.Context, ruleID int64, pagination types.Pagination, sort types.Sortable) ([]*models.DaemonGcArtifactRunner, int64, error)
	// CreateGcArtifactRunner ...
	CreateGcArtifactRunner(ctx context.Context, runnerObj *models.DaemonGcArtifactRunner) error
	// UpdateGcArtifactRunner ...
	UpdateGcArtifactRunner(ctx context.Context, runnerID int64, updates map[string]any) error
	// CreateGcArtifactRecords ...
	CreateGcArtifactRecords(ctx context.Context, records []*models.DaemonGcArtifactRecord) error
	// ListGcArtifactRecords ...
	ListGcArtifactRecords(ctx context.Context, runnerID int64, pagination types.Pagination, sort types.Sortable) ([]*models.DaemonGcArtifactRecord, int64, error)
	// GetGcArtifactRecord ...
	GetGcArtifactRecord(ctx context.Context, recordID int64) (*models.DaemonGcArtifactRecord, error)

	// GetGcBlobRule ...
	GetGcBlobRule(ctx context.Context) (*models.DaemonGcBlobRule, error)
	// CreateGcBlobRule ...
	CreateGcBlobRule(ctx context.Context, ruleObj *models.DaemonGcBlobRule) error
	// UpdateGcBlobRule ...
	UpdateGcBlobRule(ctx context.Context, ruleID int64, updates map[string]any) error
	// GetGcBlobLatestRunner ...
	GetGcBlobLatestRunner(ctx context.Context, ruleID int64) (*models.DaemonGcBlobRunner, error)
	// GetGcBlobRunner ...
	GetGcBlobRunner(ctx context.Context, runnerID int64) (*models.DaemonGcBlobRunner, error)
	// ListGcBlobRunners ...
	ListGcBlobRunners(ctx context.Context, ruleID int64, pagination types.Pagination, sort types.Sortable) ([]*models.DaemonGcBlobRunner, int64, error)
	// CreateGcBlobRunner ...
	CreateGcBlobRunner(ctx context.Context, runnerObj *models.DaemonGcBlobRunner) error
	// UpdateGcBlobRunner ...
	UpdateGcBlobRunner(ctx context.Context, runnerID int64, updates map[string]any) error
	// CreateGcBlobRecords ...
	CreateGcBlobRecords(ctx context.Context, records []*models.DaemonGcBlobRecord) error
	// ListGcBlobRecords ...
	ListGcBlobRecords(ctx context.Context, runnerID int64, pagination types.Pagination, sort types.Sortable) ([]*models.DaemonGcBlobRecord, int64, error)
	// GetGcBlobRecord ...
	GetGcBlobRecord(ctx context.Context, recordID int64) (*models.DaemonGcBlobRecord, error)
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

// New ...
func (s *daemonServiceFactory) New(txs ...*query.Query) DaemonService {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &daemonService{
		tx: tx,
	}
}

// GetGcTagRule ...
func (s *daemonService) GetGcTagRule(ctx context.Context, namespaceID *int64) (*models.DaemonGcTagRule, error) {
	q := s.tx.DaemonGcTagRule.WithContext(ctx)
	if namespaceID == nil {
		q = q.Where(s.tx.DaemonGcTagRule.NamespaceID.IsNull())
	} else {
		q = q.Where(s.tx.DaemonGcTagRule.NamespaceID.Eq(ptr.To(namespaceID)))
	}
	return q.First()
}

// CreateGcTagRule ...
func (s *daemonService) CreateGcTagRule(ctx context.Context, ruleObj *models.DaemonGcTagRule) error {
	return s.tx.DaemonGcTagRule.WithContext(ctx).Create(ruleObj)
}

// UpdateGcTagRule ...
func (s *daemonService) UpdateGcTagRule(ctx context.Context, ruleID int64, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	matched, err := s.tx.DaemonGcTagRule.WithContext(ctx).Where(s.tx.DaemonGcTagRule.ID.Eq(ruleID)).Updates(updates)
	if err != nil {
		return err
	}
	if matched.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// GetGcTagLatestRunner ...
func (s *daemonService) GetGcTagLatestRunner(ctx context.Context, ruleID int64) (*models.DaemonGcTagRunner, error) {
	return s.tx.DaemonGcTagRunner.WithContext(ctx).
		Where(s.tx.DaemonGcTagRunner.RuleID.Eq(ruleID)).
		Order(s.tx.DaemonGcTagRunner.CreatedAt.Desc()).First()
}

// GetGcTagRunner ...
func (s *daemonService) GetGcTagRunner(ctx context.Context, runnerID int64) (*models.DaemonGcTagRunner, error) {
	return s.tx.DaemonGcTagRunner.WithContext(ctx).
		Where(s.tx.DaemonGcTagRunner.ID.Eq(runnerID)).
		Preload(s.tx.DaemonGcTagRunner.Rule).
		Preload(s.tx.DaemonGcTagRunner.OperateUser).
		First()
}

// ListGcTagRunners ...
func (s *daemonService) ListGcTagRunners(ctx context.Context, ruleID int64, pagination types.Pagination, sort types.Sortable) ([]*models.DaemonGcTagRunner, int64, error) {
	pagination = utils.NormalizePagination(pagination)
	q := s.tx.DaemonGcTagRunner.WithContext(ctx).Where(s.tx.DaemonGcTagRunner.RuleID.Eq(ruleID))
	field, ok := s.tx.DaemonGcTagRunner.GetFieldByName(ptr.To(sort.Sort))
	if ok {
		switch ptr.To(sort.Method) {
		case enums.SortMethodDesc:
			q = q.Order(field.Desc())
		case enums.SortMethodAsc:
			q = q.Order(field)
		default:
			q = q.Order(s.tx.DaemonGcTagRunner.UpdatedAt.Desc())
		}
	} else {
		q = q.Order(s.tx.DaemonGcTagRunner.UpdatedAt.Desc())
	}
	return q.FindByPage(ptr.To(pagination.Limit)*(ptr.To(pagination.Page)-1), ptr.To(pagination.Limit))
}

// CreateGcTagRunner ...
func (s *daemonService) CreateGcTagRunner(ctx context.Context, runnerObj *models.DaemonGcTagRunner) error {
	return s.tx.DaemonGcTagRunner.WithContext(ctx).Create(runnerObj)
}

// UpdateGcTagRunner ...
func (s *daemonService) UpdateGcTagRunner(ctx context.Context, runnerID int64, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	_, err := s.tx.DaemonGcTagRunner.WithContext(ctx).Where(s.tx.DaemonGcTagRunner.RuleID.Eq(runnerID)).Updates(updates)
	return err
}

// CreateGcTagRecords ...
func (s *daemonService) CreateGcTagRecords(ctx context.Context, recordObjs []*models.DaemonGcTagRecord) error {
	return s.tx.DaemonGcTagRecord.WithContext(ctx).CreateInBatches(recordObjs, consts.InsertBatchSize)
}

// ListGcTagRecords ...
func (s *daemonService) ListGcTagRecords(ctx context.Context, runnerID int64, pagination types.Pagination, sort types.Sortable) ([]*models.DaemonGcTagRecord, int64, error) {
	pagination = utils.NormalizePagination(pagination)
	q := s.tx.DaemonGcTagRecord.WithContext(ctx).Where(s.tx.DaemonGcTagRecord.RunnerID.Eq(runnerID))
	field, ok := s.tx.DaemonGcTagRecord.GetFieldByName(ptr.To(sort.Sort))
	if ok {
		switch ptr.To(sort.Method) {
		case enums.SortMethodDesc:
			q = q.Order(field.Desc())
		case enums.SortMethodAsc:
			q = q.Order(field)
		default:
			q = q.Order(s.tx.DaemonGcTagRecord.UpdatedAt.Desc())
		}
	} else {
		q = q.Order(s.tx.DaemonGcTagRecord.UpdatedAt.Desc())
	}
	return q.FindByPage(ptr.To(pagination.Limit)*(ptr.To(pagination.Page)-1), ptr.To(pagination.Limit))
}

// GetGcTagRecord ...
func (s *daemonService) GetGcTagRecord(ctx context.Context, recordID int64) (*models.DaemonGcTagRecord, error) {
	return s.tx.DaemonGcTagRecord.WithContext(ctx).Where(s.tx.DaemonGcTagRecord.ID.Eq(recordID)).
		Preload(s.tx.DaemonGcTagRecord.Runner).
		Preload(s.tx.DaemonGcTagRecord.Runner.Rule).
		First()
}

// CreateGcRepositoryRule ...
func (s *daemonService) CreateGcRepositoryRule(ctx context.Context, ruleObj *models.DaemonGcRepositoryRule) error {
	return s.tx.DaemonGcRepositoryRule.WithContext(ctx).Create(ruleObj)
}

// UpdateGcRepositoryRule ...
func (s *daemonService) UpdateGcRepositoryRule(ctx context.Context, ruleID int64, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	matched, err := s.tx.DaemonGcRepositoryRule.WithContext(ctx).Where(s.tx.DaemonGcRepositoryRule.ID.Eq(ruleID)).Updates(updates)
	if err != nil {
		return err
	}
	if matched.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// GetGcRepositoryRule ...
func (s *daemonService) GetGcRepositoryRule(ctx context.Context, namespaceID *int64) (*models.DaemonGcRepositoryRule, error) {
	q := s.tx.DaemonGcRepositoryRule.WithContext(ctx)
	if namespaceID == nil {
		q = q.Where(s.tx.DaemonGcRepositoryRule.NamespaceID.IsNull())
	} else {
		q = q.Where(s.tx.DaemonGcRepositoryRule.NamespaceID.Eq(ptr.To(namespaceID)))
	}
	return q.First()
}

// GetGcRepositoryLatestRunner ...
func (s *daemonService) GetGcRepositoryLatestRunner(ctx context.Context, ruleID int64) (*models.DaemonGcRepositoryRunner, error) {
	return s.tx.DaemonGcRepositoryRunner.WithContext(ctx).
		Where(s.tx.DaemonGcRepositoryRunner.RuleID.Eq(ruleID)).
		Order(s.tx.DaemonGcRepositoryRunner.CreatedAt.Desc()).First()
}

// GetGcRepositoryRunner ...
func (s *daemonService) GetGcRepositoryRunner(ctx context.Context, runnerID int64) (*models.DaemonGcRepositoryRunner, error) {
	return s.tx.DaemonGcRepositoryRunner.WithContext(ctx).
		Where(s.tx.DaemonGcRepositoryRunner.ID.Eq(runnerID)).
		Preload(s.tx.DaemonGcRepositoryRunner.Rule).
		Preload(s.tx.DaemonGcRepositoryRunner.OperateUser).
		Order(s.tx.DaemonGcRepositoryRunner.CreatedAt.Desc()).First()
}

// ListGcRepositoryRunners ...
func (s *daemonService) ListGcRepositoryRunners(ctx context.Context, ruleID int64, pagination types.Pagination, sort types.Sortable) ([]*models.DaemonGcRepositoryRunner, int64, error) {
	pagination = utils.NormalizePagination(pagination)
	q := s.tx.DaemonGcRepositoryRunner.WithContext(ctx).Where(s.tx.DaemonGcRepositoryRunner.RuleID.Eq(ruleID))
	field, ok := s.tx.DaemonGcRepositoryRunner.GetFieldByName(ptr.To(sort.Sort))
	if ok {
		switch ptr.To(sort.Method) {
		case enums.SortMethodDesc:
			q = q.Order(field.Desc())
		case enums.SortMethodAsc:
			q = q.Order(field)
		default:
			q = q.Order(s.tx.DaemonGcRepositoryRunner.UpdatedAt.Desc())
		}
	} else {
		q = q.Order(s.tx.DaemonGcRepositoryRunner.UpdatedAt.Desc())
	}
	return q.FindByPage(ptr.To(pagination.Limit)*(ptr.To(pagination.Page)-1), ptr.To(pagination.Limit))
}

// CreateGcRepositoryRunner ...
func (s *daemonService) CreateGcRepositoryRunner(ctx context.Context, runnerObj *models.DaemonGcRepositoryRunner) error {
	return s.tx.DaemonGcRepositoryRunner.WithContext(ctx).Create(runnerObj)
}

// UpdateGcRepositoryRunner ...
func (s *daemonService) UpdateGcRepositoryRunner(ctx context.Context, runnerID int64, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	_, err := s.tx.DaemonGcRepositoryRunner.WithContext(ctx).Where(s.tx.DaemonGcRepositoryRunner.ID.Eq(runnerID)).Updates(updates)
	return err
}

// CreateGcRepositoryRecords ...
func (s *daemonService) CreateGcRepositoryRecords(ctx context.Context, records []*models.DaemonGcRepositoryRecord) error {
	return s.tx.DaemonGcRepositoryRecord.WithContext(ctx).CreateInBatches(records, consts.InsertBatchSize)
}

// ListGcRepositoryRecords lists all gc repository records.
func (s *daemonService) ListGcRepositoryRecords(ctx context.Context, runnerID int64, pagination types.Pagination, sort types.Sortable) ([]*models.DaemonGcRepositoryRecord, int64, error) {
	pagination = utils.NormalizePagination(pagination)
	q := s.tx.DaemonGcRepositoryRecord.WithContext(ctx).Where(s.tx.DaemonGcRepositoryRecord.RunnerID.Eq(runnerID))
	field, ok := s.tx.DaemonGcRepositoryRecord.GetFieldByName(ptr.To(sort.Sort))
	if ok {
		switch ptr.To(sort.Method) {
		case enums.SortMethodDesc:
			q = q.Order(field.Desc())
		case enums.SortMethodAsc:
			q = q.Order(field)
		default:
			q = q.Order(s.tx.DaemonGcRepositoryRecord.UpdatedAt.Desc())
		}
	} else {
		q = q.Order(s.tx.DaemonGcRepositoryRecord.UpdatedAt.Desc())
	}
	return q.FindByPage(ptr.To(pagination.Limit)*(ptr.To(pagination.Page)-1), ptr.To(pagination.Limit))
}

// GetGcRepositoryRecord ...
func (s *daemonService) GetGcRepositoryRecord(ctx context.Context, recordID int64) (*models.DaemonGcRepositoryRecord, error) {
	return s.tx.DaemonGcRepositoryRecord.WithContext(ctx).Where(s.tx.DaemonGcRepositoryRecord.ID.Eq(recordID)).
		Preload(s.tx.DaemonGcRepositoryRecord.Runner).
		Preload(s.tx.DaemonGcRepositoryRecord.Runner.Rule).
		First()
}

// GetGcArtifactRule ...
func (s *daemonService) GetGcArtifactRule(ctx context.Context, namespaceID *int64) (*models.DaemonGcArtifactRule, error) {
	q := s.tx.DaemonGcArtifactRule.WithContext(ctx)
	if namespaceID == nil {
		q = q.Where(s.tx.DaemonGcArtifactRule.NamespaceID.IsNull())
	} else {
		q = q.Where(s.tx.DaemonGcArtifactRule.NamespaceID.Eq(ptr.To(namespaceID)))
	}
	return q.First()
}

// CreateGcArtifactRule ...
func (s *daemonService) CreateGcArtifactRule(ctx context.Context, ruleObj *models.DaemonGcArtifactRule) error {
	return s.tx.DaemonGcArtifactRule.WithContext(ctx).Create(ruleObj)
}

// UpdateGcArtifactRule ...
func (s *daemonService) UpdateGcArtifactRule(ctx context.Context, ruleID int64, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	matched, err := s.tx.DaemonGcArtifactRule.WithContext(ctx).Where(s.tx.DaemonGcArtifactRule.ID.Eq(ruleID)).Updates(updates)
	if err != nil {
		return err
	}
	if matched.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// GetGcArtifactLatestRunner ...
func (s *daemonService) GetGcArtifactLatestRunner(ctx context.Context, ruleID int64) (*models.DaemonGcArtifactRunner, error) {
	return s.tx.DaemonGcArtifactRunner.WithContext(ctx).
		Where(s.tx.DaemonGcArtifactRunner.RuleID.Eq(ruleID)).
		Order(s.tx.DaemonGcArtifactRunner.CreatedAt.Desc()).First()
}

// GetGcArtifactRunner ...
func (s *daemonService) GetGcArtifactRunner(ctx context.Context, runnerID int64) (*models.DaemonGcArtifactRunner, error) {
	return s.tx.DaemonGcArtifactRunner.WithContext(ctx).
		Where(s.tx.DaemonGcArtifactRunner.ID.Eq(runnerID)).
		Preload(s.tx.DaemonGcArtifactRunner.Rule).
		Preload(s.tx.DaemonGcArtifactRunner.OperateUser).
		First()
}

// ListGcArtifactRunners ...
func (s *daemonService) ListGcArtifactRunners(ctx context.Context, ruleID int64, pagination types.Pagination, sort types.Sortable) ([]*models.DaemonGcArtifactRunner, int64, error) {
	pagination = utils.NormalizePagination(pagination)
	q := s.tx.DaemonGcArtifactRunner.WithContext(ctx).Where(s.tx.DaemonGcArtifactRunner.RuleID.Eq(ruleID))
	field, ok := s.tx.DaemonGcArtifactRunner.GetFieldByName(ptr.To(sort.Sort))
	if ok {
		switch ptr.To(sort.Method) {
		case enums.SortMethodDesc:
			q = q.Order(field.Desc())
		case enums.SortMethodAsc:
			q = q.Order(field)
		default:
			q = q.Order(s.tx.DaemonGcArtifactRunner.UpdatedAt.Desc())
		}
	} else {
		q = q.Order(s.tx.DaemonGcArtifactRunner.UpdatedAt.Desc())
	}
	return q.FindByPage(ptr.To(pagination.Limit)*(ptr.To(pagination.Page)-1), ptr.To(pagination.Limit))
}

// CreateGcArtifactRunner ...
func (s *daemonService) CreateGcArtifactRunner(ctx context.Context, runnerObj *models.DaemonGcArtifactRunner) error {
	return s.tx.DaemonGcArtifactRunner.WithContext(ctx).Create(runnerObj)
}

// UpdateGcArtifactRunner ...
func (s *daemonService) UpdateGcArtifactRunner(ctx context.Context, runnerID int64, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	_, err := s.tx.DaemonGcArtifactRunner.WithContext(ctx).Where(s.tx.DaemonGcArtifactRunner.ID.Eq(runnerID)).Updates(updates)
	return err
}

// CreateGcArtifactRecords ...
func (s *daemonService) CreateGcArtifactRecords(ctx context.Context, records []*models.DaemonGcArtifactRecord) error {
	return s.tx.DaemonGcArtifactRecord.WithContext(ctx).CreateInBatches(records, consts.InsertBatchSize)
}

// ListGcArtifactRecords ...
func (s *daemonService) ListGcArtifactRecords(ctx context.Context, runnerID int64, pagination types.Pagination, sort types.Sortable) ([]*models.DaemonGcArtifactRecord, int64, error) {
	pagination = utils.NormalizePagination(pagination)
	q := s.tx.DaemonGcArtifactRecord.WithContext(ctx).Where(s.tx.DaemonGcArtifactRecord.RunnerID.Eq(runnerID))
	field, ok := s.tx.DaemonGcArtifactRecord.GetFieldByName(ptr.To(sort.Sort))
	if ok {
		switch ptr.To(sort.Method) {
		case enums.SortMethodDesc:
			q = q.Order(field.Desc())
		case enums.SortMethodAsc:
			q = q.Order(field)
		default:
			q = q.Order(s.tx.DaemonGcArtifactRecord.UpdatedAt.Desc())
		}
	} else {
		q = q.Order(s.tx.DaemonGcArtifactRecord.UpdatedAt.Desc())
	}
	return q.FindByPage(ptr.To(pagination.Limit)*(ptr.To(pagination.Page)-1), ptr.To(pagination.Limit))
}

// GetGcArtifactRecord ...
func (s *daemonService) GetGcArtifactRecord(ctx context.Context, recordID int64) (*models.DaemonGcArtifactRecord, error) {
	return s.tx.DaemonGcArtifactRecord.WithContext(ctx).Where(s.tx.DaemonGcArtifactRecord.ID.Eq(recordID)).
		Preload(s.tx.DaemonGcArtifactRecord.Runner).
		Preload(s.tx.DaemonGcArtifactRecord.Runner.Rule).
		First()
}

// GetGcBlobRule ...
func (s *daemonService) GetGcBlobRule(ctx context.Context) (*models.DaemonGcBlobRule, error) {
	return s.tx.DaemonGcBlobRule.WithContext(ctx).First()
}

// CreateGcBlobRule ...
func (s *daemonService) CreateGcBlobRule(ctx context.Context, ruleObj *models.DaemonGcBlobRule) error {
	return s.tx.DaemonGcBlobRule.WithContext(ctx).Create(ruleObj)
}

// UpdateGcBlobRule ...
func (s *daemonService) UpdateGcBlobRule(ctx context.Context, ruleID int64, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	matched, err := s.tx.DaemonGcBlobRule.WithContext(ctx).Where(s.tx.DaemonGcBlobRule.ID.Eq(ruleID)).Updates(updates)
	if err != nil {
		return err
	}
	if matched.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// GetGcBlobLatestRunner ...
func (s *daemonService) GetGcBlobLatestRunner(ctx context.Context, ruleID int64) (*models.DaemonGcBlobRunner, error) {
	return s.tx.DaemonGcBlobRunner.WithContext(ctx).
		Where(s.tx.DaemonGcBlobRunner.RuleID.Eq(ruleID)).
		Order(s.tx.DaemonGcBlobRunner.CreatedAt.Desc()).First()
}

// GetGcBlobRunner ...
func (s *daemonService) GetGcBlobRunner(ctx context.Context, runnerID int64) (*models.DaemonGcBlobRunner, error) {
	return s.tx.DaemonGcBlobRunner.WithContext(ctx).
		Where(s.tx.DaemonGcBlobRunner.ID.Eq(runnerID)).
		Preload(s.tx.DaemonGcTagRunner.Rule).
		Preload(s.tx.DaemonGcTagRunner.OperateUser).
		First()
}

// ListGcBlobRunners ...
func (s *daemonService) ListGcBlobRunners(ctx context.Context, ruleID int64, pagination types.Pagination, sort types.Sortable) ([]*models.DaemonGcBlobRunner, int64, error) {
	pagination = utils.NormalizePagination(pagination)
	q := s.tx.DaemonGcBlobRunner.WithContext(ctx).Where(s.tx.DaemonGcBlobRunner.RuleID.Eq(ruleID))
	field, ok := s.tx.DaemonGcBlobRunner.GetFieldByName(ptr.To(sort.Sort))
	if ok {
		switch ptr.To(sort.Method) {
		case enums.SortMethodDesc:
			q = q.Order(field.Desc())
		case enums.SortMethodAsc:
			q = q.Order(field)
		default:
			q = q.Order(s.tx.DaemonGcBlobRunner.UpdatedAt.Desc())
		}
	} else {
		q = q.Order(s.tx.DaemonGcBlobRunner.UpdatedAt.Desc())
	}
	return q.FindByPage(ptr.To(pagination.Limit)*(ptr.To(pagination.Page)-1), ptr.To(pagination.Limit))
}

// CreateGcBlobRunner ...
func (s *daemonService) CreateGcBlobRunner(ctx context.Context, runnerObj *models.DaemonGcBlobRunner) error {
	return s.tx.DaemonGcBlobRunner.WithContext(ctx).Create(runnerObj)
}

// UpdateGcBlobRunner ...
func (s *daemonService) UpdateGcBlobRunner(ctx context.Context, runnerID int64, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	_, err := s.tx.DaemonGcBlobRunner.WithContext(ctx).Where(s.tx.DaemonGcBlobRunner.ID.Eq(runnerID)).Updates(updates)
	return err
}

// CreateGcBlobRecords ...
func (s *daemonService) CreateGcBlobRecords(ctx context.Context, records []*models.DaemonGcBlobRecord) error {
	return s.tx.DaemonGcBlobRecord.WithContext(ctx).CreateInBatches(records, consts.InsertBatchSize)
}

// ListGcBlobRecords ...
func (s *daemonService) ListGcBlobRecords(ctx context.Context, runnerID int64, pagination types.Pagination, sort types.Sortable) ([]*models.DaemonGcBlobRecord, int64, error) {
	pagination = utils.NormalizePagination(pagination)
	q := s.tx.DaemonGcBlobRecord.WithContext(ctx).Where(s.tx.DaemonGcBlobRecord.RunnerID.Eq(runnerID))
	field, ok := s.tx.DaemonGcBlobRecord.GetFieldByName(ptr.To(sort.Sort))
	if ok {
		switch ptr.To(sort.Method) {
		case enums.SortMethodDesc:
			q = q.Order(field.Desc())
		case enums.SortMethodAsc:
			q = q.Order(field)
		default:
			q = q.Order(s.tx.DaemonGcBlobRecord.UpdatedAt.Desc())
		}
	} else {
		q = q.Order(s.tx.DaemonGcBlobRecord.UpdatedAt.Desc())
	}
	return q.FindByPage(ptr.To(pagination.Limit)*(ptr.To(pagination.Page)-1), ptr.To(pagination.Limit))
}

// GetGcBlobRecord ...
func (s *daemonService) GetGcBlobRecord(ctx context.Context, recordID int64) (*models.DaemonGcBlobRecord, error) {
	return s.tx.DaemonGcBlobRecord.WithContext(ctx).Where(s.tx.DaemonGcBlobRecord.ID.Eq(recordID)).
		Preload(s.tx.DaemonGcBlobRecord.Runner).
		Preload(s.tx.DaemonGcBlobRecord.Runner.Rule).
		First()
}
