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

package overview

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/dal/models"
	repooverview "github.com/go-sigma/sigma/pkg/dal/repository/overview"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	svcanalytics "github.com/go-sigma/sigma/pkg/service/analytics"
)

const (
	// activitySummaryDays is the window of the activity card and its comparison window.
	activitySummaryDays = 30
	// namespaceUsageLimit is the number of namespaces the usage table shows.
	namespaceUsageLimit = 10
	// maxBuildDays caps the window of the build chart.
	maxBuildDays = 90
	// maxActivityDays caps the window of the activity chart.
	maxActivityDays = 400
)

// ErrScopeForbidden is returned when a regular user asks for the global scope.
var ErrScopeForbidden = errors.New("global scope requires an administrator")

//go:generate go tool mockgen -destination=mocks/overview_mocks.go -package=mocks github.com/go-sigma/sigma/pkg/service/overview Service

// Service builds the home dashboard of a user.
type Service interface {
	// GetOverview returns the cards and the namespace table of the dashboard.
	GetOverview(ctx context.Context, user *models.User) (*api.OverviewResponse, error)
	// GetDailyActivity returns the daily pushes and pulls of the requested scope.
	GetDailyActivity(ctx context.Context, user *models.User, days int, scope api.OverviewScope) ([]api.DailyActivity, error)
	// GetDailyBuilds returns the daily build results and their summary of the requested scope.
	GetDailyBuilds(ctx context.Context, user *models.User, days int, scope api.OverviewScope) ([]api.DailyBuildCount, api.BuildSummary, error)
	// ScopeOf returns the dashboard scope of a user.
	ScopeOf(user *models.User) api.OverviewScope
}

type service struct {
	dig.In

	Config       *config.Configuration
	RepoOverview repooverview.OverviewRepository
	AnalyticsSvc svcanalytics.Service
}

// NewService creates the overview service.
func NewService(params service) Service {
	return &params
}

// ScopeOf returns the dashboard scope of a user: administrators see the whole instance.
func (s *service) ScopeOf(user *models.User) api.OverviewScope {
	if user.Role == enums.UserRoleAdmin || user.Role == enums.UserRoleRoot {
		return api.OverviewScopeGlobal
	}
	return api.OverviewScopeSelf
}

// GetOverview returns the cards and the namespace table of the dashboard.
func (s *service) GetOverview(ctx context.Context, user *models.User) (*api.OverviewResponse, error) {
	scope := s.ScopeOf(user)

	var namespaceIDs []string
	if scope == api.OverviewScopeSelf {
		ids, err := s.RepoOverview.ListManagedNamespaceIDs(ctx, user.ID)
		if err != nil {
			return nil, err
		}
		namespaceIDs = ids
	}

	usage, err := s.RepoOverview.AggregateNamespaces(ctx, namespaceIDs)
	if err != nil {
		return nil, err
	}
	// Storage is the sum of the distinct blobs: images share layers, so the quota counters
	// overstate the space the instance actually occupies.
	blobSize, err := s.RepoOverview.SumBlobSize(ctx, namespaceIDs)
	if err != nil {
		return nil, err
	}
	namespaces, err := s.RepoOverview.TopNamespaces(ctx, namespaceIDs, namespaceUsageLimit)
	if err != nil {
		return nil, err
	}
	activity, err := s.activitySummary(ctx, user, scope)
	if err != nil {
		return nil, err
	}

	storage := api.OverviewStorage{Used: blobSize}
	if scope == api.OverviewScopeGlobal {
		// Only administrators have a meaningful total to compare the usage against.
		storage.Capacity = s.Config.Storage.Capacity
	}

	items := make([]api.OverviewNamespace, 0, len(namespaces))
	for _, namespaceObj := range namespaces {
		items = append(items, api.OverviewNamespace{
			ID:           namespaceObj.ID,
			Name:         namespaceObj.Name,
			Visibility:   namespaceObj.Visibility.String(),
			Repositories: namespaceObj.RepositoryCount,
			Tags:         namespaceObj.TagCount,
			Size:         namespaceObj.Size,
			SizeLimit:    namespaceObj.SizeLimit,
		})
	}

	return &api.OverviewResponse{
		Scope:      scope,
		Storage:    storage,
		Counts:     api.OverviewCounts{Namespaces: usage.Namespaces, Repositories: usage.Repositories, Tags: usage.Tags},
		Activity:   activity,
		Namespaces: items,
	}, nil
}

// GetDailyActivity returns the daily pushes and pulls of the requested scope.
func (s *service) GetDailyActivity(ctx context.Context, user *models.User, days int, scope api.OverviewScope) ([]api.DailyActivity, error) {
	scope, err := s.normalizeScope(user, scope)
	if err != nil {
		return nil, err
	}
	if days <= 0 || days > maxActivityDays {
		days = min(max(days, 1), maxActivityDays)
	}
	if scope == api.OverviewScopeGlobal {
		return s.AnalyticsSvc.GetGlobalActivity(ctx, days)
	}
	return s.AnalyticsSvc.GetUserActivity(ctx, user.ID, days)
}

