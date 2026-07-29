// Copyright 2024 sigma
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

package tags

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/dal/models"
	svctag "github.com/go-sigma/sigma/pkg/service/tags"
)

func TestListTag(t *testing.T) {
	ctrl := gomock.NewController(t)
	tagSvc := svctag.NewMockTagService(ctrl)
	tagSvc.EXPECT().ListTags(
		gomock.Any(), "namespace-1", "repository-1", nil, nil, api.Pagination{}, api.Sortable{},
	).Return([]*models.Tag{{
		ID:   "tag-1",
		Name: "latest",
		Artifact: &models.Artifact{
			ID:     "artifact-1",
			Digest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		},
	}}, int64(1), nil)
	tagSvc.EXPECT().GetArtifactRaw(gomock.Any(), "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").
		Return([]byte("manifest"), nil)

	recorder, c := newTagContext(t)
	(&handler{
		TagSvc:     tagSvc,
		Authorizer: fakeAuthorizer{repository: true},
	}).ListTag(c, &api.ListTagRequest{NamespaceID: "namespace-1", RepositoryID: "repository-1"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"total":1`)
	require.Contains(t, recorder.Body.String(), `"name":"latest"`)
}

func TestListTagUnauthorized(t *testing.T) {
	recorder, c := newTagContext(t)
	(&handler{
		Authorizer: fakeAuthorizer{repository: false},
	}).ListTag(c, &api.ListTagRequest{NamespaceID: "namespace-1", RepositoryID: "repository-1"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusUnauthorized, recorder.Code)
}
