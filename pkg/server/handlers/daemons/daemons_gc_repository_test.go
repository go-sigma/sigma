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

package daemons

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/dal/models"
	svcdaemons "github.com/go-sigma/sigma/pkg/service/daemons"
)

func TestUpdateGcRepositoryRule(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		UpdateGcRepositoryRule(gomock.Any(), gomock.Any()).
		Return(nil)

	recorder, c := newDaemonsContext(t)
	(&handler{DaemonSvc: serviceObj}).UpdateGcRepositoryRule(c, &api.UpdateGcRepositoryRuleRequest{NamespaceID: "1", RetentionDay: 7})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusNoContent, recorder.Code)
}

func TestUpdateGcRepositoryRuleReturnsInternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		UpdateGcRepositoryRule(gomock.Any(), gomock.Any()).
		Return(errors.New("update failed"))

	recorder, c := newDaemonsContext(t)
	(&handler{DaemonSvc: serviceObj}).UpdateGcRepositoryRule(c, &api.UpdateGcRepositoryRuleRequest{NamespaceID: "1"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestGetGcRepositoryRule(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		GetGcRepositoryRule(gomock.Any(), "1").
		Return(&models.DaemonGcRule{
			ID:              "1",
			IsRunning:       false,
			RetentionDay:    7,
			CronEnabled:     true,
			CronRule:        new("0 0 * * 6"),
			CronNextTrigger: nil,
			CreatedAt:       1,
			UpdatedAt:       1,
		}, nil)

	recorder, c := newDaemonsContext(t)
	(&handler{DaemonSvc: serviceObj}).GetGcRepositoryRule(c, &api.GetGcRepositoryRuleRequest{NamespaceID: "1"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"cron_enabled":true`)
	require.Contains(t, recorder.Body.String(), `"retention_day":7`)
}

func TestGetGcRepositoryRuleReturnsInternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		GetGcRepositoryRule(gomock.Any(), "1").
		Return(nil, errors.New("get failed"))

	recorder, c := newDaemonsContext(t)
	(&handler{DaemonSvc: serviceObj}).GetGcRepositoryRule(c, &api.GetGcRepositoryRuleRequest{NamespaceID: "1"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestGetGcRepositoryLatestRunner(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		GetGcRepositoryLatestRunner(gomock.Any(), "1").
		Return(&models.DaemonGcRunner{
			ID:           "2",
			Status:       enums.TaskCommonStatusPending,
			Message:      []byte("log"),
			SuccessCount: new(int64(10)),
			FailedCount:  new(int64(1)),
			StartedAt:    new(int64(1)),
			EndedAt:      new(int64(2)),
			Duration:     new(int64(1000)),
			CreatedAt:    1,
			UpdatedAt:    1,
		}, nil)

	recorder, c := newDaemonsContext(t)
	(&handler{DaemonSvc: serviceObj}).GetGcRepositoryLatestRunner(c, &api.GetGcRepositoryLatestRunnerRequest{NamespaceID: "1"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"id":"2"`)
	require.Contains(t, recorder.Body.String(), `"status":"Pending"`)
}

func TestGetGcRepositoryLatestRunnerReturnsInternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		GetGcRepositoryLatestRunner(gomock.Any(), "1").
		Return(nil, errors.New("get failed"))

	recorder, c := newDaemonsContext(t)
	(&handler{DaemonSvc: serviceObj}).GetGcRepositoryLatestRunner(c, &api.GetGcRepositoryLatestRunnerRequest{NamespaceID: "1"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestCreateGcRepositoryRunner(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		CreateGcRepositoryRunner(gomock.Any(), gomock.Any()).
		Return(nil)

	recorder, c := newDaemonsContext(t)
	(&handler{DaemonSvc: serviceObj}).CreateGcRepositoryRunner(c, &api.CreateGcRepositoryRunnerRequest{NamespaceID: "1"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusCreated, recorder.Code)
}

func TestCreateGcRepositoryRunnerReturnsInternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		CreateGcRepositoryRunner(gomock.Any(), gomock.Any()).
		Return(errors.New("create failed"))

	recorder, c := newDaemonsContext(t)
	(&handler{DaemonSvc: serviceObj}).CreateGcRepositoryRunner(c, &api.CreateGcRepositoryRunnerRequest{NamespaceID: "1"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestListGcRepositoryRunners(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		ListGcRepositoryRunners(gomock.Any(), "1", gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, namespaceID string, pagination api.Pagination, sort api.Sortable) ([]*models.DaemonGcRunner, int64, error) {
			require.Equal(t, "1", namespaceID)
			return []*models.DaemonGcRunner{{
				ID:           "2",
				Status:       enums.TaskCommonStatusSuccess,
				Message:      []byte("log"),
				SuccessCount: new(int64(10)),
				FailedCount:  new(int64(1)),
				CreatedAt:    1,
				UpdatedAt:    1,
			}}, 1, nil
		})

	recorder, c := newDaemonsContext(t)
	(&handler{DaemonSvc: serviceObj}).ListGcRepositoryRunners(c, &api.ListGcRepositoryRunnersRequest{NamespaceID: "1"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"total":1`)
	require.Contains(t, recorder.Body.String(), `"id":"2"`)
}

func TestListGcRepositoryRunnersReturnsInternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		ListGcRepositoryRunners(gomock.Any(), "1", gomock.Any(), gomock.Any()).
		Return(nil, int64(0), errors.New("list failed"))

	recorder, c := newDaemonsContext(t)
	(&handler{DaemonSvc: serviceObj}).ListGcRepositoryRunners(c, &api.ListGcRepositoryRunnersRequest{NamespaceID: "1"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestGetGcRepositoryRunner(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		GetGcRepositoryRunner(gomock.Any(), "1", "2").
		Return(&models.DaemonGcRunner{
			ID:        "2",
			Status:    enums.TaskCommonStatusSuccess,
			Message:   []byte("log"),
			CreatedAt: 1,
			UpdatedAt: 1,
		}, nil)

	recorder, c := newDaemonsContext(t)
	(&handler{DaemonSvc: serviceObj}).GetGcRepositoryRunner(c, &api.GetGcRepositoryRunnerRequest{NamespaceID: "1", RunnerID: "2"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"id":"2"`)
}

func TestGetGcRepositoryRunnerReturnsInternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		GetGcRepositoryRunner(gomock.Any(), "1", "2").
		Return(nil, errors.New("get failed"))

	recorder, c := newDaemonsContext(t)
	(&handler{DaemonSvc: serviceObj}).GetGcRepositoryRunner(c, &api.GetGcRepositoryRunnerRequest{NamespaceID: "1", RunnerID: "2"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestListGcRepositoryRecords(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		ListGcRepositoryRecords(gomock.Any(), "2", gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, runnerID string, pagination api.Pagination, sort api.Sortable) ([]*models.DaemonGcRecord, int64, error) {
			require.Equal(t, "2", runnerID)
			return []*models.DaemonGcRecord{{
				ID:        "3",
				RunnerID:  "2",
				Resource:  "library/busybox",
				Status:    enums.GcRecordStatusSuccess,
				Message:   []byte("log"),
				CreatedAt: 1,
				UpdatedAt: 1,
			}}, 1, nil
		})

	recorder, c := newDaemonsContext(t)
	(&handler{DaemonSvc: serviceObj}).ListGcRepositoryRecords(c, &api.ListGcRepositoryRecordsRequest{RunnerID: "2"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"total":1`)
	require.Contains(t, recorder.Body.String(), `"repository":"library/busybox"`)
}

func TestListGcRepositoryRecordsReturnsInternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		ListGcRepositoryRecords(gomock.Any(), "2", gomock.Any(), gomock.Any()).
		Return(nil, int64(0), errors.New("list failed"))

	recorder, c := newDaemonsContext(t)
	(&handler{DaemonSvc: serviceObj}).ListGcRepositoryRecords(c, &api.ListGcRepositoryRecordsRequest{RunnerID: "2"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestGetGcRepositoryRecord(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		GetGcRepositoryRecord(gomock.Any(), "1", "2", "3").
		Return(&models.DaemonGcRecord{
			ID:        "3",
			RunnerID:  "2",
			Resource:  "library/busybox",
			Status:    enums.GcRecordStatusSuccess,
			Message:   []byte("log"),
			CreatedAt: 1,
			UpdatedAt: 1,
		}, nil)

	recorder, c := newDaemonsContext(t)
	(&handler{DaemonSvc: serviceObj}).GetGcRepositoryRecord(c, &api.GetGcRepositoryRecordRequest{NamespaceID: "1", RunnerID: "2", RecordID: "3"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"id":"3"`)
}

func TestGetGcRepositoryRecordReturnsInternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		GetGcRepositoryRecord(gomock.Any(), "1", "2", "3").
		Return(nil, errors.New("get failed"))

	recorder, c := newDaemonsContext(t)
	(&handler{DaemonSvc: serviceObj}).GetGcRepositoryRecord(c, &api.GetGcRepositoryRecordRequest{NamespaceID: "1", RunnerID: "2", RecordID: "3"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}
