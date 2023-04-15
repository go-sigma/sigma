// Copyright 2023 XImager
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

package tag

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"

	"github.com/ximager/ximager/pkg/dal"
	"github.com/ximager/ximager/pkg/dal/dao"
	"github.com/ximager/ximager/pkg/dal/models"
	"github.com/ximager/ximager/pkg/dal/query"
	"github.com/ximager/ximager/pkg/logger"
	"github.com/ximager/ximager/pkg/tests"
	"github.com/ximager/ximager/pkg/validators"
)

func TestDeleteTag(t *testing.T) {
	logger.SetLevel("debug")
	e := echo.New()
	validators.Initialize(e)
	err := tests.Initialize()
	assert.NoError(t, err)
	err = tests.DB.Init()
	assert.NoError(t, err)
	defer func() {
		conn, err := dal.DB.DB()
		assert.NoError(t, err)
		err = conn.Close()
		assert.NoError(t, err)
		err = tests.DB.DeInit()
		assert.NoError(t, err)
	}()

	ctx := context.Background()

	const (
		namespaceName  = "test"
		repositoryName = "busybox"
	)

	var tagObj *models.Tag
	err = query.Q.Transaction(func(tx *query.Query) error {
		namespaceService := dao.NewNamespaceService(tx)
		namespaceObj, err := namespaceService.Create(ctx, &models.Namespace{Name: namespaceName})
		assert.NoError(t, err)
		log.Info().Interface("namespace", namespaceObj).Msg("namespace created")
		repositoryService := dao.NewRepositoryService(tx)
		repositoryObj, err := repositoryService.Create(ctx, &models.Repository{Name: repositoryName, NamespaceID: namespaceObj.ID})
		assert.NoError(t, err)
		artifactService := dao.NewArtifactService(tx)
		artifactObj, err := artifactService.Save(ctx, &models.Artifact{RepositoryID: repositoryObj.ID, Digest: "sha256:1234567890", Size: 1234, ContentType: "application/octet-stream", Raw: "test", PushedAt: time.Now()})
		assert.NoError(t, err)
		tagService := dao.NewTagService(tx)
		tagObj, err = tagService.Save(ctx, &models.Tag{Name: "latest", RepositoryID: repositoryObj.ID, ArtifactID: artifactObj.ID, PushedAt: time.Now()})
		assert.NoError(t, err)
		return nil
	})
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	q := req.URL.Query()
	q.Add("repository", repositoryName)
	req.URL.RawQuery = q.Encode()
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.SetPath("/namespace/:namespace/tag/:id")
	c.SetParamNames("namespace", "id")
	c.SetParamValues(namespaceName, strconv.FormatUint(tagObj.ID, 10))

	tagHandler := New()
	if assert.NoError(t, tagHandler.DeleteTag(c)) {
		assert.Equal(t, http.StatusNoContent, rec.Code)
	}
}
