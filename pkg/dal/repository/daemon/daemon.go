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

package daemon

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

//go:generate go tool mockgen -destination=daemon_mocks.go -package=daemon github.com/go-sigma/sigma/pkg/dal/repository/daemon DaemonRepository

type DaemonRepository interface {
	GetGcRule(context.Context, enums.Daemon, *string) (*models.DaemonGcRule, error)
	CreateGcRule(context.Context, *models.DaemonGcRule) error
	UpdateGcRule(context.Context, string, map[string]any) error
	GetGcLatestRunner(context.Context, string) (*models.DaemonGcRunner, error)
	GetGcRunner(context.Context, string) (*models.DaemonGcRunner, error)
	ListGcRunners(context.Context, string, api.Pagination, api.Sortable) ([]*models.DaemonGcRunner, int64, error)
	CreateGcRunner(context.Context, *models.DaemonGcRunner) error
	UpdateGcRunner(context.Context, string, map[string]any) error
	CreateGcRecords(context.Context, []*models.DaemonGcRecord) error
	ListGcRecords(context.Context, string, api.Pagination, api.Sortable) ([]*models.DaemonGcRecord, int64, error)
	GetGcRecord(context.Context, string) (*models.DaemonGcRecord, error)
	UpsertGcStorageDeletionTask(context.Context, *models.GcStorageDeletionTask) error
	UpdateGcStorageDeletionTask(context.Context, string, map[string]any) error
}

type daemonRepository struct{ tx *query.Query }

func NewDaemonRepository(txs ...*query.Query) DaemonRepository {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &daemonRepository{tx: tx}
}

