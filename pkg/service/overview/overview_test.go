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

package overview_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/dig"
	"go.uber.org/mock/gomock"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/dal/models"
	repooverview "github.com/go-sigma/sigma/pkg/dal/repository/overview"
	repomocks "github.com/go-sigma/sigma/pkg/dal/repository/overview/mocks"
	svcanalytics "github.com/go-sigma/sigma/pkg/service/analytics"
	analyticsmocks "github.com/go-sigma/sigma/pkg/service/analytics/mocks"
	svcoverview "github.com/go-sigma/sigma/pkg/service/overview"
)

const testCapacity = int64(100 * 1024 * 1024)

func newTestService(t *testing.T, repo repooverview.OverviewRepository, analytics svcanalytics.Service) svcoverview.Service {
	t.Helper()

	digCon := dig.New()
	require.NoError(t, digCon.Provide(func() *config.Configuration {
		return &config.Configuration{Storage: config.ConfigurationStorage{Capacity: testCapacity}}
	}))
	require.NoError(t, digCon.Provide(func() repooverview.OverviewRepository { return repo }))
	require.NoError(t, digCon.Provide(func() svcanalytics.Service { return analytics }))
	require.NoError(t, digCon.Provide(svcoverview.NewService))

	var svc svcoverview.Service
	require.NoError(t, digCon.Invoke(func(service svcoverview.Service) { svc = service }))
	return svc
}

func TestGetOverviewGlobal(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := repomocks.NewMockOverviewRepository(ctrl)
	analyticsSvc := analyticsmocks.NewMockService(ctrl)
	svc := newTestService(t, repo, analyticsSvc)

	user := &models.User{ID: "admin", Role: enums.UserRoleAdmin}
	ctx := t.Context()

	repo.EXPECT().AggregateNamespaces(ctx, nil).Return(&api.NamespaceUsage{
		Namespaces: 2, Repositories: 5, Tags: 11, Size: 4096,
	}, nil)
	repo.EXPECT().SumBlobSize(ctx, nil).Return(int64(2048), nil)
	repo.EXPECT().TopNamespaces(ctx, nil, 10).Return([]*models.Namespace{{
		ID: "ns1", Name: "library", Visibility: enums.VisibilityPublic,
		RepositoryCount: 5, TagCount: 11, Size: 4096, SizeLimit: 8192,
	}}, nil)
	analyticsSvc.EXPECT().GetGlobalActivity(ctx, 60).Return(activityItems(60, 1, 2), nil)

	resp, err := svc.GetOverview(ctx, user)
	require.NoError(t, err)
	require.Equal(t, api.OverviewScopeGlobal, resp.Scope)
	require.Equal(t, testCapacity, resp.Storage.Capacity)
	// The card shows the distinct blob size, not the sum of the namespace quotas.
	require.Equal(t, int64(2048), resp.Storage.Used)
	require.Equal(t, api.OverviewCounts{Namespaces: 2, Repositories: 5, Tags: 11}, resp.Counts)
	require.Len(t, resp.Namespaces, 1)
	require.Equal(t, "library", resp.Namespaces[0].Name)
	require.Equal(t, int64(8192), resp.Namespaces[0].SizeLimit)

	// Current window doubles the previous one, so the deltas are +100%.
	require.Equal(t, int64(60), resp.Activity.Push)
	require.Equal(t, int64(120), resp.Activity.Pull)
	require.InDelta(t, 100, resp.Activity.PushDelta, 0.01)
	require.InDelta(t, 100, resp.Activity.PullDelta, 0.01)
}

