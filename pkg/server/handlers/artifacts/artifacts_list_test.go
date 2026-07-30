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

package artifacts

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/dal/models"
	svcartifact "github.com/go-sigma/sigma/pkg/service/artifacts"
)

func TestListArtifact(t *testing.T) {
	ctrl := gomock.NewController(t)
	service := svcartifact.NewMockArtifactService(ctrl)
	request := api.ListArtifactRequest{Namespace: "library", Repository: "library/alpine"}
	service.EXPECT().ListArtifacts(gomock.Any(), request).Return([]*models.Artifact{{
		ID:     "artifact-1",
		Digest: "sha256:abc",
		Size:   123,
	}}, int64(1), nil)
	recorder, c := newArtifactContext()

	(&handler{ArtifactSvc: service}).ListArtifact(c, &request)
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"total":1`)
	require.Contains(t, recorder.Body.String(), `"id":"artifact-1"`)
}

func TestListServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	service := svcartifact.NewMockArtifactService(ctrl)
	request := api.ListArtifactRequest{Namespace: "library", Repository: "library/alpine"}
	service.EXPECT().ListArtifacts(gomock.Any(), request).Return(nil, int64(0), errors.New("failed"))
	recorder, c := newArtifactContext()

	(&handler{ArtifactSvc: service}).ListArtifact(c, &request)
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}