// GetDailyBuilds returns the daily build results and their summary of the requested scope.
func (s *service) GetDailyBuilds(ctx context.Context, user *models.User, days int, scope api.OverviewScope) ([]api.DailyBuildCount, api.BuildSummary, error) {
	scope, err := s.normalizeScope(user, scope)
	if err != nil {
		return nil, api.BuildSummary{}, err
	}
	if days <= 0 || days > maxBuildDays {
		days = min(max(days, 1), maxBuildDays)
	}

	userID := ""
	if scope == api.OverviewScopeSelf {
		userID = user.ID
	}
	now := time.Now().UTC()
	startDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -days+1)
	records, err := s.RepoOverview.ListBuildRecords(ctx, startDay, userID)
	if err != nil {
		return nil, api.BuildSummary{}, err
	}

	success := make(map[string]int64, days)
	failed := make(map[string]int64, days)
	summary := api.BuildSummary{Fastest: math.MaxInt64}
	for _, record := range records {
		date := time.UnixMilli(record.StartedAt).UTC().Format(time.DateOnly)
		switch record.Status {
		case enums.BuildStatusSuccess:
			success[date]++
			summary.Success++
		case enums.BuildStatusFailed:
			failed[date]++
			summary.Failed++
		default:
			continue
		}
		if record.EndedAt == nil || record.StartedAt <= 0 {
			continue
		}
		duration := (ptrValue(record.EndedAt) - record.StartedAt) / 1000
		if duration < 0 {
			continue
		}
		summary.Fastest = min(summary.Fastest, duration)
		summary.Slowest = max(summary.Slowest, duration)
	}
	if summary.Fastest == math.MaxInt64 {
		summary.Fastest = 0
	}
	if total := summary.Success + summary.Failed; total > 0 {
		summary.SuccessRate = round1(float64(summary.Success) / float64(total) * 100)
	}

	items := make([]api.DailyBuildCount, 0, days)
	for dayOffset := range days {
		date := startDay.AddDate(0, 0, dayOffset).Format(time.DateOnly)
		items = append(items, api.DailyBuildCount{Date: date, Success: success[date], Failed: failed[date]})
	}
	return items, summary, nil
}

// activitySummary sums the last 30 days and compares them with the previous 30 days.
func (s *service) activitySummary(ctx context.Context, user *models.User, scope api.OverviewScope) (api.OverviewActivity, error) {
	windowDays := activitySummaryDays * 2
	items, err := s.daily(ctx, user, scope, windowDays)
	if err != nil {
		return api.OverviewActivity{}, err
	}
	split := max(len(items)-activitySummaryDays, 0)
	current := sumActivity(items[split:])
	previous := sumActivity(items[:split])
	return api.OverviewActivity{
		Push:      current.push,
		Pull:      current.pull,
		PushDelta: round1(deltaRatio(current.push, previous.push)),
		PullDelta: round1(deltaRatio(current.pull, previous.pull)),
	}, nil
}

func (s *service) daily(ctx context.Context, user *models.User, scope api.OverviewScope, days int) ([]api.DailyActivity, error) {
	if scope == api.OverviewScopeGlobal {
		return s.AnalyticsSvc.GetGlobalActivity(ctx, days)
	}
	return s.AnalyticsSvc.GetUserActivity(ctx, user.ID, days)
}

// normalizeScope validates the requested scope against the role of the user.
func (s *service) normalizeScope(user *models.User, scope api.OverviewScope) (api.OverviewScope, error) {
	if scope == "" {
		return s.ScopeOf(user), nil
	}
	switch scope {
	case api.OverviewScopeSelf:
		return scope, nil
	case api.OverviewScopeGlobal:
		if s.ScopeOf(user) != api.OverviewScopeGlobal {
			return "", errcode.HTTPErrCodeForbidden.Detail(ErrScopeForbidden.Error())
		}
		return scope, nil
	default:
		return "", errcode.HTTPErrCodeBadRequest.Detail(fmt.Sprintf("invalid scope: %s", scope))
	}
}

type activityTotal struct {
	push int64
	pull int64
}

func sumActivity(items []api.DailyActivity) activityTotal {
	total := activityTotal{}
	for _, item := range items {
		total.push += item.PushCount
		total.pull += item.PullCount
	}
	return total
}

// deltaRatio returns the percentage change between two periods. A zero base yields 100% when
// the current period is not empty, which is how a new activity reads on a dashboard.
func deltaRatio(current, previous int64) float64 {
	if previous == 0 {
		if current == 0 {
			return 0
		}
		return 100
	}
	return float64(current-previous) / float64(previous) * 100
}

func round1(value float64) float64 {
	return math.Round(value*10) / 10
}

func ptrValue(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}
