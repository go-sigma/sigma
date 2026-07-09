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

package blob

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/authz"
	"github.com/go-sigma/sigma/pkg/config"
	repoaudit "github.com/go-sigma/sigma/pkg/dal/repository/audit"
	reponamespace "github.com/go-sigma/sigma/pkg/dal/repository/namespace"
	reporegistry "github.com/go-sigma/sigma/pkg/dal/repository/registry"
	"github.com/go-sigma/sigma/pkg/server/handlers/distribution"
	svcblob "github.com/go-sigma/sigma/pkg/service/distribution/blob"
	"github.com/go-sigma/sigma/pkg/storage"
)

func TestFactory(t *testing.T) {
	digCon := dig.New()
	require.NoError(t, digCon.Provide(func() *config.Configuration { return &config.Configuration{} }))
	require.NoError(t, digCon.Provide(func() authz.Authorizer { return nil }))
	require.NoError(t, digCon.Provide(func() repoaudit.AuditRepository { return nil }))
	require.NoError(t, digCon.Provide(func() reponamespace.NamespaceRepository { return nil }))
	require.NoError(t, digCon.Provide(func() reporegistry.RepositoryRepository { return nil }))
	require.NoError(t, digCon.Provide(func() reporegistry.BlobRepository { return nil }))
	require.NoError(t, digCon.Provide(func() svcblob.DistributionBlobService { return nil }))
	require.NoError(t, digCon.Provide(func() storage.StorageDriver { return nil }))
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/v2/test-none-exist", nil)
	require.Equal(t, distribution.ErrNext, factory{}.Initialize(c, digCon))
}
