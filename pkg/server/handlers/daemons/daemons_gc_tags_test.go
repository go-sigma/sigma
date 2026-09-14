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

func TestUpdateGcTagRule(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		UpdateGcTagRule(gomock.Any(), gomock.Any()).
		Return(nil)

	recorder, c := newDaemonsContext(t)
	(&handler{DaemonSvc: serviceObj}).UpdateGcTagRule(c, &api.UpdateGcTagRuleRequest{NamespaceID: "1", CronEnabled: true, RetentionRuleType: enums.RetentionRuleTypeDay, RetentionRuleAmount: 30})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusNoContent, recorder.Code)
}

func TestUpdateGcTagRuleReturnsInternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		UpdateGcTagRule(gomock.Any(), gomock.Any()).
		Return(errors.New("update failed"))

	recorder, c := newDaemonsContext(t)
	(&handler{DaemonSvc: serviceObj}).UpdateGcTagRule(c, &api.UpdateGcTagRuleRequest{NamespaceID: "1"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestGetGcTagRule(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		GetGcTagRule(gomock.Any(), "1").
		Return(&models.DaemonGcRule{
			ID:                  "1",
			IsRunning:           false,
			CronEnabled:         true,
			RetentionRuleType:   enums.RetentionRuleTypeDay,
			RetentionRuleAmount: 30,
			CronRule:            new("0 0 * * 6"),
			CronNextTrigger:     nil,
			CreatedAt:           1,
			UpdatedAt:           1,
		}, nil)

	recorder, c := newDaemonsContext(t)
	(&handler{DaemonSvc: serviceObj}).GetGcTagRule(c, &api.GetGcTagRuleRequest{NamespaceID: "1"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"cron_enabled":true`)
	require.Contains(t, recorder.Body.String(), `"retention_rule_amount":30`)
}

func TestGetGcTagRuleReturnsInternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		GetGcTagRule(gomock.Any(), "1").
		Return(nil, errors.New("get failed"))

	recorder, c := newDaemonsContext(t)
	(&handler{DaemonSvc: serviceObj}).GetGcTagRule(c, &api.GetGcTagRuleRequest{NamespaceID: "1"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestGetGcTagLatestRunner(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		GetGcTagLatestRunner(gomock.Any(), "1").
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
	(&handler{DaemonSvc: serviceObj}).GetGcTagLatestRunner(c, &api.GetGcTagLatestRunnerRequest{NamespaceID: "1"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"id":"2"`)
	require.Contains(t, recorder.Body.String(), `"status":"Pending"`)
}

func TestGetGcTagLatestRunnerReturnsInternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		GetGcTagLatestRunner(gomock.Any(), "1").
		Return(nil, errors.New("get failed"))

	recorder, c := newDaemonsContext(t)
	(&handler{DaemonSvc: serviceObj}).GetGcTagLatestRunner(c, &api.GetGcTagLatestRunnerRequest{NamespaceID: "1"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestCreateGcTagRunner(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		CreateGcTagRunner(gomock.Any(), gomock.Any()).
		Return(nil)

	recorder, c := newDaemonsContext(t)
	(&handler{DaemonSvc: serviceObj}).CreateGcTagRunner(c, &api.CreateGcTagRunnerRequest{NamespaceID: "1"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusCreated, recorder.Code)
}

func TestCreateGcTagRunnerReturnsInternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		CreateGcTagRunner(gomock.Any(), gomock.Any()).
		Return(errors.New("create failed"))

	recorder, c := newDaemonsContext(t)
	(&handler{DaemonSvc: serviceObj}).CreateGcTagRunner(c, &api.CreateGcTagRunnerRequest{NamespaceID: "1"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestListGcTagRunners(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		ListGcTagRunners(gomock.Any(), "1", gomock.Any(), gomock.Any()).
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
	(&handler{DaemonSvc: serviceObj}).ListGcTagRunners(c, &api.ListGcTagRunnersRequest{NamespaceID: "1"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"total":1`)
	require.Contains(t, recorder.Body.String(), `"id":"2"`)
}

func TestListGcTagRunnersReturnsInternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		ListGcTagRunners(gomock.Any(), "1", gomock.Any(), gomock.Any()).
		Return(nil, int64(0), errors.New("list failed"))

	recorder, c := newDaemonsContext(t)
	(&handler{DaemonSvc: serviceObj}).ListGcTagRunners(c, &api.ListGcTagRunnersRequest{NamespaceID: "1"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestGetGcTagRunner(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		GetGcTagRunner(gomock.Any(), "1", "2").
		Return(&models.DaemonGcRunner{
			ID:        "2",
			Status:    enums.TaskCommonStatusSuccess,
			Message:   []byte("log"),
			CreatedAt: 1,
			UpdatedAt: 1,
		}, nil)

	recorder, c := newDaemonsContext(t)
	(&handler{DaemonSvc: serviceObj}).GetGcTagRunner(c, &api.GetGcTagRunnerRequest{NamespaceID: "1", RunnerID: "2"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"id":"2"`)
}

func TestGetGcTagRunnerReturnsInternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		GetGcTagRunner(gomock.Any(), "1", "2").
		Return(nil, errors.New("get failed"))

	recorder, c := newDaemonsContext(t)
	(&handler{DaemonSvc: serviceObj}).GetGcTagRunner(c, &api.GetGcTagRunnerRequest{NamespaceID: "1", RunnerID: "2"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestListGcTagRecords(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		ListGcTagRecords(gomock.Any(), "2", gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, runnerID string, pagination api.Pagination, sort api.Sortable) ([]*models.DaemonGcRecord, int64, error) {
			require.Equal(t, "2", runnerID)
			return []*models.DaemonGcRecord{{
				ID:        "3",
				RunnerID:  "2",
				Resource:  "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				Status:    enums.GcRecordStatusSuccess,
				Message:   []byte("log"),
				CreatedAt: 1,
				UpdatedAt: 1,
			}}, 1, nil
		})

	recorder, c := newDaemonsContext(t)
	(&handler{DaemonSvc: serviceObj}).ListGcTagRecords(c, &api.ListGcTagRecordsRequest{RunnerID: "2"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"total":1`)
	require.Contains(t, recorder.Body.String(), `"digest":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"`)
}

func TestListGcTagRecordsReturnsInternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		ListGcTagRecords(gomock.Any(), "2", gomock.Any(), gomock.Any()).
		Return(nil, int64(0), errors.New("list failed"))

	recorder, c := newDaemonsContext(t)
	(&handler{DaemonSvc: serviceObj}).ListGcTagRecords(c, &api.ListGcTagRecordsRequest{RunnerID: "2"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestGetGcTagRecord(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		GetGcTagRecord(gomock.Any(), "1", "2", "3").
		Return(&models.DaemonGcRecord{
			ID:        "3",
			RunnerID:  "2",
			Resource:  "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			Status:    enums.GcRecordStatusSuccess,
			Message:   []byte("log"),
			CreatedAt: 1,
			UpdatedAt: 1,
		}, nil)

	recorder, c := newDaemonsContext(t)
	(&handler{DaemonSvc: serviceObj}).GetGcTagRecord(c, &api.GetGcTagRecordRequest{NamespaceID: "1", RunnerID: "2", RecordID: "3"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"id":"3"`)
}

func TestGetGcTagRecordReturnsInternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		GetGcTagRecord(gomock.Any(), "1", "2", "3").
		Return(nil, errors.New("get failed"))

	recorder, c := newDaemonsContext(t)
	(&handler{DaemonSvc: serviceObj}).GetGcTagRecord(c, &api.GetGcTagRecordRequest{NamespaceID: "1", RunnerID: "2", RecordID: "3"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}
