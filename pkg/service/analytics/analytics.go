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

package analytics

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/consts"
	repoanalytics "github.com/go-sigma/sigma/pkg/dal/repository/analytics"
	"github.com/go-sigma/sigma/pkg/infra/counter"
	"github.com/go-sigma/sigma/pkg/infra/lock"
)

const (
	defaultRetentionDays     = 400
	redisKeyPrefix           = "sigma:analytics"
	redisCounterBackend      = "redis"
	analyticsLockExpire      = 30 * time.Second
	analyticsLockWaitTimeout = 100 * time.Millisecond
)

//go:generate mockgen -destination=analytics_mocks.go -package=analytics github.com/go-sigma/sigma/pkg/service/analytics Service

// PushEvent describes a successful manifest push.
type PushEvent struct {
	UserID       string
	NamespaceID  string
	RepositoryID string
	ArtifactID   string
	Digest       string
}

// PullEvent describes a successful manifest pull.
type PullEvent struct {
	UserID       string
	NamespaceID  string
	RepositoryID string
	ArtifactID   string
	Reference    string
}

// Service records and queries backend analytics rollups.
type Service interface {
	RecordPush(ctx context.Context, event PushEvent) error
	RecordPull(ctx context.Context, event PullEvent) error
	RecordNamespaceSizeDelta(ctx context.Context, namespaceID string, delta int64) error
	RecordNamespaceTagDelta(ctx context.Context, namespaceID string, delta int64) error
	Flush(ctx context.Context) error
	GetUserPushHeatmap(ctx context.Context, userID string, days int) ([]api.DailyCount, error)
	GetNamespaceTrends(ctx context.Context, namespaceID string, days int) ([]api.NamespaceHourlyMetric, error)
}

type service struct {
	dig.In

	Config        *config.Configuration
	RepoAnalytics repoanalytics.AnalyticsRepository
	Counter       counter.Counter
	Locker        lock.Locker
}

// NewService registers the analytics service in the dependency container.
func NewService(digCon *dig.Container) error {
	return digCon.Provide(func(params service) Service {
		params.Config = withDefaults(params.Config)
		return &params
	})
}

func withDefaults(config *config.Configuration) *config.Configuration {
	if config.Analytics.RetentionDays == 0 {
		config.Analytics.RetentionDays = defaultRetentionDays
	}
	return config
}

func (s *service) RecordPush(ctx context.Context, event PushEvent) error {
	if !s.Config.Analytics.Enabled || event.UserID == "" || event.NamespaceID == "" {
		return nil
	}
	hour := hourValue(time.Now().UTC())
	if event.UserID != "" {
		if err := s.Counter.HIncrBy(ctx, userPushKey(hour), map[string]int64{event.UserID: 1}); err != nil {
			return err
		}
	}
	if err := s.Counter.HIncrBy(ctx, namespaceKey(hour), map[string]int64{
		event.NamespaceID + ":push": 1,
	}); err != nil {
		return err
	}
	return s.Counter.SAdd(ctx, dirtyHoursKey(), strconv.FormatInt(hour, 10))
}

func (s *service) RecordPull(ctx context.Context, event PullEvent) error {
	if !s.Config.Analytics.Enabled || event.NamespaceID == "" {
		return nil
	}
	hour := hourValue(time.Now().UTC())
	if err := s.Counter.HIncrBy(ctx, namespaceKey(hour), map[string]int64{
		event.NamespaceID + ":pull": 1,
	}); err != nil {
		return err
	}
	return s.Counter.SAdd(ctx, dirtyHoursKey(), strconv.FormatInt(hour, 10))
}

func (s *service) RecordNamespaceSizeDelta(ctx context.Context, namespaceID string, delta int64) error {
	if !s.Config.Analytics.Enabled || namespaceID == "" || delta == 0 {
		return nil
	}
	hour := hourValue(time.Now().UTC())
	if err := s.Counter.HIncrBy(ctx, namespaceKey(hour), map[string]int64{
		namespaceID + ":size_delta": delta,
	}); err != nil {
		return err
	}
	return s.Counter.SAdd(ctx, dirtyHoursKey(), strconv.FormatInt(hour, 10))
}

