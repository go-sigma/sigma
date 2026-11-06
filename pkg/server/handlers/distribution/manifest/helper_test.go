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

package manifest

import (
	"context"
	"encoding/json/v2"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	dtspecv1 "github.com/opencontainers/distribution-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/testkit"
)

const (
	testNamespaceID         = "namespace-1"
	testNamespaceName       = "test"
	testRepositoryName      = "test/busybox"
	testTagName             = "latest"
	testUserID              = "user-1"
	testManifestDigest      = "sha256:2776ee23722eaabcffed77dafd22b7a1da734971bf268a323b6819926dfe1ebd"
	testManifestContentType = "application/vnd.oci.image.manifest.v1+json"
)

// newManifestContext builds a gin context carrying an authenticated user, as the
// distribution middleware would for a registry request.
func newManifestContext(t *testing.T, method, target string, body io.Reader) (*httptest.ResponseRecorder, *gin.Context) {
	t.Helper()
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(method, target, body)
	c.Set(consts.ContextUser, &models.User{ID: testUserID, Username: "tester"})
	return recorder, c
}

func TestParseRef(t *testing.T) {
	h := &handler{}
	refs := h.parseRef("latest")
	assert.Equal(t, refs.Tag, "latest")

	refs = h.parseRef("sha256:f7d81d5be30e617068bf53a9b136400b13d91c0f54d097a72bf91127f43d0157")
	assert.Equal(t, refs.Digest.String(), "sha256:f7d81d5be30e617068bf53a9b136400b13d91c0f54d097a72bf91127f43d0157")
}

func TestDsErrorClientClosedRequest(t *testing.T) {
	ginEngine := testkit.NewGin()
	ginEngine.PUT("/v2/library/alpine/manifests/latest", func(c *gin.Context) {
		ctx, cancel := context.WithCancel(c.Request.Context())
		cancel()
		c.Request = c.Request.WithContext(ctx)
		(&handler{}).dsError(c, errcode.DSErrCodeUnknown)
	})

	recorder := httptest.NewRecorder()
	ginEngine.ServeHTTP(recorder, httptest.NewRequest(http.MethodPut, "/v2/library/alpine/manifests/latest", nil))
	require.Equal(t, errcode.StatusClientClosedRequest, recorder.Code)
	require.Empty(t, recorder.Body.Bytes())
}

func TestDsErrorMapsErrCode(t *testing.T) {
	ginEngine := testkit.NewGin()
	ginEngine.PUT("/v2/library/alpine/manifests/latest", func(c *gin.Context) {
		(&handler{}).dsError(c, errcode.DSErrCodeManifestUnknown)
	})

	recorder := httptest.NewRecorder()
	ginEngine.ServeHTTP(recorder, httptest.NewRequest(http.MethodPut, "/v2/library/alpine/manifests/latest", nil))
	require.Equal(t, http.StatusNotFound, recorder.Code)

	var response dtspecv1.ErrorResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.Len(t, response.Errors, 1)
	require.Equal(t, errcode.DSErrCodeManifestUnknown.Code, response.Errors[0].Code)
}
