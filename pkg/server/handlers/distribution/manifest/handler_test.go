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

	"github.com/stretchr/testify/require"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/auth"
	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/server/handlers/distribution"
	"github.com/go-sigma/sigma/pkg/tests"
)

// func TestHandlerNew(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	daoMockArtifactServiceFactory := daomock.NewMockArtifactServiceFactory(ctrl)
// 	daoMockBlobServiceFactory := daomock.NewMockBlobServiceFactory(ctrl)
// 	daoMockTagServiceFactory := daomock.NewMockTagServiceFactory(ctrl)
// 	daoMockRepositoryServiceFactory := daomock.NewMockRepositoryServiceFactory(ctrl)

// 	handler := handlerNew(inject{
// 		artifactServiceFactory:   daoMockArtifactServiceFactory,
// 		blobServiceFactory:       daoMockBlobServiceFactory,
// 		tagServiceFactory:        daoMockTagServiceFactory,
// 		repositoryServiceFactory: daoMockRepositoryServiceFactory,
// 	})
// 	assert.NotNil(t, handler)

// 	req := httptest.NewRequest(http.MethodGet, "/v2/test-none-exist", nil)
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec := httptest.NewRecorder()
// 	c := echo.New().NewContext(req, rec)
// 	f := &factory{}
// 	err := f.Initialize(c)
// 	assert.ErrorIs(t, err, distribution.ErrNext)
// }

func TestFactory(t *testing.T) {
	digCon := dig.New()
	require.NoError(t, digCon.Provide(func() configs.Configuration { return configs.Configuration{} }))
	require.NoError(t, digCon.Provide(func() auth.AuthServiceFactory { return nil }))
	require.NoError(t, digCon.Provide(func() dao.AuditServiceFactory { return nil }))
	require.NoError(t, digCon.Provide(func() dao.NamespaceServiceFactory { return nil }))
	require.NoError(t, digCon.Provide(func() dao.RepositoryServiceFactory { return nil }))
	require.NoError(t, digCon.Provide(func() dao.ArtifactServiceFactory { return nil }))
	require.NoError(t, digCon.Provide(func() dao.TagServiceFactory { return nil }))
	require.NoError(t, digCon.Provide(func() dao.BlobServiceFactory { return nil }))
	c := tests.NewEcho().NewContext(httptest.NewRequest(http.MethodGet, "/v2/test-none-exist", nil), httptest.NewRecorder())
	require.Equal(t, distribution.ErrNext, factory{}.Initialize(c, digCon))
}
