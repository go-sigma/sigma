// Copyright 2026 sigma
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

package audit_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	repoaudit "github.com/go-sigma/sigma/pkg/dal/repository/audit"
	mwaudit "github.com/go-sigma/sigma/pkg/server/middlewares/audit"
	"github.com/go-sigma/sigma/pkg/testkit"
)

var _ repoaudit.AuditRepository = (*recordingAuditRepository)(nil)

type recordingAuditRepository struct {
	mu      sync.Mutex
	records []*models.Audit
	err     error
}

func (r *recordingAuditRepository) Create(_ context.Context, audit *models.Audit) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.records = append(r.records, audit)
	return r.err
}

func (r *recordingAuditRepository) SoftDeleteBefore(_ context.Context, _ int64, _ int) (int64, error) {
	return 0, nil
}

func (r *recordingAuditRepository) HardDeleteBefore(_ context.Context, _ int64, _ int) (int64, error) {
	return 0, nil
}

func (r *recordingAuditRepository) all() []*models.Audit {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]*models.Audit(nil), r.records...)
}

func newRouter(repo *recordingAuditRepository, recordListGet bool) *gin.Engine {
	router := testkit.NewGin()
	router.Use(func(c *gin.Context) {
		c.Set(consts.ContextUser, &models.User{ID: "user-1", Username: "sigma", Role: enums.UserRoleAdmin})
		c.Next()
	})
	router.Use(mwaudit.AuditWithConfig(mwaudit.Config{
		RecordListGet:   recordListGet,
		AuditRepository: repo,
	}))
	router.GET(consts.APIV1+"/namespaces/:namespace_id", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	router.GET(consts.APIV1+"/namespaces/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"items": []string{}})
	})
	router.GET(consts.APIV1+"/fail", func(c *gin.Context) {
		c.Status(http.StatusInternalServerError)
	})
	router.GET("/healthz", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	return router
}

func TestAuditRecordsCompletedRequest(t *testing.T) {
	repo := &recordingAuditRepository{}
	router := newRouter(repo, false)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, consts.APIV1+"/namespaces/ns-1?q=1", nil)
	req.Header.Set("User-Agent", "audit-test")

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
	records := repo.all()
	require.Len(t, records, 1)
	require.Equal(t, "user-1", records[0].UserID)
	require.Equal(t, "sigma", records[0].Username)
	require.Equal(t, enums.UserRoleAdmin.String(), records[0].UserRole)
	require.Equal(t, http.MethodGet, records[0].Method)
	require.Equal(t, consts.APIV1+"/namespaces/ns-1", records[0].Path)
	require.Equal(t, consts.APIV1+"/namespaces/:namespace_id", records[0].Route)
	require.Equal(t, "q=1", records[0].Query)
	require.Equal(t, http.StatusNoContent, records[0].StatusCode)
	require.Equal(t, "audit-test", records[0].UserAgent)
	require.NotNil(t, records[0].NamespaceID)
	require.Equal(t, "ns-1", *records[0].NamespaceID)
}

func TestAuditRecordsErrorStatus(t *testing.T) {
	repo := &recordingAuditRepository{}
	router := newRouter(repo, false)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, consts.APIV1+"/fail", nil)

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	records := repo.all()
	require.Len(t, records, 1)
	require.Equal(t, http.StatusInternalServerError, records[0].StatusCode)
}

func TestAuditSkipsListGetByDefault(t *testing.T) {
	repo := &recordingAuditRepository{}
	router := newRouter(repo, false)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, consts.APIV1+"/namespaces/", nil)

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Empty(t, repo.all())
}

func TestAuditRecordsListGetWhenEnabled(t *testing.T) {
	repo := &recordingAuditRepository{}
	router := newRouter(repo, true)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, consts.APIV1+"/namespaces/", nil)

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Len(t, repo.all(), 1)
}

func TestAuditAlwaysSkipsProbePath(t *testing.T) {
	repo := &recordingAuditRepository{}
	router := newRouter(repo, true)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Empty(t, repo.all())
}

func TestAuditSkipsWithoutUser(t *testing.T) {
	repo := &recordingAuditRepository{}
	router := testkit.NewGin()
	router.Use(mwaudit.AuditWithConfig(mwaudit.Config{AuditRepository: repo}))
	router.GET(consts.APIV1+"/namespaces/:namespace_id", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, consts.APIV1+"/namespaces/ns-1", nil)

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
	require.Empty(t, repo.all())
}

func TestAuditCreateErrorDoesNotChangeResponse(t *testing.T) {
	repo := &recordingAuditRepository{err: errors.New("insert failed")}
	router := newRouter(repo, false)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, consts.APIV1+"/namespaces/ns-1", nil)

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
	require.Len(t, repo.all(), 1)
}
