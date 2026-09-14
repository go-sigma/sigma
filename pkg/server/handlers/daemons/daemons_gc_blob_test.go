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
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	svcdaemons "github.com/go-sigma/sigma/pkg/service/daemons"
)

func TestUpdateGcBlobRule(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		UpdateGcBlobRule(gomock.Any(), gomock.Any()).
		Return(nil)

	recorder, c := newDaemonsContext(t)
	(&handler{DaemonSvc: serviceObj}).UpdateGcBlobRule(c, &api.UpdateGcBlobRuleRequest{NamespaceID: "1", RetentionDay: 7})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusNoContent, recorder.Code)
}

func TestUpdateGcBlobRuleReturnsInternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		UpdateGcBlobRule(gomock.Any(), gomock.Any()).
		Return(errors.New("update failed"))

	recorder, c := newDaemonsContext(t)
	(&handler{DaemonSvc: serviceObj}).UpdateGcBlobRule(c, &api.UpdateGcBlobRuleRequest{NamespaceID: "1"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestGetGcBlobRule(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		GetGcBlobRule(gomock.Any()).
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
	(&handler{DaemonSvc: serviceObj}).GetGcBlobRule(c, &api.GetGcBlobRuleRequest{})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"retention_day":7`)
	require.Contains(t, recorder.Body.String(), `"cron_enabled":true`)
}

func TestGetGcBlobRuleReturnsInternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		GetGcBlobRule(gomock.Any()).
		Return(nil, errors.New("get failed"))

	recorder, c := newDaemonsContext(t)
	(&handler{DaemonSvc: serviceObj}).GetGcBlobRule(c, &api.GetGcBlobRuleRequest{})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestGetGcBlobLatestRunner(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		GetGcBlobLatestRunner(gomock.Any(), "1").
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
	(&handler{DaemonSvc: serviceObj}).GetGcBlobLatestRunner(c, &api.GetGcBlobLatestRunnerRequest{NamespaceID: "1"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"id":"2"`)
	require.Contains(t, recorder.Body.String(), `"status":"Pending"`)
}

func TestGetGcBlobLatestRunnerReturnsInternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		GetGcBlobLatestRunner(gomock.Any(), "1").
		Return(nil, errors.New("get failed"))

	recorder, c := newDaemonsContext(t)
	(&handler{DaemonSvc: serviceObj}).GetGcBlobLatestRunner(c, &api.GetGcBlobLatestRunnerRequest{NamespaceID: "1"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestCreateGcBlobRunner(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		CreateGcBlobRunner(gomock.Any(), "10", gomock.Any()).
		Return(nil)

	recorder, c := newDaemonsContext(t)
	(&handler{DaemonSvc: serviceObj}).CreateGcBlobRunner(c, &api.CreateGcBlobRunnerRequest{NamespaceID: "1"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusCreated, recorder.Code)
}

func TestCreateGcBlobRunnerReturnsInternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		CreateGcBlobRunner(gomock.Any(), "10", gomock.Any()).
		Return(errors.New("create failed"))

	recorder, c := newDaemonsContext(t)
	(&handler{DaemonSvc: serviceObj}).CreateGcBlobRunner(c, &api.CreateGcBlobRunnerRequest{NamespaceID: "1"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestListGcBlobRunners(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		ListGcBlobRunners(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, pagination api.Pagination, sort api.Sortable) ([]*models.DaemonGcRunner, int64, error) {
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
	(&handler{DaemonSvc: serviceObj}).ListGcBlobRunners(c, &api.ListGcBlobRunnersRequest{})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"total":1`)
	require.Contains(t, recorder.Body.String(), `"id":"2"`)
}

func TestListGcBlobRunnersReturnsInternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		ListGcBlobRunners(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, int64(0), errors.New("list failed"))

	recorder, c := newDaemonsContext(t)
	(&handler{DaemonSvc: serviceObj}).ListGcBlobRunners(c, &api.ListGcBlobRunnersRequest{})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestGetGcBlobRunner(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		GetGcBlobRunner(gomock.Any(), "2").
		Return(&models.DaemonGcRunner{
			ID:        "2",
			Status:    enums.TaskCommonStatusSuccess,
			Message:   []byte("log"),
			CreatedAt: 1,
			UpdatedAt: 1,
		}, nil)

	recorder, c := newDaemonsContext(t)
	(&handler{DaemonSvc: serviceObj}).GetGcBlobRunner(c, &api.GetGcBlobRunnerRequest{RunnerID: "2"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"id":"2"`)
}

func TestGetGcBlobRunnerReturnsInternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		GetGcBlobRunner(gomock.Any(), "2").
		Return(nil, errors.New("get failed"))

	recorder, c := newDaemonsContext(t)
	(&handler{DaemonSvc: serviceObj}).GetGcBlobRunner(c, &api.GetGcBlobRunnerRequest{RunnerID: "2"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestListGcBlobRecords(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		ListGcBlobRecords(gomock.Any(), "2", gomock.Any(), gomock.Any()).
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
	(&handler{DaemonSvc: serviceObj}).ListGcBlobRecords(c, &api.ListGcBlobRecordsRequest{RunnerID: "2"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"total":1`)
	require.Contains(t, recorder.Body.String(), `"digest":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"`)
}

func TestListGcBlobRecordsReturnsInternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		ListGcBlobRecords(gomock.Any(), "2", gomock.Any(), gomock.Any()).
		Return(nil, int64(0), errors.New("list failed"))

	recorder, c := newDaemonsContext(t)
	(&handler{DaemonSvc: serviceObj}).ListGcBlobRecords(c, &api.ListGcBlobRecordsRequest{RunnerID: "2"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestGetGcBlobRecord(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		GetGcBlobRecord(gomock.Any(), "2", "3").
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
	(&handler{DaemonSvc: serviceObj}).GetGcBlobRecord(c, &api.GetGcBlobRecordRequest{RunnerID: "2", RecordID: "3"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"id":"3"`)
}

func TestGetGcBlobRecordReturnsInternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcdaemons.NewMockDaemonService(ctrl)
	serviceObj.EXPECT().
		GetGcBlobRecord(gomock.Any(), "2", "3").
		Return(nil, errors.New("get failed"))

	recorder, c := newDaemonsContext(t)
	(&handler{DaemonSvc: serviceObj}).GetGcBlobRecord(c, &api.GetGcBlobRecordRequest{RunnerID: "2", RecordID: "3"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}
func newDaemonsContext(t *testing.T) (*httptest.ResponseRecorder, *gin.Context) {
	t.Helper()
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Set(consts.ContextUser, &models.User{ID: "10", Role: enums.UserRoleAdmin})
	return recorder, c
}