func (s *service) RecordNamespaceTagDelta(ctx context.Context, namespaceID string, delta int64) error {
	if !s.Config.Analytics.Enabled || namespaceID == "" || delta == 0 {
		return nil
	}
	hour := hourValue(time.Now().UTC())
	if err := s.Counter.HIncrBy(ctx, namespaceKey(hour), map[string]int64{
		namespaceID + ":tag_delta": delta,
	}); err != nil {
		return err
	}
	return s.Counter.SAdd(ctx, dirtyHoursKey(), strconv.FormatInt(hour, 10))
}

func (s *service) Flush(ctx context.Context) error {
	if !s.Config.Analytics.Enabled {
		return nil
	}
	dirtyHours, err := s.Counter.SMembers(ctx, dirtyHoursKey())
	if err != nil {
		return err
	}
	for _, dirtyHour := range dirtyHours {
		hour, err := strconv.ParseInt(dirtyHour, 10, 64)
		if err != nil {
			slog.Warn("parse analytics dirty hour failed", "hour", dirtyHour, "err", err)
			continue
		}
		if err := s.flushDirtyHour(ctx, hour); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) flushDirtyHour(ctx context.Context, hour int64) error {
	if s.Config.Analytics.CounterBackend != redisCounterBackend {
		return s.flushHour(ctx, hour)
	}

	lockCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	key := analyticsLockKey(hour)
	if err := s.Locker.AcquireWithRenew(lockCtx, key, analyticsLockExpire, analyticsLockWaitTimeout); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			slog.Info("skip analytics flush because another runner holds the lock", "hour", hour)
			return nil
		}
		return fmt.Errorf("acquire analytics flush lock: %w", err)
	}
	return s.flushHour(lockCtx, hour)
}

func (s *service) GetUserPushHeatmap(ctx context.Context, userID string, days int) ([]api.DailyCount, error) {
	if days <= 0 || days > s.Config.Analytics.RetentionDays {
		days = min(max(days, 1), s.Config.Analytics.RetentionDays)
	}
	now := time.Now().UTC()
	startDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -days+1)
	startHour := hourValue(startDay)
	endHour := hourValue(now)
	rows, err := s.RepoAnalytics.ListUserPush(ctx, userID, startHour, endHour)
	if err != nil {
		return nil, err
	}
	counts := make(map[string]int64, len(rows))
	for _, row := range rows {
		rowTime, err := timeFromHour(row.Hour)
		if err != nil {
			return nil, err
		}
		counts[rowTime.Format(time.DateOnly)] += row.PushCount
	}
	items := make([]api.DailyCount, 0, days)
	for dayOffset := range days {
		date := startDay.AddDate(0, 0, dayOffset).Format(time.DateOnly)
		items = append(items, api.DailyCount{Date: date, PushCount: counts[date]})
	}
	return items, nil
}

