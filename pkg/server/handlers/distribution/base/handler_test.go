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

package distribution

import (
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
	reporegistry "github.com/go-sigma/sigma/pkg/dal/repository/registry"
	"github.com/go-sigma/sigma/pkg/server/handlers/distribution"
	svcrepository "github.com/go-sigma/sigma/pkg/service/repositories"
	repositoriesmocks "github.com/go-sigma/sigma/pkg/service/repositories/mocks"
	svctag "github.com/go-sigma/sigma/pkg/service/tags"
	tagsmocks "github.com/go-sigma/sigma/pkg/service/tags/mocks"
)

func TestFactory(t *testing.T) {
	digCon := dig.New()
	require.NoError(t, digCon.Provide(func() *config.Configuration { return &config.Configuration{} }))
	require.NoError(t, digCon.Provide(func() authz.Authorizer { return nil }))
	require.NoError(t, digCon.Provide(func() reporegistry.RepositoryRepository { return nil }))
	require.NoError(t, digCon.Provide(func() reporegistry.TagRepository { return nil }))
	require.NoError(t, digCon.Provide(func() svcrepository.Service { return nil }))
	require.NoError(t, digCon.Provide(func() svctag.Service { return nil }))
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/v2/test-none-exist", nil)
	f := &factory{}
	require.NoError(t, f.Prepare(digCon))
	require.Equal(t, distribution.ErrNext, f.Initialize(c))
}

// TestFactoryListTagsWithQuery ensures pagination query parameters do not stop the tags
// list route from matching: /v2/<name>/tags/list?n=<n>&last=<tag> is the standard way
// clients page through tags.
func TestFactoryListTagsWithQuery(t *testing.T) {
	ctrl := gomock.NewController(t)
	repoSvc := repositoriesmocks.NewMockRepositoryService(ctrl)
	tagSvc := tagsmocks.NewMockTagService(ctrl)
	authorizer := authz.NewMockAuthorizer(ctrl)
	repoSvc.EXPECT().GetRepositoryByName(gomock.Any(), "library/alpine").Return(&models.Repository{ID: "repository-1", NamespaceID: "namespace-1"}, nil)
	authorizer.EXPECT().Repository(gomock.Any(), gomock.Any(), "repository-1", enums.AuthRead).Return(true, nil)
	tagSvc.EXPECT().ListTags(gomock.Any(), "namespace-1", "repository-1", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]*models.Tag{{Name: "latest"}}, int64(1), nil)

	recorder, c := newDistributionContext("/v2/library/alpine/tags/list?n=10&last=latest")
	c.Set(consts.ContextUser, &models.User{ID: "user-1"})

	f := &factory{handler: handler{RepoSvc: repoSvc, TagSvc: tagSvc, Authorizer: authorizer}}
	require.NoError(t, f.Initialize(c))
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusOK, recorder.Code)
	require.JSONEq(t, `{"name":"library/alpine","tags":["latest"]}`, recorder.Body.String())
}
