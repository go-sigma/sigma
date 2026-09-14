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

package webhooks

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/authz"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	svcwebhook "github.com/go-sigma/sigma/pkg/service/webhooks"
)

func TestPostWebhook(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcwebhook.NewMockWebhookService(ctrl)
	serviceObj.EXPECT().
		CreateWebhook(gomock.Any(), "10", gomock.Any()).
		DoAndReturn(func(_ context.Context, userID string, req api.PostWebhookRequest) error {
			require.Equal(t, "10", userID)
			require.Equal(t, "http://example.com/hook", req.URL)
			require.Nil(t, req.NamespaceID)
			return nil
		})

	recorder, c := newWebhooksContext(t)
	(&handler{
		WebhookSvc: serviceObj,
		Authorizer: fakeAuthorizer{},
	}).PostWebhook(c, &api.PostWebhookRequest{URL: "http://example.com/hook"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusCreated, recorder.Code)
}

func TestPostWebhookWithNamespace(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcwebhook.NewMockWebhookService(ctrl)
	serviceObj.EXPECT().
		CreateWebhook(gomock.Any(), "10", gomock.Any()).
		Return(nil)

	authorizer := fakeAuthorizer{
		namespace: func(ctx context.Context, user models.User, namespaceID string, auth enums.Auth) (bool, error) {
			require.Equal(t, "10", user.ID)
			require.Equal(t, "1", namespaceID)
			require.Equal(t, enums.AuthManage, auth)
			return true, nil
		},
	}

	recorder, c := newWebhooksContext(t)
	(&handler{
		WebhookSvc: serviceObj,
		Authorizer: authorizer,
	}).PostWebhook(c, &api.PostWebhookRequest{NamespaceID: new("1"), URL: "http://example.com/hook"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusCreated, recorder.Code)
}

func TestPostWebhookReturnsForbiddenForNonAdmin(t *testing.T) {
	t.Helper()
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Set(consts.ContextUser, &models.User{ID: "10", Role: enums.UserRoleUser})

	(&handler{}).PostWebhook(c, &api.PostWebhookRequest{URL: "http://example.com/hook"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestPostWebhookReturnsUnauthorizedWhenAuthFails(t *testing.T) {
	authorizer := fakeAuthorizer{
		namespace: func(context.Context, models.User, string, enums.Auth) (bool, error) {
			return false, nil
		},
	}

	recorder, c := newWebhooksContext(t)
	(&handler{Authorizer: authorizer}).PostWebhook(c, &api.PostWebhookRequest{NamespaceID: new("1"), URL: "http://example.com/hook"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestPostWebhookReturnsNotFoundWhenNamespaceMissing(t *testing.T) {
	authorizer := fakeAuthorizer{
		namespace: func(context.Context, models.User, string, enums.Auth) (bool, error) {
			return false, gorm.ErrRecordNotFound
		},
	}

	recorder, c := newWebhooksContext(t)
	(&handler{Authorizer: authorizer}).PostWebhook(c, &api.PostWebhookRequest{NamespaceID: new("1"), URL: "http://example.com/hook"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestPostWebhookReturnsInternalErrorWhenAuthCheckFails(t *testing.T) {
	authorizer := fakeAuthorizer{
		namespace: func(context.Context, models.User, string, enums.Auth) (bool, error) {
			return false, errors.New("namespace find failed")
		},
	}

	recorder, c := newWebhooksContext(t)
	(&handler{Authorizer: authorizer}).PostWebhook(c, &api.PostWebhookRequest{NamespaceID: new("1"), URL: "http://example.com/hook"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestPostWebhookReturnsErrCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcwebhook.NewMockWebhookService(ctrl)
	serviceObj.EXPECT().
		CreateWebhook(gomock.Any(), "10", gomock.Any()).
		Return(errcode.HTTPErrCodeBadRequest.Detail("URL is invalid, should start with 'http://' or 'https://'"))

	recorder, c := newWebhooksContext(t)
	(&handler{WebhookSvc: serviceObj}).PostWebhook(c, &api.PostWebhookRequest{URL: "http://example.com/hook"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestPostWebhookReturnsInternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcwebhook.NewMockWebhookService(ctrl)
	serviceObj.EXPECT().
		CreateWebhook(gomock.Any(), "10", gomock.Any()).
		Return(errors.New("create failed"))

	recorder, c := newWebhooksContext(t)
	(&handler{WebhookSvc: serviceObj}).PostWebhook(c, &api.PostWebhookRequest{URL: "http://example.com/hook"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestListWebhook(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcwebhook.NewMockWebhookService(ctrl)
	serviceObj.EXPECT().
		ListWebhooks(gomock.Any(), nil, gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, namespaceID *string, pagination api.Pagination, sort api.Sortable) ([]*models.Webhook, int64, error) {
			require.Nil(t, namespaceID)
			return []*models.Webhook{{
				ID:                "1",
				NamespaceID:       nil,
				URL:               "http://example.com/hook",
				Secret:            nil,
				SslVerify:         true,
				RetryTimes:        3,
				RetryDuration:     5,
				Enable:            true,
				EventNamespace:    nil,
				EventRepository:   true,
				EventTag:          true,
				EventArtifact:     true,
				EventMember:       true,
				EventDaemonTaskGc: true,
				CreatedAt:         1,
				UpdatedAt:         1,
			}}, 1, nil
		})

	recorder, c := newWebhooksContext(t)
	(&handler{WebhookSvc: serviceObj}).ListWebhook(c, &api.ListWebhookRequest{})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"total":1`)
	require.Contains(t, recorder.Body.String(), `"url":"http://example.com/hook"`)
}

func TestListWebhookWithNamespace(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcwebhook.NewMockWebhookService(ctrl)
	serviceObj.EXPECT().
		ListWebhooks(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, namespaceID *string, pagination api.Pagination, sort api.Sortable) ([]*models.Webhook, int64, error) {
			require.NotNil(t, namespaceID)
			require.Equal(t, "1", *namespaceID)
			return nil, 0, nil
		})

	authorizer := fakeAuthorizer{
		namespace: func(context.Context, models.User, string, enums.Auth) (bool, error) {
			return true, nil
		},
	}

	recorder, c := newWebhooksContext(t)
	(&handler{
		WebhookSvc: serviceObj,
		Authorizer: authorizer,
	}).ListWebhook(c, &api.ListWebhookRequest{NamespaceID: new("1")})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusOK, recorder.Code)
}

func TestListWebhookReturnsInternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcwebhook.NewMockWebhookService(ctrl)
	serviceObj.EXPECT().
		ListWebhooks(gomock.Any(), nil, gomock.Any(), gomock.Any()).
		Return(nil, int64(0), errors.New("list failed"))

	recorder, c := newWebhooksContext(t)
	(&handler{WebhookSvc: serviceObj}).ListWebhook(c, &api.ListWebhookRequest{})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestGetWebhook(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcwebhook.NewMockWebhookService(ctrl)
	serviceObj.EXPECT().
		GetWebhook(gomock.Any(), "1").
		Return(&models.Webhook{
			ID:                "1",
			NamespaceID:       nil,
			URL:               "http://example.com/hook",
			Secret:            nil,
			SslVerify:         true,
			RetryTimes:        3,
			RetryDuration:     5,
			Enable:            true,
			EventNamespace:    nil,
			EventRepository:   true,
			EventTag:          true,
			EventArtifact:     true,
			EventMember:       true,
			EventDaemonTaskGc: true,
			CreatedAt:         1,
			UpdatedAt:         1,
		}, nil)

	recorder, c := newWebhooksContext(t)
	(&handler{WebhookSvc: serviceObj}).GetWebhook(c, &api.GetWebhookRequest{ID: "1"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"id":"1"`)
	require.Contains(t, recorder.Body.String(), `"url":"http://example.com/hook"`)
}

func TestGetWebhookReturnsUnauthorizedForNonAdmin(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcwebhook.NewMockWebhookService(ctrl)
	serviceObj.EXPECT().
		GetWebhook(gomock.Any(), "1").
		Return(&models.Webhook{
			ID:          "1",
			NamespaceID: nil,
			URL:         "http://example.com/hook",
			CreatedAt:   1,
			UpdatedAt:   1,
		}, nil)

	t.Helper()
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Set(consts.ContextUser, &models.User{ID: "10", Role: enums.UserRoleUser})

	(&handler{WebhookSvc: serviceObj}).GetWebhook(c, &api.GetWebhookRequest{ID: "1"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestGetWebhookReturnsInternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcwebhook.NewMockWebhookService(ctrl)
	serviceObj.EXPECT().
		GetWebhook(gomock.Any(), "1").
		Return(nil, errors.New("get failed"))

	recorder, c := newWebhooksContext(t)
	(&handler{WebhookSvc: serviceObj}).GetWebhook(c, &api.GetWebhookRequest{ID: "1"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestPutWebhook(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcwebhook.NewMockWebhookService(ctrl)
	serviceObj.EXPECT().
		GetWebhook(gomock.Any(), "1").
		Return(&models.Webhook{
			ID:          "1",
			NamespaceID: nil,
			URL:         "http://example.com/hook",
			CreatedAt:   1,
			UpdatedAt:   1,
		}, nil)
	serviceObj.EXPECT().
		UpdateWebhook(gomock.Any(), "10", "1", gomock.Any()).
		Return(nil)

	recorder, c := newWebhooksContext(t)
	(&handler{WebhookSvc: serviceObj}).PutWebhook(c, &api.PutWebhookRequest{ID: "1"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusNoContent, recorder.Code)
}

func TestPutWebhookReturnsInternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcwebhook.NewMockWebhookService(ctrl)
	serviceObj.EXPECT().
		GetWebhook(gomock.Any(), "1").
		Return(&models.Webhook{
			ID:          "1",
			NamespaceID: nil,
			URL:         "http://example.com/hook",
			CreatedAt:   1,
			UpdatedAt:   1,
		}, nil)
	serviceObj.EXPECT().
		UpdateWebhook(gomock.Any(), "10", "1", gomock.Any()).
		Return(errors.New("update failed"))

	recorder, c := newWebhooksContext(t)
	(&handler{WebhookSvc: serviceObj}).PutWebhook(c, &api.PutWebhookRequest{ID: "1"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestDeleteWebhook(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcwebhook.NewMockWebhookService(ctrl)
	serviceObj.EXPECT().
		GetWebhook(gomock.Any(), "1").
		Return(&models.Webhook{
			ID:          "1",
			NamespaceID: nil,
			URL:         "http://example.com/hook",
			CreatedAt:   1,
			UpdatedAt:   1,
		}, nil)
	serviceObj.EXPECT().
		DeleteWebhook(gomock.Any(), "10", "1").
		Return(nil)

	recorder, c := newWebhooksContext(t)
	(&handler{WebhookSvc: serviceObj}).DeleteWebhook(c, &api.DeleteWebhookRequest{ID: "1"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusNoContent, recorder.Code)
}

func TestGetWebhookPing(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcwebhook.NewMockWebhookService(ctrl)
	serviceObj.EXPECT().
		GetWebhook(gomock.Any(), "1").
		Return(&models.Webhook{
			ID:          "1",
			NamespaceID: nil,
			URL:         "http://example.com/hook",
			CreatedAt:   1,
			UpdatedAt:   1,
		}, nil)
	serviceObj.EXPECT().
		PingWebhook(gomock.Any(), "10", "1").
		Return(nil)

	recorder, c := newWebhooksContext(t)
	(&handler{WebhookSvc: serviceObj}).GetWebhookPing(c, &api.GetWebhookPingRequest{WebhookID: "1"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusNoContent, recorder.Code)
}

func TestGetWebhookLog(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcwebhook.NewMockWebhookService(ctrl)
	serviceObj.EXPECT().
		GetWebhook(gomock.Any(), "1").
		Return(&models.Webhook{
			ID:          "1",
			NamespaceID: nil,
			URL:         "http://example.com/hook",
			CreatedAt:   1,
			UpdatedAt:   1,
		}, nil)
	serviceObj.EXPECT().
		GetWebhookLog(gomock.Any(), "2").
		Return(&models.WebhookLog{
			ID:           "2",
			WebhookID:    "1",
			Action:       enums.WebhookActionCreate,
			ResourceType: enums.WebhookResourceTypeWebhook,
			StatusCode:   200,
			TraceContext: []byte("{}"),
			ReqHeader:    []byte("{}"),
			ReqBody:      []byte("{}"),
			RespHeader:   []byte("{}"),
			RespBody:     []byte("{}"),
			CreatedAt:    1,
			UpdatedAt:    1,
		}, nil)

	recorder, c := newWebhooksContext(t)
	(&handler{WebhookSvc: serviceObj}).GetWebhookLog(c, &api.GetWebhookLogRequest{WebhookID: "1", WebhookLogID: "2"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"id":"2"`)
	require.Contains(t, recorder.Body.String(), `"status_code":200`)
}

func TestListWebhookLogs(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcwebhook.NewMockWebhookService(ctrl)
	serviceObj.EXPECT().
		GetWebhook(gomock.Any(), "1").
		Return(&models.Webhook{
			ID:          "1",
			NamespaceID: nil,
			URL:         "http://example.com/hook",
			CreatedAt:   1,
			UpdatedAt:   1,
		}, nil)
	serviceObj.EXPECT().
		ListWebhookLogs(gomock.Any(), "1", gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, webhookID string, pagination api.Pagination, sort api.Sortable) ([]*models.WebhookLog, int64, error) {
			require.Equal(t, "1", webhookID)
			return []*models.WebhookLog{{
				ID:           "2",
				WebhookID:    "1",
				Action:       enums.WebhookActionCreate,
				ResourceType: enums.WebhookResourceTypeWebhook,
				StatusCode:   200,
				TraceContext: []byte("{}"),
				CreatedAt:    1,
				UpdatedAt:    1,
			}}, 1, nil
		})

	recorder, c := newWebhooksContext(t)
	(&handler{WebhookSvc: serviceObj}).ListWebhookLogs(c, &api.ListWebhookLogRequest{WebhookID: "1"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"total":1`)
	require.Contains(t, recorder.Body.String(), `"id":"2"`)
}

func TestDeleteWebhookLog(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcwebhook.NewMockWebhookService(ctrl)
	serviceObj.EXPECT().
		GetWebhook(gomock.Any(), "1").
		Return(&models.Webhook{
			ID:          "1",
			NamespaceID: nil,
			URL:         "http://example.com/hook",
			CreatedAt:   1,
			UpdatedAt:   1,
		}, nil)
	serviceObj.EXPECT().
		DeleteWebhookLog(gomock.Any(), "10", "1", "2").
		Return(nil)

	recorder, c := newWebhooksContext(t)
	(&handler{WebhookSvc: serviceObj}).DeleteWebhookLog(c, &api.DeleteWebhookLogRequest{WebhookID: "1", WebhookLogID: "2"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusNoContent, recorder.Code)
}

func TestGetWebhookLogResend(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcwebhook.NewMockWebhookService(ctrl)
	serviceObj.EXPECT().
		GetWebhook(gomock.Any(), "1").
		Return(&models.Webhook{
			ID:          "1",
			NamespaceID: nil,
			URL:         "http://example.com/hook",
			CreatedAt:   1,
			UpdatedAt:   1,
		}, nil)
	serviceObj.EXPECT().
		ResendWebhookLog(gomock.Any(), "10", "1", "2").
		Return(nil)

	recorder, c := newWebhooksContext(t)
	(&handler{WebhookSvc: serviceObj}).GetWebhookLogResend(c, &api.GetWebhookLogResendRequest{WebhookID: "1", WebhookLogID: "2"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusNoContent, recorder.Code)
}

func TestGetWebhookLogResendReturnsInternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svcwebhook.NewMockWebhookService(ctrl)
	serviceObj.EXPECT().
		GetWebhook(gomock.Any(), "1").
		Return(&models.Webhook{
			ID:          "1",
			NamespaceID: nil,
			URL:         "http://example.com/hook",
			CreatedAt:   1,
			UpdatedAt:   1,
		}, nil)
	serviceObj.EXPECT().
		ResendWebhookLog(gomock.Any(), "10", "1", "2").
		Return(errors.New("resend failed"))

	recorder, c := newWebhooksContext(t)
	(&handler{WebhookSvc: serviceObj}).GetWebhookLogResend(c, &api.GetWebhookLogResendRequest{WebhookID: "1", WebhookLogID: "2"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func newWebhooksContext(t *testing.T) (*httptest.ResponseRecorder, *gin.Context) {
	t.Helper()
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Set(consts.ContextUser, &models.User{ID: "10", Role: enums.UserRoleAdmin})
	return recorder, c
}

type fakeAuthorizer struct {
	namespace func(ctx context.Context, user models.User, namespaceID string, auth enums.Auth) (bool, error)
}

var _ authz.Authorizer = fakeAuthorizer{}

func (f fakeAuthorizer) Authorize(context.Context, string, bool, string, string) (bool, error) {
	return false, nil
}

func (f fakeAuthorizer) Namespace(ctx context.Context, user models.User, namespaceID string, auth enums.Auth) (bool, error) {
	if f.namespace == nil {
		return false, nil
	}
	return f.namespace(ctx, user, namespaceID, auth)
}

func (f fakeAuthorizer) NamespaceRole(context.Context, models.User, string) (*enums.NamespaceRole, error) {
	return nil, nil
}

func (f fakeAuthorizer) NamespacesRole(context.Context, models.User, []string) (map[string]*enums.NamespaceRole, error) {
	return nil, nil
}

func (f fakeAuthorizer) Repository(context.Context, models.User, string, enums.Auth) (bool, error) {
	return false, nil
}

func (f fakeAuthorizer) Tag(context.Context, models.User, string, enums.Auth) (bool, error) {
	return false, nil
}

func (f fakeAuthorizer) Artifact(context.Context, models.User, string, enums.Auth) (bool, error) {
	return false, nil
}
