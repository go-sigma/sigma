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
	svcupload "github.com/go-sigma/sigma/pkg/service/distribution/upload"
)

const uploadDigest = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestPutUpload(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := svcupload.NewMockDistributionUploadService(ctrl)
	authorizer := authz.NewMockAuthorizer(ctrl)
	svc.EXPECT().GetNamespaceByName(gomock.Any(), "library").Return(&models.Namespace{ID: "namespace-1"}, nil)
	authorizer.EXPECT().Namespace(gomock.Any(), gomock.Any(), "namespace-1", enums.AuthManage).Return(true, nil)
	svc.EXPECT().PutUpload(gomock.Any(), "upload-1", uploadDigest, gomock.Any(), "library/alpine/blobs/uploads", "application/octet-stream", int64(5)).Return(int64(10), int64(5), nil)
	path := "/v2/library/alpine/blobs/uploads/upload-1?digest=" + uploadDigest
	recorder, c := newUploadContext(t, http.MethodPut, path, strings.NewReader("layer"))
	c.Request.Host = "registry.example.com"
	c.Request.Header.Set(consts.HeaderContentType, "application/octet-stream")
	c.Request.Header.Set(consts.HeaderContentLength, "5")
	c.Set(consts.ContextUser, &models.User{ID: "user-1"})

	(&handler{UploadSvc: svc, Authorizer: authorizer}).PutUpload(c)
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusCreated, recorder.Code)
	require.Equal(t, uploadDigest, recorder.Header().Get(consts.ContentDigest))
	require.Equal(t, "10-15", recorder.Header().Get(consts.HeaderContentRange))
	require.Equal(t, "http://registry.example.com/v2/library/alpine/blobs/uploads/upload-1", recorder.Header().Get(consts.HeaderLocation))
}

func TestPutUploadInvalidDigest(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := svcupload.NewMockDistributionUploadService(ctrl)
	authorizer := authz.NewMockAuthorizer(ctrl)
	svc.EXPECT().GetNamespaceByName(gomock.Any(), "library").Return(&models.Namespace{ID: "namespace-1"}, nil)
	authorizer.EXPECT().Namespace(gomock.Any(), gomock.Any(), "namespace-1", enums.AuthManage).Return(true, nil)
	recorder, c := newUploadContext(t, http.MethodPut, "/v2/library/alpine/blobs/uploads/upload-1?digest=invalid", nil)
	c.Set(consts.ContextUser, &models.User{ID: "user-1"})

	(&handler{UploadSvc: svc, Authorizer: authorizer}).PutUpload(c)
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusBadRequest, recorder.Code)
}
