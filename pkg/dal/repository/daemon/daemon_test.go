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

func TestDaemonRepositoryGcFlows(t *testing.T) {
	testkit.InitRepository(t)

	ctx := t.Context()
	daemonRepository := repodaemon.NewDaemonRepository()
	namespaceID := uuid.NewV7String()
	cronRule := "0 0 * * *"
	nextTrigger := int64(100)
	tests := []struct {
		name        string
		daemon      enums.Daemon
		namespaceID *string
		resource    string
	}{
		{name: "tag", daemon: enums.DaemonGcTag, namespaceID: &namespaceID, resource: "v1"},
		{name: "repository", daemon: enums.DaemonGcRepository, namespaceID: &namespaceID, resource: "library/alpine"},
		{name: "artifact", daemon: enums.DaemonGcArtifact, namespaceID: &namespaceID, resource: "sha256:artifact"},
		{name: "blob", daemon: enums.DaemonGcBlob, resource: "sha256:blob"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := &models.DaemonGcRule{
				ID:                  uuid.NewV7String(),
				Type:                tt.daemon,
				NamespaceID:         tt.namespaceID,
				CronEnabled:         true,
				CronRule:            &cronRule,
				CronNextTrigger:     &nextTrigger,
				RetentionDay:        7,
				RetentionRuleType:   enums.RetentionRuleTypeDay,
				RetentionRuleAmount: 7,
			}
			require.NoError(t, daemonRepository.CreateGcRule(ctx, rule))
			gotRule, err := daemonRepository.GetGcRule(ctx, tt.daemon, tt.namespaceID)
			require.NoError(t, err)
			require.Equal(t, rule.ID, gotRule.ID)
			require.NoError(t, daemonRepository.UpdateGcRule(ctx, rule.ID, map[string]any{
				query.DaemonGcRule.IsRunning.ColumnName().String(): true,
			}))

			runner := &models.DaemonGcRunner{ID: uuid.NewV7String(), RuleID: rule.ID, Status: enums.TaskCommonStatusDoing, OperateType: enums.OperateTypeManual}
			require.NoError(t, daemonRepository.CreateGcRunner(ctx, runner))
			gotRunner, err := daemonRepository.GetGcLatestRunner(ctx, rule.ID)
			require.NoError(t, err)
			require.Equal(t, runner.ID, gotRunner.ID)
			gotRunner, err = daemonRepository.GetGcRunner(ctx, runner.ID)
			require.NoError(t, err)
			require.Equal(t, tt.daemon, gotRunner.Rule.Type)

			record := &models.DaemonGcRecord{ID: uuid.NewV7String(), RunnerID: runner.ID, Resource: tt.resource, Status: enums.GcRecordStatusSuccess}
			require.NoError(t, daemonRepository.CreateGcRecords(ctx, []*models.DaemonGcRecord{record}))
			records, total, err := daemonRepository.ListGcRecords(ctx, runner.ID, testPagination(), testSort("resource"))
			require.NoError(t, err)
			require.Equal(t, int64(1), total)
			require.Equal(t, tt.resource, records[0].Resource)
			gotRecord, err := daemonRepository.GetGcRecord(ctx, record.ID)
			require.NoError(t, err)
			require.Equal(t, rule.ID, gotRecord.Runner.Rule.ID)
		})
	}
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
