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

package builders

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
	"github.com/go-sigma/sigma/pkg/server/errcode"
	svcbuilder "github.com/go-sigma/sigma/pkg/service/builders"
)

func TestCreateBuilder(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcbuilder.NewMockBuilderService(ctrl)
	serviceObj.EXPECT().
		CreateBuilder(gomock.Any(), "10", gomock.Any()).
		Return(nil)

	recorder, c := newBuildersContext(t)
	(&handler{BuilderSvc: serviceObj}).CreateBuilder(c, &api.CreateBuilderRequest{})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusCreated, recorder.Code)
}

func TestCreateBuilderReturnsInternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcbuilder.NewMockBuilderService(ctrl)
	serviceObj.EXPECT().
		CreateBuilder(gomock.Any(), "10", gomock.Any()).
		Return(errors.New("create failed"))

	recorder, c := newBuildersContext(t)
	(&handler{BuilderSvc: serviceObj}).CreateBuilder(c, &api.CreateBuilderRequest{})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestCreateBuilderReturnsErrCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcbuilder.NewMockBuilderService(ctrl)
	serviceObj.EXPECT().
		CreateBuilder(gomock.Any(), "10", gomock.Any()).
		Return(errcode.HTTPErrCodeBadRequest.Detail("parameter 'dockerfile' is invalid"))

	recorder, c := newBuildersContext(t)
	(&handler{BuilderSvc: serviceObj}).CreateBuilder(c, &api.CreateBuilderRequest{})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestUpdateBuilder(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcbuilder.NewMockBuilderService(ctrl)
	serviceObj.EXPECT().
		UpdateBuilder(gomock.Any(), "10", "1", gomock.Any()).
		Return(nil)

	recorder, c := newBuildersContext(t)
	(&handler{BuilderSvc: serviceObj}).UpdateBuilder(c, &api.UpdateBuilderRequest{BuilderID: "1"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusNoContent, recorder.Code)
}

func TestUpdateBuilderReturnsInternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcbuilder.NewMockBuilderService(ctrl)
	serviceObj.EXPECT().
		UpdateBuilder(gomock.Any(), "10", "1", gomock.Any()).
		Return(errors.New("update failed"))

	recorder, c := newBuildersContext(t)
	(&handler{BuilderSvc: serviceObj}).UpdateBuilder(c, &api.UpdateBuilderRequest{BuilderID: "1"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestListRunners(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcbuilder.NewMockBuilderService(ctrl)
	serviceObj.EXPECT().
		ListRunners(gomock.Any(), "1", gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, builderID string, pagination api.Pagination, sort api.Sortable) ([]*models.BuilderRunner, int64, error) {
			require.Equal(t, "1", builderID)
			require.NotNil(t, pagination.Page)
			require.NotNil(t, pagination.Limit)
			return []*models.BuilderRunner{{
				ID:            "2",
				BuilderID:     "1",
				Log:           []byte("log"),
				Status:        enums.BuildStatusSuccess,
				StatusMessage: nil,
				Tag:           new("v1.0"),
				RawTag:        "v1.0",
				Description:   nil,
				ScmBranch:     new("main"),
				StartedAt:     new(int64(1)),
				EndedAt:       new(int64(2)),
				Duration:      new(int64(1000)),
				CreatedAt:     1,
				UpdatedAt:     1,
			}}, 1, nil
		})

	recorder, c := newBuildersContext(t)
	(&handler{BuilderSvc: serviceObj}).ListRunners(c, &api.ListBuilderRunnersRequest{BuilderID: "1"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"total":1`)
	require.Contains(t, recorder.Body.String(), `"id":"2"`)
	require.Contains(t, recorder.Body.String(), `"duration":"1 second"`)
}

func TestListRunnersReturnsInternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcbuilder.NewMockBuilderService(ctrl)
	serviceObj.EXPECT().
		ListRunners(gomock.Any(), "1", gomock.Any(), gomock.Any()).
		Return(nil, int64(0), errors.New("list failed"))

	recorder, c := newBuildersContext(t)
	(&handler{BuilderSvc: serviceObj}).ListRunners(c, &api.ListBuilderRunnersRequest{BuilderID: "1"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestGetRunner(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcbuilder.NewMockBuilderService(ctrl)
	serviceObj.EXPECT().
		GetRunner(gomock.Any(), "2").
		Return(&models.BuilderRunner{
			ID:            "2",
			BuilderID:     "1",
			Log:           []byte("log"),
			Status:        enums.BuildStatusSuccess,
			StatusMessage: nil,
			Tag:           new("v1.0"),
			RawTag:        "v1.0",
			Description:   nil,
			ScmBranch:     nil,
			StartedAt:     nil,
			EndedAt:       nil,
			Duration:      nil,
			CreatedAt:     1,
			UpdatedAt:     1,
		}, nil)

	recorder, c := newBuildersContext(t)
	(&handler{BuilderSvc: serviceObj}).GetRunner(c, &api.GetRunner{BuilderID: "1", RunnerID: "2"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"id":"2"`)
	require.Contains(t, recorder.Body.String(), `"status":"Success"`)
}

func TestGetRunnerReturnsForbiddenOnBuilderMismatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcbuilder.NewMockBuilderService(ctrl)
	serviceObj.EXPECT().
		GetRunner(gomock.Any(), "2").
		Return(&models.BuilderRunner{
			ID:        "2",
			BuilderID: "other",
			Status:    enums.BuildStatusSuccess,
			RawTag:    "v1.0",
			CreatedAt: 1,
			UpdatedAt: 1,
		}, nil)

	recorder, c := newBuildersContext(t)
	(&handler{BuilderSvc: serviceObj}).GetRunner(c, &api.GetRunner{BuilderID: "1", RunnerID: "2"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusForbidden, recorder.Code)
}

func TestGetRunnerReturnsInternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcbuilder.NewMockBuilderService(ctrl)
	serviceObj.EXPECT().
		GetRunner(gomock.Any(), "2").
		Return(nil, errors.New("get failed"))

	recorder, c := newBuildersContext(t)
	(&handler{BuilderSvc: serviceObj}).GetRunner(c, &api.GetRunner{BuilderID: "1", RunnerID: "2"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestPostRunnerRun(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcbuilder.NewMockBuilderService(ctrl)
	serviceObj.EXPECT().
		RunRunner(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, req api.PostRunnerRun) (string, error) {
			require.Equal(t, "1", req.BuilderID)
			return "2", nil
		})

	recorder, c := newBuildersContext(t)
	(&handler{BuilderSvc: serviceObj}).PostRunnerRun(c, &api.PostRunnerRun{BuilderID: "1"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusCreated, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"runner_id":"2"`)
}

func TestPostRunnerRunReturnsInternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcbuilder.NewMockBuilderService(ctrl)
	serviceObj.EXPECT().
		RunRunner(gomock.Any(), gomock.Any()).
		Return("", errors.New("run failed"))

	recorder, c := newBuildersContext(t)
	(&handler{BuilderSvc: serviceObj}).PostRunnerRun(c, &api.PostRunnerRun{BuilderID: "1"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestGetRunnerRerun(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcbuilder.NewMockBuilderService(ctrl)
	serviceObj.EXPECT().
		RerunRunner(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, req api.GetRunnerStop) (string, error) {
			require.Equal(t, "2", req.RunnerID)
			return "3", nil
		})

	recorder, c := newBuildersContext(t)
	(&handler{BuilderSvc: serviceObj}).GetRunnerRerun(c, &api.GetRunnerStop{RunnerID: "2"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"runner_id":"3"`)
}

func TestGetRunnerStop(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcbuilder.NewMockBuilderService(ctrl)
	serviceObj.EXPECT().
		StopRunner(gomock.Any(), gomock.Any()).
		Return(nil)

	recorder, c := newBuildersContext(t)
	(&handler{BuilderSvc: serviceObj}).GetRunnerStop(c, &api.GetRunnerStop{RunnerID: "2"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusNoContent, recorder.Code)
}

func TestGetRunnerStopReturnsInternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcbuilder.NewMockBuilderService(ctrl)
	serviceObj.EXPECT().
		StopRunner(gomock.Any(), gomock.Any()).
		Return(errors.New("stop failed"))

	recorder, c := newBuildersContext(t)
	(&handler{BuilderSvc: serviceObj}).GetRunnerStop(c, &api.GetRunnerStop{RunnerID: "2"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestGetRunnerLogReturnsInternalErrorWhenBuilderLookupFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcbuilder.NewMockBuilderService(ctrl)
	serviceObj.EXPECT().
		GetBuilderByRepositoryID(gomock.Any(), "1").
		Return(nil, errors.New("get builder failed"))

	recorder, c := newBuildersContext(t)
	(&handler{BuilderSvc: serviceObj}).GetRunnerLog(c, &api.GetRunnerLog{RepositoryID: "1", BuilderID: "1", RunnerID: "2"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestGetRunnerLogReturnsInternalErrorOnBuilderMismatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcbuilder.NewMockBuilderService(ctrl)
	serviceObj.EXPECT().
		GetBuilderByRepositoryID(gomock.Any(), "1").
		Return(&models.Builder{ID: "other"}, nil)

	recorder, c := newBuildersContext(t)
	(&handler{BuilderSvc: serviceObj}).GetRunnerLog(c, &api.GetRunnerLog{RepositoryID: "1", BuilderID: "1", RunnerID: "2"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestGetRunnerLogReturnsInternalErrorWhenRunnerLookupFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcbuilder.NewMockBuilderService(ctrl)
	serviceObj.EXPECT().
		GetBuilderByRepositoryID(gomock.Any(), "1").
		Return(&models.Builder{ID: "1"}, nil)
	serviceObj.EXPECT().
		GetRunner(gomock.Any(), "2").
		Return(nil, errors.New("get runner failed"))

	recorder, c := newBuildersContext(t)
	(&handler{BuilderSvc: serviceObj}).GetRunnerLog(c, &api.GetRunnerLog{RepositoryID: "1", BuilderID: "1", RunnerID: "2"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func newBuildersContext(t *testing.T) (*httptest.ResponseRecorder, *gin.Context) {
	t.Helper()
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Set(consts.ContextUser, &models.User{ID: "10", Role: enums.UserRoleAdmin})
	return recorder, c
}