func (s *daemonRepository) GetGcRule(ctx context.Context, daemon enums.Daemon, namespaceID *string) (*models.DaemonGcRule, error) {
	q := s.tx.DaemonGcRule.WithContext(ctx).Where(s.tx.DaemonGcRule.Type.Eq(daemon))
	if namespaceID == nil {
		q = q.Where(s.tx.DaemonGcRule.NamespaceID.IsNull())
	} else {
		q = q.Where(s.tx.DaemonGcRule.NamespaceID.Eq(ptr.To(namespaceID)))
	}
	return q.First()
}
func (s *daemonRepository) CreateGcRule(ctx context.Context, rule *models.DaemonGcRule) error {
	return s.tx.DaemonGcRule.WithContext(ctx).Create(rule)
}
func (s *daemonRepository) UpdateGcRule(ctx context.Context, id string, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	matched, err := s.tx.DaemonGcRule.WithContext(ctx).Where(s.tx.DaemonGcRule.ID.Eq(id)).Updates(updates)
	if err != nil {
		return err
	}
	if matched.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
func (s *daemonRepository) GetGcLatestRunner(ctx context.Context, ruleID string) (*models.DaemonGcRunner, error) {
	return s.tx.DaemonGcRunner.WithContext(ctx).Where(s.tx.DaemonGcRunner.RuleID.Eq(ruleID)).Order(s.tx.DaemonGcRunner.CreatedAt.Desc()).First()
}
func (s *daemonRepository) GetGcRunner(ctx context.Context, id string) (*models.DaemonGcRunner, error) {
	return s.tx.DaemonGcRunner.WithContext(ctx).Where(s.tx.DaemonGcRunner.ID.Eq(id)).Preload(s.tx.DaemonGcRunner.Rule).Preload(s.tx.DaemonGcRunner.OperateUser).First()
}
func (s *daemonRepository) ListGcRunners(ctx context.Context, ruleID string, pagination api.Pagination, sort api.Sortable) ([]*models.DaemonGcRunner, int64, error) {
	pagination = utils.NormalizePagination(pagination)
	q := s.tx.DaemonGcRunner.WithContext(ctx).Where(s.tx.DaemonGcRunner.RuleID.Eq(ruleID))
	field, ok := s.tx.DaemonGcRunner.GetFieldByName(ptr.To(sort.Sort))
	switch {
	case ok && ptr.To(sort.Method) == enums.SortMethodAsc:
		q = q.Order(field)
	case ok && ptr.To(sort.Method) == enums.SortMethodDesc:
		q = q.Order(field.Desc())
	default:
		q = q.Order(s.tx.DaemonGcRunner.UpdatedAt.Desc())
	}
	return q.FindByPage(ptr.To(pagination.Limit)*(ptr.To(pagination.Page)-1), ptr.To(pagination.Limit))
}
func (s *daemonRepository) CreateGcRunner(ctx context.Context, runner *models.DaemonGcRunner) error {
	return s.tx.DaemonGcRunner.WithContext(ctx).Create(runner)
}
func (s *daemonRepository) UpdateGcRunner(ctx context.Context, id string, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	_, err := s.tx.DaemonGcRunner.WithContext(ctx).Where(s.tx.DaemonGcRunner.ID.Eq(id)).Updates(updates)
	return err
}
func (s *daemonRepository) CreateGcRecords(ctx context.Context, records []*models.DaemonGcRecord) error {
	return s.tx.DaemonGcRecord.WithContext(ctx).CreateInBatches(records, consts.InsertBatchSize)
}
func (s *daemonRepository) ListGcRecords(ctx context.Context, runnerID string, pagination api.Pagination, sort api.Sortable) ([]*models.DaemonGcRecord, int64, error) {
	pagination = utils.NormalizePagination(pagination)
	q := s.tx.DaemonGcRecord.WithContext(ctx).Where(s.tx.DaemonGcRecord.RunnerID.Eq(runnerID))
	field, ok := s.tx.DaemonGcRecord.GetFieldByName(ptr.To(sort.Sort))
	switch {
	case ok && ptr.To(sort.Method) == enums.SortMethodAsc:
		q = q.Order(field)
	case ok && ptr.To(sort.Method) == enums.SortMethodDesc:
		q = q.Order(field.Desc())
	default:
		q = q.Order(s.tx.DaemonGcRecord.UpdatedAt.Desc())
	}
	return q.FindByPage(ptr.To(pagination.Limit)*(ptr.To(pagination.Page)-1), ptr.To(pagination.Limit))
}
func (s *daemonRepository) GetGcRecord(ctx context.Context, id string) (*models.DaemonGcRecord, error) {
	return s.tx.DaemonGcRecord.WithContext(ctx).Where(s.tx.DaemonGcRecord.ID.Eq(id)).Preload(s.tx.DaemonGcRecord.Runner).Preload(s.tx.DaemonGcRecord.Runner.Rule).First()
}
func (s *daemonRepository) UpsertGcStorageDeletionTask(ctx context.Context, task *models.GcStorageDeletionTask) error {
	task.StoragePathHash = storagePathHash(task.StoragePath)
	db := s.tx.UnderlyingDB().WithContext(ctx)
	if err := db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "resource_type"}, {Name: "resource_id"}, {Name: "storage_path_hash"}, {Name: "deleted_at"}}, DoUpdates: clause.AssignmentColumns([]string{"daemon", "runner_id", "status", "message", "updated_at"})}).Create(task).Error; err != nil {
		return err
	}
	if task.ID != "" {
		return nil
	}
	return db.Where("resource_type = ? AND resource_id = ? AND storage_path_hash = ? AND storage_path = ? AND deleted_at = 0", task.ResourceType, task.ResourceID, task.StoragePathHash, task.StoragePath).First(task).Error
}
func storagePathHash(path string) string {
	sum := sha256.Sum256([]byte(path))
	return hex.EncodeToString(sum[:])
}
func (s *daemonRepository) UpdateGcStorageDeletionTask(ctx context.Context, id string, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	matched := s.tx.UnderlyingDB().WithContext(ctx).Model(&models.GcStorageDeletionTask{}).Where("id = ? AND deleted_at = 0", id).Updates(updates)
	if matched.Error != nil {
		return matched.Error
	}
	if matched.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
