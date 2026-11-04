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
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/authz"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	uploadmocks "github.com/go-sigma/sigma/pkg/service/distribution/upload/mocks"
)

func TestPostUpload(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := uploadmocks.NewMockDistributionUploadService(ctrl)
	authorizer := authz.NewMockAuthorizer(ctrl)
	svc.EXPECT().GetNamespaceByName(gomock.Any(), "library").Return(&models.Namespace{ID: "namespace-1"}, nil)
	authorizer.EXPECT().Namespace(gomock.Any(), gomock.Any(), "namespace-1", enums.AuthManage).Return(true, nil)
	svc.EXPECT().PostUpload(gomock.Any(), "library/alpine/blobs/uploads", "", gomock.Any(), "application/octet-stream").Return("upload-1", nil)
	recorder, c := newUploadContext(t, http.MethodPost, "/v2/library/alpine/blobs/uploads/", strings.NewReader("layer"))
	c.Request.Host = "registry.example.com"
	c.Request.Header.Set(consts.HeaderContentType, "application/octet-stream")
	c.Set(consts.ContextUser, &models.User{ID: "user-1"})

	(&handler{UploadSvc: svc, Authorizer: authorizer}).PostUpload(c)
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusAccepted, recorder.Code)
	require.Equal(t, "upload-1", recorder.Header().Get(consts.UploadUUID))
	require.Equal(t, "http://registry.example.com/v2/library/alpine/blobs/uploads/upload-1", recorder.Header().Get(consts.HeaderLocation))
	require.Equal(t, "0-0", recorder.Header().Get(consts.HeaderRange))
}

func TestPostUploadInvalidDigest(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := uploadmocks.NewMockDistributionUploadService(ctrl)
	authorizer := authz.NewMockAuthorizer(ctrl)
	svc.EXPECT().GetNamespaceByName(gomock.Any(), "library").Return(&models.Namespace{ID: "namespace-1"}, nil)
	authorizer.EXPECT().Namespace(gomock.Any(), gomock.Any(), "namespace-1", enums.AuthManage).Return(true, nil)
	recorder, c := newUploadContext(t, http.MethodPost, "/v2/library/alpine/blobs/uploads/?digest=invalid", nil)
	c.Set(consts.ContextUser, &models.User{ID: "user-1"})

	(&handler{UploadSvc: svc, Authorizer: authorizer}).PostUpload(c)
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusNotFound, recorder.Code)
}

const mountDigest = "sha256:fac304336cb8d9bcd2cbf47dbd4b6543677ee9c14bbe691621e76a07ad500b40"

func TestPostUploadMount(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := uploadmocks.NewMockDistributionUploadService(ctrl)
	authorizer := authz.NewMockAuthorizer(ctrl)
	// The target and source namespaces are both "test" (repositories test/test and test/test2).
	svc.EXPECT().GetNamespaceByName(gomock.Any(), "test").Return(&models.Namespace{ID: "namespace-1"}, nil).Times(2)
	authorizer.EXPECT().Namespace(gomock.Any(), gomock.Any(), "namespace-1", enums.AuthManage).Return(true, nil)
	authorizer.EXPECT().Namespace(gomock.Any(), gomock.Any(), "namespace-1", enums.AuthRead).Return(true, nil)
	svc.EXPECT().CanMountBlob(gomock.Any(), "test/test2", mountDigest).Return(true, nil)

	recorder, c := newUploadContext(t, http.MethodPost, "/v2/test/test/blobs/uploads/?mount="+mountDigest+"&from=test/test2", nil)
	c.Request.Host = "registry.example.com"
	c.Set(consts.ContextUser, &models.User{ID: "user-1"})

	(&handler{UploadSvc: svc, Authorizer: authorizer}).PostUpload(c)
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusCreated, recorder.Code)
	require.Equal(t, "http://registry.example.com/v2/test/test/blobs/"+mountDigest, recorder.Header().Get(consts.HeaderLocation))
	require.Equal(t, mountDigest, recorder.Header().Get(consts.ContentDigest))
}

func TestPostUploadMountFallback(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := uploadmocks.NewMockDistributionUploadService(ctrl)
	authorizer := authz.NewMockAuthorizer(ctrl)
	svc.EXPECT().GetNamespaceByName(gomock.Any(), "test").Return(&models.Namespace{ID: "namespace-1"}, nil).Times(2)
	authorizer.EXPECT().Namespace(gomock.Any(), gomock.Any(), "namespace-1", enums.AuthManage).Return(true, nil)
	authorizer.EXPECT().Namespace(gomock.Any(), gomock.Any(), "namespace-1", enums.AuthRead).Return(true, nil)
	svc.EXPECT().CanMountBlob(gomock.Any(), "test/test2", mountDigest).Return(false, nil)
	svc.EXPECT().PostUpload(gomock.Any(), "test/test/blobs/uploads", "", gomock.Any(), "").Return("upload-1", nil)

	recorder, c := newUploadContext(t, http.MethodPost, "/v2/test/test/blobs/uploads/?mount="+mountDigest+"&from=test/test2", nil)
	c.Request.Host = "registry.example.com"
	c.Set(consts.ContextUser, &models.User{ID: "user-1"})

	(&handler{UploadSvc: svc, Authorizer: authorizer}).PostUpload(c)
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusAccepted, recorder.Code)
	require.Equal(t, "upload-1", recorder.Header().Get(consts.UploadUUID))
}
