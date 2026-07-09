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

package audit_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	repoaudit "github.com/go-sigma/sigma/pkg/dal/repository/audit"
	"github.com/go-sigma/sigma/pkg/testkit"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

func TestNewAuditRepository(t *testing.T) {
	require.NotNil(t, repoaudit.NewAuditRepository())
	require.NotNil(t, repoaudit.NewAuditRepository(query.Q))
}

func TestAuditRepositoryDeleteBefore(t *testing.T) {
	testkit.InitRepository(t)

	ctx := t.Context()
	auditRepository := repoaudit.NewAuditRepository()

	oldAuditID := uuid.NewV7String()
	olderAuditID := uuid.NewV7String()
	newAuditID := uuid.NewV7String()

	for _, auditObj := range []*models.Audit{
		{ID: oldAuditID, UserID: uuid.NewV7String(), Username: "old", Method: "GET", Path: "/old", CreatedAt: 10},
		{ID: olderAuditID, UserID: uuid.NewV7String(), Username: "older", Method: "GET", Path: "/older", CreatedAt: 20},
		{ID: newAuditID, UserID: uuid.NewV7String(), Username: "new", Method: "GET", Path: "/new", CreatedAt: 30},
	} {
		require.NoError(t, auditRepository.Create(ctx, auditObj))
	}

	matched, err := auditRepository.SoftDeleteBefore(ctx, 25, 0)
	require.NoError(t, err)
	require.Zero(t, matched)

	matched, err = auditRepository.SoftDeleteBefore(ctx, 25, 1)
	require.NoError(t, err)
	require.Equal(t, int64(1), matched)

	_, err = query.Q.Audit.WithContext(ctx).Where(query.Q.Audit.ID.Eq(oldAuditID)).First()
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)

	matched, err = auditRepository.SoftDeleteBefore(ctx, 25, 10)
	require.NoError(t, err)
	require.Equal(t, int64(1), matched)

	matched, err = auditRepository.SoftDeleteBefore(ctx, 25, 10)
	require.NoError(t, err)
	require.Zero(t, matched)

	matched, err = auditRepository.HardDeleteBefore(ctx, 25, 10)
	require.NoError(t, err)
	require.Equal(t, int64(2), matched)

	total, err := query.Q.Audit.WithContext(ctx).Unscoped().Count()
	require.NoError(t, err)
	require.Equal(t, int64(1), total)

	auditObj, err := query.Q.Audit.WithContext(ctx).Unscoped().Where(query.Q.Audit.ID.Eq(newAuditID)).First()
	require.NoError(t, err)
	require.Equal(t, "new", auditObj.Username)
}
