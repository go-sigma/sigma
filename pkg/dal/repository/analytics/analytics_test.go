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

package analytics_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/go-sigma/sigma/pkg/dal/query"
	repoanalytics "github.com/go-sigma/sigma/pkg/dal/repository/analytics"
	"github.com/go-sigma/sigma/pkg/testkit"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

func TestNewAnalyticsRepository(t *testing.T) {
	require.NotNil(t, repoanalytics.NewAnalyticsRepository())
	require.NotNil(t, repoanalytics.NewAnalyticsRepository(query.Q))
}

func TestAnalyticsRepository(t *testing.T) {
	testkit.InitRepository(t)

	ctx := t.Context()
	analyticsRepository := repoanalytics.NewAnalyticsRepository()

	userID := uuid.NewV7String()
	otherUserID := uuid.NewV7String()
	namespaceID := uuid.NewV7String()
	otherNamespaceID := uuid.NewV7String()

	require.NoError(t, analyticsRepository.IncrUserPush(ctx, 10, userID, 2))
	require.NoError(t, analyticsRepository.IncrUserPush(ctx, 10, userID, 3))
	require.NoError(t, analyticsRepository.IncrUserPush(ctx, 11, userID, 5))
	require.NoError(t, analyticsRepository.IncrUserPush(ctx, 10, otherUserID, 7))

	userRollups, err := analyticsRepository.ListUserPush(ctx, userID, 9, 10)
	require.NoError(t, err)
	require.Len(t, userRollups, 1)
	require.Equal(t, int64(10), userRollups[0].Hour)
	require.Equal(t, int64(5), userRollups[0].PushCount)

	userRollups, err = analyticsRepository.ListUserPush(ctx, userID, 10, 11)
	require.NoError(t, err)
	require.Len(t, userRollups, 2)
	require.Equal(t, []int64{10, 11}, []int64{userRollups[0].Hour, userRollups[1].Hour})

	require.NoError(t, analyticsRepository.IncrNamespace(ctx, 20, namespaceID, 1, 2, 3, 4))
	require.NoError(t, analyticsRepository.IncrNamespace(ctx, 20, namespaceID, 5, 6, 7, 8))
	require.NoError(t, analyticsRepository.IncrNamespace(ctx, 21, namespaceID, 9, 10, 11, 12))
	require.NoError(t, analyticsRepository.IncrNamespace(ctx, 20, otherNamespaceID, 13, 14, 15, 16))

	namespaceRollups, err := analyticsRepository.ListNamespaceActivity(ctx, namespaceID, 19, 20)
	require.NoError(t, err)
	require.Len(t, namespaceRollups, 1)
	require.Equal(t, int64(20), namespaceRollups[0].Hour)
	require.Equal(t, int64(6), namespaceRollups[0].PushCount)
	require.Equal(t, int64(8), namespaceRollups[0].PullCount)
	require.Equal(t, int64(10), namespaceRollups[0].SizeDelta)
	require.Equal(t, int64(12), namespaceRollups[0].TagDelta)

	namespaceRollups, err = analyticsRepository.ListNamespaceActivity(ctx, namespaceID, 20, 21)
	require.NoError(t, err)
	require.Len(t, namespaceRollups, 2)
	require.Equal(t, []int64{20, 21}, []int64{namespaceRollups[0].Hour, namespaceRollups[1].Hour})

	require.NoError(t, analyticsRepository.DeleteBeforeHour(ctx, 11))

	userRollups, err = analyticsRepository.ListUserPush(ctx, userID, 9, 12)
	require.NoError(t, err)
	require.Len(t, userRollups, 1)
	require.Equal(t, int64(11), userRollups[0].Hour)

	namespaceRollups, err = analyticsRepository.ListNamespaceActivity(ctx, namespaceID, 19, 22)
	require.NoError(t, err)
	require.Len(t, namespaceRollups, 2)
	require.Equal(t, []int64{20, 21}, []int64{namespaceRollups[0].Hour, namespaceRollups[1].Hour})
}
