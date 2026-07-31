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
	"testing"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/authz"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	svcrepository "github.com/go-sigma/sigma/pkg/service/repositories"
	svctag "github.com/go-sigma/sigma/pkg/service/tags"
)

func TestListTags(t *testing.T) {
	ctrl := gomock.NewController(t)
	repoSvc := svcrepository.NewMockRepositoryService(ctrl)
	tagSvc := svctag.NewMockTagService(ctrl)
	authorizer := authz.NewMockAuthorizer(ctrl)
	repoSvc.EXPECT().GetRepositoryByName(gomock.Any(), "library/alpine").Return(&models.Repository{ID: "repository-1", NamespaceID: "namespace-1"}, nil)
	authorizer.EXPECT().Repository(gomock.Any(), gomock.Any(), "repository-1", enums.AuthRead).Return(true, nil)
	tagSvc.EXPECT().ListTags(gomock.Any(), "namespace-1", "repository-1", nil, nil, gomock.Any(), gomock.Any()).Return([]*models.Tag{{Name: "latest"}}, int64(1), nil)
	recorder, c := newDistributionContext(http.MethodGet, "/v2/library/alpine/tags/list?n=10")
	c.Request.Host = "registry.example.com"
	c.Set(consts.ContextUser, &models.User{ID: "user-1"})

	(&handler{RepoSvc: repoSvc, TagSvc: tagSvc, Authorizer: authorizer}).ListTags(c)
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusOK, recorder.Code)
	require.JSONEq(t, `{"name":"library/alpine","tags":["latest"]}`, recorder.Body.String())
	require.Equal(t, `<http://registry.example.com/v2/library/alpine/tags/list?last=latest&n=10>; rel="next"`, recorder.Header().Get("Link"))
}

func TestListTagsInvalidRepository(t *testing.T) {
	recorder, c := newDistributionContext(http.MethodGet, "/v2/INVALID/tags/list")
	c.Set(consts.ContextUser, &models.User{ID: "user-1"})

	(&handler{}).ListTags(c)
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusBadRequest, recorder.Code)
}
