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

package registry_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/go-sigma/sigma/pkg/dal/models"
	reporegistry "github.com/go-sigma/sigma/pkg/dal/repository/registry"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

func TestQuotaUsageRepositoryReserveAndRelease(t *testing.T) {
	_ = newRegistryFixture(t)
	ctx := t.Context()
	quotaUsageRepository := reporegistry.NewQuotaUsageRepository()

	require.NoError(t, quotaUsageRepository.ReserveGlobal(ctx, 60, 100))
	used, err := quotaUsageRepository.GetGlobal(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(60), used)

	// 60 + 50 > 100: the reservation must be rejected with 507.
	err = quotaUsageRepository.ReserveGlobal(ctx, 50, 100)
	require.Error(t, err)
	code, ok := errcode.AsType[errcode.ErrCode](err)
	require.True(t, ok)
	require.Equal(t, http.StatusInsufficientStorage, code.HTTPStatusCode)

	used, err = quotaUsageRepository.GetGlobal(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(60), used)

	require.NoError(t, quotaUsageRepository.ReleaseGlobal(ctx, 60))
	used, err = quotaUsageRepository.GetGlobal(ctx)
	require.NoError(t, err)
	require.Zero(t, used)
}

func TestQuotaUsageRepositorySyncGlobal(t *testing.T) {
	fixture := newRegistryFixture(t)
	ctx := t.Context()
	quotaUsageRepository := reporegistry.NewQuotaUsageRepository()

	blobOne := &models.Blob{
		ID:          uuid.NewV7String(),
		Digest:      "sha256:quota-usage-one",
		Size:        120,
		ContentType: "application/octet-stream",
	}
	blobTwo := &models.Blob{
		ID:          uuid.NewV7String(),
		Digest:      "sha256:quota-usage-two",
		Size:        30,
		ContentType: "application/octet-stream",
	}
	require.NoError(t, fixture.BlobRepository.Create(ctx, blobOne))
	require.NoError(t, fixture.BlobRepository.Create(ctx, blobTwo))

	used, err := quotaUsageRepository.SyncGlobal(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(150), used)

	got, err := quotaUsageRepository.GetGlobal(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(150), got)
}
