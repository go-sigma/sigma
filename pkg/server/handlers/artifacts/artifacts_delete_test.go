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
	svcartifact "github.com/go-sigma/sigma/pkg/service/artifacts"
)

func TestDeleteArtifact(t *testing.T) {
	ctrl := gomock.NewController(t)
	service := svcartifact.NewMockArtifactService(ctrl)
	service.EXPECT().DeleteArtifact(gomock.Any(), "library/alpine", "sha256:abc").Return(nil)
	recorder, c := newArtifactContext()

	(&handler{ArtifactSvc: service}).DeleteArtifact(c, &api.DeleteArtifactRequest{
		Repository: "library/alpine",
		Digest:     "sha256:abc",
	})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusNoContent, recorder.Code)
}

func TestDeleteArtifactServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	service := svcartifact.NewMockArtifactService(ctrl)
	service.EXPECT().DeleteArtifact(gomock.Any(), "library/alpine", "sha256:abc").Return(errors.New("failed"))
	recorder, c := newArtifactContext()

	(&handler{ArtifactSvc: service}).DeleteArtifact(c, &api.DeleteArtifactRequest{
		Repository: "library/alpine",
		Digest:     "sha256:abc",
	})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}
