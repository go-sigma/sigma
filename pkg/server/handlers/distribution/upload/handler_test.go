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

package upload

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/dig"
	"go.uber.org/mock/gomock"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/authz"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	repoaudit "github.com/go-sigma/sigma/pkg/dal/repository/audit"
	reponamespace "github.com/go-sigma/sigma/pkg/dal/repository/namespace"
	reporegistry "github.com/go-sigma/sigma/pkg/dal/repository/registry"
	"github.com/go-sigma/sigma/pkg/server/handlers/distribution"
	svcupload "github.com/go-sigma/sigma/pkg/service/distribution/upload"
	uploadmocks "github.com/go-sigma/sigma/pkg/service/distribution/upload/mocks"
	"github.com/go-sigma/sigma/pkg/storage"
)

func newUploadContext(t *testing.T, method, target string, body io.Reader) (*httptest.ResponseRecorder, *gin.Context) {
	t.Helper()
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(method, target, body)
	return recorder, c
}

func TestFactory(t *testing.T) {
	digCon := dig.New()
	require.NoError(t, digCon.Provide(func() *config.Configuration { return &config.Configuration{} }))
	require.NoError(t, digCon.Provide(func() authz.Authorizer { return nil }))
	require.NoError(t, digCon.Provide(func() repoaudit.AuditRepository { return nil }))
	require.NoError(t, digCon.Provide(func() reponamespace.NamespaceRepository { return nil }))
	require.NoError(t, digCon.Provide(func() reporegistry.RepositoryRepository { return nil }))
	require.NoError(t, digCon.Provide(func() reporegistry.BlobRepository { return nil }))
	require.NoError(t, digCon.Provide(func() svcupload.Service { return nil }))
	require.NoError(t, digCon.Provide(func() storage.StorageDriver { return nil }))
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/v2/test-none-exist", nil)
	f := &factory{}
	require.NoError(t, f.Prepare(digCon))
	require.Equal(t, distribution.ErrNext, f.Initialize(c))
}

// TestFactoryPostUploadWithQuery ensures the route matcher ignores query parameters:
// clients create an upload with ?digest=... or with cross-repository mount parameters
// (?mount=...&from=...), and treating the raw RequestURI as the path made those requests
// fall through to 405 Method Not Allowed.
func TestFactoryPostUploadWithQuery(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := uploadmocks.NewMockDistributionUploadService(ctrl)
	authorizer := authz.NewMockAuthorizer(ctrl)
	svc.EXPECT().GetNamespaceByName(gomock.Any(), "test").Return(&models.Namespace{ID: "namespace-1"}, nil)
	authorizer.EXPECT().Namespace(gomock.Any(), gomock.Any(), "namespace-1", enums.AuthManage).Return(true, nil)
	svc.EXPECT().PostUpload(gomock.Any(), "test/test/blobs/uploads", "", gomock.Any(), "application/octet-stream").Return("upload-1", nil)

	recorder, c := newUploadContext(t, http.MethodPost, "/v2/test/test/blobs/uploads/?mount=sha256:fac304336cb8d9bcd2cbf47dbd4b6543677ee9c14bbe691621e76a07ad500b40&from=test/test2", nil)
	c.Request.Host = "registry.example.com"
	c.Request.Header.Set(consts.HeaderContentType, "application/octet-stream")
	c.Set(consts.ContextUser, &models.User{ID: "user-1"})

	f := &factory{handler: handler{UploadSvc: svc, Authorizer: authorizer}}
	require.NoError(t, f.Initialize(c))
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusAccepted, recorder.Code)
	require.Equal(t, "upload-1", recorder.Header().Get(consts.UploadUUID))
	require.Equal(t, "http://registry.example.com/v2/test/test/blobs/uploads/upload-1", recorder.Header().Get(consts.HeaderLocation))
}

// TestFactoryUploadSessionWithQuery ensures session routes still match when the client
// appends query parameters, such as the digest on the final PUT.
func TestFactoryUploadSessionWithQuery(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := uploadmocks.NewMockDistributionUploadService(ctrl)
	authorizer := authz.NewMockAuthorizer(ctrl)
	svc.EXPECT().GetNamespaceByName(gomock.Any(), "library").Return(&models.Namespace{ID: "namespace-1"}, nil)
	authorizer.EXPECT().Namespace(gomock.Any(), gomock.Any(), "namespace-1", enums.AuthManage).Return(true, nil)

	recorder, c := newUploadContext(t, http.MethodPut, "/v2/library/alpine/blobs/uploads/upload-1?digest=invalid", nil)
	c.Set(consts.ContextUser, &models.User{ID: "user-1"})

	f := &factory{handler: handler{UploadSvc: svc, Authorizer: authorizer}}
	require.NoError(t, f.Initialize(c))
	c.Writer.WriteHeaderNow()

	// PutUpload reaches the digest validation and rejects the invalid digest, which proves
	// the request was dispatched to the upload session handler.
	require.Equal(t, http.StatusBadRequest, recorder.Code)
}