func (s *service) GetNamespaceTrends(
	ctx context.Context,
	namespaceID string,
	days int,
) ([]api.NamespaceHourlyMetric, error) {

	if days <= 0 || days > 30 {
		days = min(max(days, 1), 30)
	}
	now := time.Now().UTC().Truncate(time.Hour)
	start := now.Add(-time.Duration(days*24-1) * time.Hour)
	startHour := hourValue(start)
	endHour := hourValue(now)
	rows, err := s.RepoAnalytics.ListNamespaceActivity(ctx, namespaceID, startHour, endHour)
	if err != nil {
		return nil, err
	}
	byHour := make(map[int64]api.NamespaceHourlyMetric, len(rows))
	for _, row := range rows {
		rowTime, err := timeFromHour(row.Hour)
		if err != nil {
			return nil, err
		}
		byHour[row.Hour] = api.NamespaceHourlyMetric{
			Hour:      rowTime.Format(time.RFC3339),
			PushCount: row.PushCount,
			PullCount: row.PullCount,
			SizeDelta: row.SizeDelta,
			TagDelta:  row.TagDelta,
		}
	}
	items := make([]api.NamespaceHourlyMetric, 0, days*24)
	for offset := range days * 24 {
		pointTime := start.Add(time.Duration(offset) * time.Hour)
		hour := hourValue(pointTime)
		item, ok := byHour[hour]
		if !ok {
			item.Hour = pointTime.Format(time.RFC3339)
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *service) flushHour(ctx context.Context, hour int64) error {
	userKey := userPushKey(hour)
	namespaceKey := namespaceKey(hour)
	userFlushingKey := flushingKey(userKey)
	namespaceFlushingKey := flushingKey(namespaceKey)
	if err := s.Counter.Rename(ctx, userKey, userFlushingKey); err != nil {
		return err
	}
	if err := s.Counter.Rename(ctx, namespaceKey, namespaceFlushingKey); err != nil {
		return err
	}
	userCounts, err := s.Counter.HGetAll(ctx, userFlushingKey)
	if err != nil {
		return err
	}
	namespaceCounts, err := s.Counter.HGetAll(ctx, namespaceFlushingKey)
	if err != nil {
		return err
	}
	for userID, count := range userCounts {
		if incrErr := s.RepoAnalytics.IncrUserPush(ctx, hour, userID, count); incrErr != nil {
			return incrErr
		}
	}
	namespaceDeltas := parseNamespaceDeltas(namespaceCounts)
	for namespaceID, delta := range namespaceDeltas {
		if err := s.RepoAnalytics.IncrNamespace(
			ctx,
			hour,
			namespaceID,
			delta.push,
			delta.pull,
			delta.size,
			delta.tag,
		); err != nil {
			return err
		}
	}
	if err := s.Counter.Del(ctx, userFlushingKey, namespaceFlushingKey); err != nil {
		return err
	}
	return s.Counter.SRem(ctx, dirtyHoursKey(), strconv.FormatInt(hour, 10))
}

type namespaceDelta struct {
	push int64
	pull int64
	size int64
	tag  int64
}

func parseNamespaceDeltas(values map[string]int64) map[string]namespaceDelta {
	result := make(map[string]namespaceDelta, len(values))
	for field, value := range values {
		namespaceID, metric, ok := strings.Cut(field, ":")
		if !ok {
			continue
		}
		delta := result[namespaceID]
		switch metric {
		case "push":
			delta.push = value
		case "pull":
			delta.pull = value
		case "size_delta":
			delta.size = value
		case "tag_delta":
			delta.tag = value
		}
		result[namespaceID] = delta
	}
	return result
}

func hourValue(t time.Time) int64 {
	year, month, day := t.UTC().Date()
	return int64(year)*1000000 + int64(month)*10000 + int64(day)*100 + int64(t.UTC().Hour())
}

func timeFromHour(hour int64) (time.Time, error) {
	hourString := strconv.FormatInt(hour, 10)
	if len(hourString) != 10 {
		return time.Time{}, fmt.Errorf("invalid hour value %d", hour)
	}
	return time.ParseInLocation("2006010215", hourString, time.UTC)
}

func userPushKey(hour int64) string {
	return fmt.Sprintf("%s:hour:%d:user_push", redisKeyPrefix, hour)
}

func namespaceKey(hour int64) string {
	return fmt.Sprintf("%s:hour:%d:namespace", redisKeyPrefix, hour)
}

func flushingKey(key string) string {
	return key + ":flushing"
}

func dirtyHoursKey() string {
	return redisKeyPrefix + ":dirty_hours"
}

func analyticsLockKey(hour int64) string {
	return fmt.Sprintf("%s:%d", consts.LockerAnalyticsFlush, hour)
}
