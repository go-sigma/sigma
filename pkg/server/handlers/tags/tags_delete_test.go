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
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/go-sigma/sigma/pkg/api"
	svctag "github.com/go-sigma/sigma/pkg/service/tags"
)

func TestDeleteTag(t *testing.T) {
	ctrl := gomock.NewController(t)
	tagSvc := svctag.NewMockTagService(ctrl)
	tagSvc.EXPECT().DeleteTag(gomock.Any(), "namespace-1", "repository-1", "tag-1").Return(nil)

	recorder, c := newTagContext(t)
	(&handler{
		TagSvc:     tagSvc,
		Authorizer: fakeAuthorizer{tag: true},
	}).DeleteTag(c, &api.DeleteTagRequest{NamespaceID: "namespace-1", RepositoryID: "repository-1", ID: "tag-1"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusNoContent, recorder.Code)
}

func TestDeleteTagUnauthorized(t *testing.T) {
	recorder, c := newTagContext(t)
	(&handler{
		Authorizer: fakeAuthorizer{tag: false},
	}).DeleteTag(c, &api.DeleteTagRequest{NamespaceID: "namespace-1", RepositoryID: "repository-1", ID: "tag-1"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusUnauthorized, recorder.Code)
}
