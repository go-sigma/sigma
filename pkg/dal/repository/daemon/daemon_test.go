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

package daemon_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	repodaemon "github.com/go-sigma/sigma/pkg/dal/repository/daemon"
	"github.com/go-sigma/sigma/pkg/testkit"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

func TestNewDaemonRepository(t *testing.T) {
	require.NotNil(t, repodaemon.NewDaemonRepository())
	require.NotNil(t, repodaemon.NewDaemonRepository(query.Q))
}

func TestDaemonRepositoryGcTagFlow(t *testing.T) {
	testkit.InitRepository(t)

	ctx := t.Context()
	daemonRepository := repodaemon.NewDaemonRepository()
	namespaceID := uuid.NewV7String()
	cronRule := "0 0 * * *"
	nextTrigger := int64(100)
	rule := &models.DaemonGcTagRule{
		ID:                  uuid.NewV7String(),
		NamespaceID:         &namespaceID,
		CronEnabled:         true,
		CronRule:            &cronRule,
		CronNextTrigger:     &nextTrigger,
		RetentionRuleType:   enums.RetentionRuleTypeDay,
		RetentionRuleAmount: 7,
	}
	require.NoError(t, daemonRepository.CreateGcTagRule(ctx, rule))

	gotRule, err := daemonRepository.GetGcTagRule(ctx, &namespaceID)
	require.NoError(t, err)
	require.Equal(t, rule.ID, gotRule.ID)

	require.NoError(t, daemonRepository.UpdateGcTagRule(ctx, rule.ID, map[string]any{
		query.DaemonGcTagRule.IsRunning.ColumnName().String():           true,
		query.DaemonGcTagRule.RetentionRuleAmount.ColumnName().String(): int64(14),
	}))
	gotRule, err = daemonRepository.GetGcTagRule(ctx, &namespaceID)
	require.NoError(t, err)
	require.True(t, gotRule.IsRunning)
	require.Equal(t, int64(14), gotRule.RetentionRuleAmount)

	oldRunner := &models.DaemonGcTagRunner{
		ID:          uuid.NewV7String(),
		CreatedAt:   10,
		RuleID:      rule.ID,
		Status:      enums.TaskCommonStatusPending,
		OperateType: enums.OperateTypeAutomatic,
	}
	runner := &models.DaemonGcTagRunner{
		ID:          uuid.NewV7String(),
		CreatedAt:   20,
		RuleID:      rule.ID,
		Status:      enums.TaskCommonStatusDoing,
		OperateType: enums.OperateTypeManual,
	}
	require.NoError(t, daemonRepository.CreateGcTagRunner(ctx, oldRunner))
	require.NoError(t, daemonRepository.CreateGcTagRunner(ctx, runner))

	latestRunner, err := daemonRepository.GetGcTagLatestRunner(ctx, rule.ID)
	require.NoError(t, err)
	require.Equal(t, runner.ID, latestRunner.ID)

	gotRunner, err := daemonRepository.GetGcTagRunner(ctx, runner.ID)
	require.NoError(t, err)
	require.Equal(t, rule.ID, gotRunner.Rule.ID)

	runners, total, err := daemonRepository.ListGcTagRunners(ctx, rule.ID, testPagination(), testSort("created_at"))
	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Equal(t, []string{oldRunner.ID, runner.ID}, []string{runners[0].ID, runners[1].ID})

	require.NoError(t, daemonRepository.UpdateGcTagRunner(ctx, runner.ID, map[string]any{
		query.DaemonGcTagRunner.Status.ColumnName().String(): enums.TaskCommonStatusSuccess,
	}))
	gotRunner, err = daemonRepository.GetGcTagRunner(ctx, runner.ID)
	require.NoError(t, err)
	require.Equal(t, enums.TaskCommonStatusSuccess, gotRunner.Status)

	recordOne := &models.DaemonGcTagRecord{
		ID:       uuid.NewV7String(),
		RunnerID: runner.ID,
		Tag:      "v1",
		Status:   enums.GcRecordStatusSuccess,
	}
	recordTwo := &models.DaemonGcTagRecord{
		ID:       uuid.NewV7String(),
		RunnerID: runner.ID,
		Tag:      "v2",
		Status:   enums.GcRecordStatusFailed,
	}
	require.NoError(t, daemonRepository.CreateGcTagRecords(ctx, []*models.DaemonGcTagRecord{recordOne, recordTwo}))

	records, total, err := daemonRepository.ListGcTagRecords(ctx, runner.ID, testPagination(), testSort("tag"))
	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Equal(t, []string{"v1", "v2"}, []string{records[0].Tag, records[1].Tag})

	gotRecord, err := daemonRepository.GetGcTagRecord(ctx, recordOne.ID)
	require.NoError(t, err)
	require.Equal(t, runner.ID, gotRecord.Runner.ID)
	require.Equal(t, rule.ID, gotRecord.Runner.Rule.ID)
}

