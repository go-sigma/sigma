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

package tags

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/authz"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	svctag "github.com/go-sigma/sigma/pkg/service/tags"
)

func TestGetTag(t *testing.T) {
	ctrl := gomock.NewController(t)
	tagSvc := svctag.NewMockTagService(ctrl)
	tagSvc.EXPECT().GetTag(gomock.Any(), "tag-1").Return(&models.Tag{
		ID:   "tag-1",
		Name: "latest",
		Artifact: &models.Artifact{
			ID:     "artifact-1",
			Digest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		},
	}, nil)
	tagSvc.EXPECT().GetArtifactRaw(gomock.Any(), "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").
		Return([]byte("manifest"), nil)

	recorder, c := newTagContext(t)
	(&handler{
		TagSvc:     tagSvc,
		Authorizer: fakeAuthorizer{tag: true},
	}).GetTag(c, &api.GetTagRequest{NamespaceID: "namespace-1", RepositoryID: "repository-1", ID: "tag-1"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"id":"tag-1"`)
	require.Contains(t, recorder.Body.String(), `"name":"latest"`)
	require.Contains(t, recorder.Body.String(), `"raw":"manifest"`)
}

func TestGetTagUnauthorized(t *testing.T) {
	recorder, c := newTagContext(t)
	(&handler{
		Authorizer: fakeAuthorizer{tag: false},
	}).GetTag(c, &api.GetTagRequest{NamespaceID: "namespace-1", RepositoryID: "repository-1", ID: "tag-1"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func newTagContext(t *testing.T) (*httptest.ResponseRecorder, *gin.Context) {
	t.Helper()
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Set(consts.ContextUser, &models.User{ID: "user-1"})
	return recorder, c
}

type fakeAuthorizer struct {
	tag        bool
	repository bool
}

var _ authz.Authorizer = fakeAuthorizer{}

func (fakeAuthorizer) Authorize(context.Context, string, bool, string, string) (bool, error) {
	return false, nil
}

func (fakeAuthorizer) Namespace(context.Context, models.User, string, enums.Auth) (bool, error) {
	return false, nil
}

func (fakeAuthorizer) NamespaceRole(context.Context, models.User, string) (*enums.NamespaceRole, error) {
	return nil, nil
}

func (fakeAuthorizer) NamespacesRole(context.Context, models.User, []string) (map[string]*enums.NamespaceRole, error) {
	return nil, nil
}

func (f fakeAuthorizer) Repository(context.Context, models.User, string, enums.Auth) (bool, error) {
	return f.repository, nil
}

func (f fakeAuthorizer) Tag(context.Context, models.User, string, enums.Auth) (bool, error) {
	return f.tag, nil
}

func (fakeAuthorizer) Artifact(context.Context, models.User, string, enums.Auth) (bool, error) {
	return false, nil
}