func TestGetOverviewSelf(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := repomocks.NewMockOverviewRepository(ctrl)
	analyticsSvc := analyticsmocks.NewMockService(ctrl)
	svc := newTestService(t, repo, analyticsSvc)

	user := &models.User{ID: "user1", Role: enums.UserRoleUser}
	ctx := t.Context()
	managed := []string{"ns1"}

	repo.EXPECT().ListManagedNamespaceIDs(ctx, user.ID).Return(managed, nil)
	repo.EXPECT().AggregateNamespaces(ctx, managed).Return(&api.NamespaceUsage{Namespaces: 1, Repositories: 2, Tags: 3, Size: 100}, nil)
	repo.EXPECT().SumBlobSize(ctx, managed).Return(int64(80), nil)
	repo.EXPECT().TopNamespaces(ctx, managed, 10).Return([]*models.Namespace{{
		ID: "ns1", Name: "demo", Visibility: enums.VisibilityPrivate, RepositoryCount: 2, TagCount: 3, Size: 100,
	}}, nil)
	analyticsSvc.EXPECT().GetUserActivity(ctx, user.ID, 60).Return(activityItems(60, 0, 0), nil)

	resp, err := svc.GetOverview(ctx, user)
	require.NoError(t, err)
	require.Equal(t, api.OverviewScopeSelf, resp.Scope)
	// A regular user has no total to compare against, so the capacity stays unset.
	require.Zero(t, resp.Storage.Capacity)
	require.Equal(t, int64(80), resp.Storage.Used)
	require.Equal(t, int64(1), resp.Counts.Namespaces)
	require.InDelta(t, 0, resp.Activity.PushDelta, 0.01)
}

func TestGetDailyActivityScope(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := repomocks.NewMockOverviewRepository(ctrl)
	analyticsSvc := analyticsmocks.NewMockService(ctrl)
	svc := newTestService(t, repo, analyticsSvc)

	user := &models.User{ID: "user1", Role: enums.UserRoleUser}
	ctx := t.Context()

	analyticsSvc.EXPECT().GetUserActivity(ctx, user.ID, 7).Return(activityItems(7, 1, 2), nil)
	items, err := svc.GetDailyActivity(ctx, user, 7, api.OverviewScopeSelf)
	require.NoError(t, err)
	require.Len(t, items, 7)

	_, err = svc.GetDailyActivity(ctx, user, 7, api.OverviewScopeGlobal)
	require.Error(t, err)

	_, err = svc.GetDailyActivity(ctx, user, 7, api.OverviewScope("instance"))
	require.Error(t, err)
}

func TestGetDailyBuilds(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := repomocks.NewMockOverviewRepository(ctrl)
	analyticsSvc := analyticsmocks.NewMockService(ctrl)
	svc := newTestService(t, repo, analyticsSvc)

	user := &models.User{ID: "user1", Role: enums.UserRoleUser}
	ctx := t.Context()

	now := time.Now().UTC()
	started := now.Add(-time.Hour).UnixMilli()
	fastEnd := now.Add(-time.Hour + 90*time.Second).UnixMilli()
	slowEnd := now.Add(-time.Hour + 300*time.Second).UnixMilli()
	repo.EXPECT().ListBuildRecords(ctx, gomock.Any(), user.ID).Return([]repooverview.BuildRecord{
		{StartedAt: started, EndedAt: &fastEnd, Status: enums.BuildStatusSuccess},
		{StartedAt: started, EndedAt: &slowEnd, Status: enums.BuildStatusSuccess},
		{StartedAt: started, Status: enums.BuildStatusFailed},
		{StartedAt: started, Status: enums.BuildStatusBuilding},
	}, nil)

	items, summary, err := svc.GetDailyBuilds(ctx, user, 7, api.OverviewScopeSelf)
	require.NoError(t, err)
	require.Len(t, items, 7)
	require.Equal(t, int64(2), summary.Success)
	require.Equal(t, int64(1), summary.Failed)
	require.InDelta(t, 66.7, summary.SuccessRate, 0.01)
	require.Equal(t, int64(90), summary.Fastest)
	require.Equal(t, int64(300), summary.Slowest)
}

// activityItems builds a window where the current half is multiplied by currentFactor.
func activityItems(days, previousValue, currentValue int64) []api.DailyActivity {
	items := make([]api.DailyActivity, 0, days)
	half := days / 2
	for index := range days {
		value := previousValue
		if index >= half {
			value = currentValue
		}
		items = append(items, api.DailyActivity{Date: "2026-01-01", PushCount: value, PullCount: value * 2})
	}
	return items
}