func TestDaemonRepositoryGcRepositoryFlow(t *testing.T) {
	testkit.InitRepository(t)

	ctx := t.Context()
	daemonRepository := repodaemon.NewDaemonRepository()
	namespaceID := uuid.NewV7String()
	rule := &models.DaemonGcRepositoryRule{
		ID:           uuid.NewV7String(),
		NamespaceID:  &namespaceID,
		RetentionDay: 7,
	}
	require.NoError(t, daemonRepository.CreateGcRepositoryRule(ctx, rule))

	gotRule, err := daemonRepository.GetGcRepositoryRule(ctx, &namespaceID)
	require.NoError(t, err)
	require.Equal(t, rule.ID, gotRule.ID)

	require.NoError(t, daemonRepository.UpdateGcRepositoryRule(ctx, rule.ID, map[string]any{
		query.DaemonGcRepositoryRule.RetentionDay.ColumnName().String(): 14,
	}))
	gotRule, err = daemonRepository.GetGcRepositoryRule(ctx, &namespaceID)
	require.NoError(t, err)
	require.Equal(t, 14, gotRule.RetentionDay)

	runner := &models.DaemonGcRepositoryRunner{
		ID:          uuid.NewV7String(),
		RuleID:      rule.ID,
		Status:      enums.TaskCommonStatusDoing,
		OperateType: enums.OperateTypeManual,
	}
	require.NoError(t, daemonRepository.CreateGcRepositoryRunner(ctx, runner))

	gotRunner, err := daemonRepository.GetGcRepositoryLatestRunner(ctx, rule.ID)
	require.NoError(t, err)
	require.Equal(t, runner.ID, gotRunner.ID)

	gotRunner, err = daemonRepository.GetGcRepositoryRunner(ctx, runner.ID)
	require.NoError(t, err)
	require.Equal(t, rule.ID, gotRunner.Rule.ID)

	runners, total, err := daemonRepository.ListGcRepositoryRunners(ctx, rule.ID, testPagination(), testSort("created_at"))
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Equal(t, runner.ID, runners[0].ID)

	require.NoError(t, daemonRepository.UpdateGcRepositoryRunner(ctx, runner.ID, map[string]any{
		query.DaemonGcRepositoryRunner.Status.ColumnName().String(): enums.TaskCommonStatusSuccess,
	}))

	record := &models.DaemonGcRepositoryRecord{
		ID:         uuid.NewV7String(),
		RunnerID:   runner.ID,
		Repository: "library/alpine",
		Status:     enums.GcRecordStatusSuccess,
	}
	require.NoError(t, daemonRepository.CreateGcRepositoryRecords(ctx, []*models.DaemonGcRepositoryRecord{record}))

	records, total, err := daemonRepository.ListGcRepositoryRecords(ctx, runner.ID, testPagination(), testSort("repository"))
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Equal(t, "library/alpine", records[0].Repository)

	gotRecord, err := daemonRepository.GetGcRepositoryRecord(ctx, record.ID)
	require.NoError(t, err)
	require.Equal(t, rule.ID, gotRecord.Runner.Rule.ID)
}

