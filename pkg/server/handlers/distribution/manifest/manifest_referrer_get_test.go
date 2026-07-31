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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	imgspecv1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	svcmanifest "github.com/go-sigma/sigma/pkg/service/distribution/manifest"
)

const referrerDigest = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestGetReferrer(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := svcmanifest.NewMockDistributionManifestService(ctrl)
	svc.EXPECT().GetReferrer(gomock.Any(), "library/alpine", referrerDigest, []string{"application/example", "application/other"}).Return([]byte(`{"manifests":[]}`), nil)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/v2/library/alpine/referrers/"+referrerDigest+"?artifactType=application/example,application/other", nil)

	(&handler{ManifestSvc: svc}).GetReferrer(c)
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, imgspecv1.MediaTypeImageIndex, recorder.Header().Get("Content-Type"))
	require.Equal(t, "application/example,application/other", recorder.Header().Get("OCI-Filters-Applied"))
	require.JSONEq(t, `{"manifests":[]}`, recorder.Body.String())
}

func TestGetReferrerInvalidDigest(t *testing.T) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/v2/library/alpine/referrers/invalid", nil)

	(&handler{}).GetReferrer(c)
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusBadRequest, recorder.Code)
}
