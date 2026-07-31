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

package blob

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/authz"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	svcblob "github.com/go-sigma/sigma/pkg/service/distribution/blob"
)

const blobDigest = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestGetBlob(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := svcblob.NewMockDistributionBlobService(ctrl)
	authorizer := authz.NewMockAuthorizer(ctrl)
	svc.EXPECT().GetNamespaceByName(gomock.Any(), "library").Return(&models.Namespace{ID: "namespace-1"}, nil)
	authorizer.EXPECT().Namespace(gomock.Any(), gomock.Any(), "namespace-1", enums.AuthRead).Return(true, nil)
	svc.EXPECT().GetBlob(gomock.Any(), "namespace-1", blobDigest, http.MethodGet, "/v2/library/alpine/blobs/"+blobDigest).
		Return(io.NopCloser(strings.NewReader("layer")), &models.Blob{Size: 5, ContentType: "application/octet-stream", Digest: blobDigest}, "", nil)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/v2/library/alpine/blobs/"+blobDigest, nil)
	c.Set(consts.ContextUser, &models.User{ID: "user-1"})

	(&handler{BlobSvc: svc, Authorizer: authorizer}).GetBlob(c)
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "layer", recorder.Body.String())
	require.Equal(t, "5", recorder.Header().Get(consts.HeaderContentLength))
	require.Equal(t, blobDigest, recorder.Header().Get(consts.ContentDigest))
}

func TestGetBlobRedirect(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := svcblob.NewMockDistributionBlobService(ctrl)
	authorizer := authz.NewMockAuthorizer(ctrl)
	svc.EXPECT().GetNamespaceByName(gomock.Any(), "library").Return(&models.Namespace{ID: "namespace-1"}, nil)
	authorizer.EXPECT().Namespace(gomock.Any(), gomock.Any(), "namespace-1", enums.AuthRead).Return(true, nil)
	svc.EXPECT().GetBlob(gomock.Any(), "namespace-1", blobDigest, http.MethodGet, gomock.Any()).Return(nil, nil, "https://storage.example.com/layer", nil)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/v2/library/alpine/blobs/"+blobDigest, nil)
	c.Set(consts.ContextUser, &models.User{ID: "user-1"})

	(&handler{BlobSvc: svc, Authorizer: authorizer}).GetBlob(c)
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusMovedPermanently, recorder.Code)
	require.Equal(t, "https://storage.example.com/layer", recorder.Header().Get(consts.HeaderLocation))
}