func TestDaemonRepositoryGcArtifactFlow(t *testing.T) {
	testkit.InitRepository(t)

	ctx := t.Context()
	daemonRepository := repodaemon.NewDaemonRepository()
	namespaceID := uuid.NewV7String()
	rule := &models.DaemonGcArtifactRule{
		ID:           uuid.NewV7String(),
		NamespaceID:  &namespaceID,
		RetentionDay: 7,
	}
	require.NoError(t, daemonRepository.CreateGcArtifactRule(ctx, rule))

	gotRule, err := daemonRepository.GetGcArtifactRule(ctx, &namespaceID)
	require.NoError(t, err)
	require.Equal(t, rule.ID, gotRule.ID)

	require.NoError(t, daemonRepository.UpdateGcArtifactRule(ctx, rule.ID, map[string]any{
		query.DaemonGcArtifactRule.RetentionDay.ColumnName().String(): 14,
	}))
	gotRule, err = daemonRepository.GetGcArtifactRule(ctx, &namespaceID)
	require.NoError(t, err)
	require.Equal(t, 14, gotRule.RetentionDay)

	runner := &models.DaemonGcArtifactRunner{
		ID:          uuid.NewV7String(),
		RuleID:      rule.ID,
		Status:      enums.TaskCommonStatusDoing,
		OperateType: enums.OperateTypeManual,
	}
	require.NoError(t, daemonRepository.CreateGcArtifactRunner(ctx, runner))

	gotRunner, err := daemonRepository.GetGcArtifactLatestRunner(ctx, rule.ID)
	require.NoError(t, err)
	require.Equal(t, runner.ID, gotRunner.ID)

	gotRunner, err = daemonRepository.GetGcArtifactRunner(ctx, runner.ID)
	require.NoError(t, err)
	require.Equal(t, rule.ID, gotRunner.Rule.ID)

	runners, total, err := daemonRepository.ListGcArtifactRunners(ctx, rule.ID, testPagination(), testSort("created_at"))
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Equal(t, runner.ID, runners[0].ID)

	require.NoError(t, daemonRepository.UpdateGcArtifactRunner(ctx, runner.ID, map[string]any{
		query.DaemonGcArtifactRunner.Status.ColumnName().String(): enums.TaskCommonStatusSuccess,
	}))

	record := &models.DaemonGcArtifactRecord{
		ID:       uuid.NewV7String(),
		RunnerID: runner.ID,
		Digest:   "sha256:artifact",
		Status:   enums.GcRecordStatusSuccess,
	}
	require.NoError(t, daemonRepository.CreateGcArtifactRecords(ctx, []*models.DaemonGcArtifactRecord{record}))

	records, total, err := daemonRepository.ListGcArtifactRecords(ctx, runner.ID, testPagination(), testSort("digest"))
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Equal(t, "sha256:artifact", records[0].Digest)

	gotRecord, err := daemonRepository.GetGcArtifactRecord(ctx, record.ID)
	require.NoError(t, err)
	require.Equal(t, rule.ID, gotRecord.Runner.Rule.ID)
}

func TestDaemonRepositoryGcBlobFlow(t *testing.T) {
	testkit.InitRepository(t)

	ctx := t.Context()
	daemonRepository := repodaemon.NewDaemonRepository()
	rule := &models.DaemonGcBlobRule{
		ID:           uuid.NewV7String(),
		RetentionDay: 7,
	}
	require.NoError(t, daemonRepository.CreateGcBlobRule(ctx, rule))

	gotRule, err := daemonRepository.GetGcBlobRule(ctx)
	require.NoError(t, err)
	require.Equal(t, rule.ID, gotRule.ID)

	require.NoError(t, daemonRepository.UpdateGcBlobRule(ctx, rule.ID, map[string]any{
		query.DaemonGcBlobRule.RetentionDay.ColumnName().String(): 14,
	}))
	gotRule, err = daemonRepository.GetGcBlobRule(ctx)
	require.NoError(t, err)
	require.Equal(t, 14, gotRule.RetentionDay)

	runner := &models.DaemonGcBlobRunner{
		ID:          uuid.NewV7String(),
		RuleID:      rule.ID,
		Status:      enums.TaskCommonStatusDoing,
		OperateType: enums.OperateTypeManual,
	}
	require.NoError(t, daemonRepository.CreateGcBlobRunner(ctx, runner))

	gotRunner, err := daemonRepository.GetGcBlobLatestRunner(ctx, rule.ID)
	require.NoError(t, err)
	require.Equal(t, runner.ID, gotRunner.ID)

	gotRunner, err = daemonRepository.GetGcBlobRunner(ctx, runner.ID)
	require.NoError(t, err)
	require.Equal(t, rule.ID, gotRunner.Rule.ID)

	runners, total, err := daemonRepository.ListGcBlobRunners(ctx, rule.ID, testPagination(), testSort("created_at"))
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Equal(t, runner.ID, runners[0].ID)

	require.NoError(t, daemonRepository.UpdateGcBlobRunner(ctx, runner.ID, map[string]any{
		query.DaemonGcBlobRunner.Status.ColumnName().String(): enums.TaskCommonStatusSuccess,
	}))

	record := &models.DaemonGcBlobRecord{
		ID:       uuid.NewV7String(),
		RunnerID: runner.ID,
		Digest:   "sha256:blob",
		Status:   enums.GcRecordStatusSuccess,
	}
	require.NoError(t, daemonRepository.CreateGcBlobRecords(ctx, []*models.DaemonGcBlobRecord{record}))

	records, total, err := daemonRepository.ListGcBlobRecords(ctx, runner.ID, testPagination(), testSort("digest"))
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Equal(t, "sha256:blob", records[0].Digest)

	gotRecord, err := daemonRepository.GetGcBlobRecord(ctx, record.ID)
	require.NoError(t, err)
	require.Equal(t, rule.ID, gotRecord.Runner.Rule.ID)
}

func TestDaemonRepositoryStorageDeletionTask(t *testing.T) {
	testkit.InitRepository(t)

	ctx := t.Context()
	daemonRepository := repodaemon.NewDaemonRepository()
	db := query.Q.UnderlyingDB().WithContext(ctx)
	task := &models.GcStorageDeletionTask{
		ID:           uuid.NewV7String(),
		Daemon:       enums.DaemonGcBlob,
		RunnerID:     uuid.NewV7String(),
		ResourceType: "blob",
		ResourceID:   "sha256:blob",
		StoragePath:  "/data/blobs/sha256/blob",
		Status:       enums.TaskCommonStatusPending,
	}
	require.NoError(t, daemonRepository.UpsertGcStorageDeletionTask(ctx, task))
	require.NotEmpty(t, task.StoragePathHash)

	var stored models.GcStorageDeletionTask
	require.NoError(t, db.Where("id = ?", task.ID).First(&stored).Error)
	require.Equal(t, enums.TaskCommonStatusPending, stored.Status)

	refreshed := &models.GcStorageDeletionTask{
		ID:           uuid.NewV7String(),
		Daemon:       enums.DaemonGcBlob,
		RunnerID:     uuid.NewV7String(),
		ResourceType: task.ResourceType,
		ResourceID:   task.ResourceID,
		StoragePath:  task.StoragePath,
		Status:       enums.TaskCommonStatusDoing,
		Message:      []byte("retry"),
	}
	require.NoError(t, daemonRepository.UpsertGcStorageDeletionTask(ctx, refreshed))

	var count int64
	require.NoError(t, db.Model(&models.GcStorageDeletionTask{}).Where(
		"resource_type = ? AND resource_id = ? AND storage_path_hash = ? AND deleted_at = 0",
		task.ResourceType, task.ResourceID, task.StoragePathHash,
	).Count(&count).Error)
	require.Equal(t, int64(1), count)

	require.NoError(t, db.Where("id = ?", task.ID).First(&stored).Error)
	require.Equal(t, enums.TaskCommonStatusDoing, stored.Status)
	require.Equal(t, []byte("retry"), stored.Message)

	require.NoError(t, daemonRepository.UpdateGcStorageDeletionTask(ctx, task.ID, map[string]any{
		"status":   enums.TaskCommonStatusSuccess,
		"attempts": 1,
		"message":  []byte("deleted"),
	}))
	require.NoError(t, db.Where("id = ?", task.ID).First(&stored).Error)
	require.Equal(t, enums.TaskCommonStatusSuccess, stored.Status)
	require.Equal(t, 1, stored.Attempts)
	require.Equal(t, []byte("deleted"), stored.Message)
}

func testPagination() api.Pagination {
	page := 1
	limit := 10
	return api.Pagination{Page: &page, Limit: &limit}
}

func testSort(column string) api.Sortable {
	method := enums.SortMethodAsc
	return api.Sortable{Sort: &column, Method: &method}
}
