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
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/dal/models"
	repoanalytics "github.com/go-sigma/sigma/pkg/dal/repository/analytics"
	"github.com/go-sigma/sigma/pkg/infra/counter"
	"github.com/go-sigma/sigma/pkg/infra/lock"
)

func TestRecordPushAndFlush(t *testing.T) {
	repo := newFakeAnalyticsRepository()
	locker := &fakeLocker{}
	svc := newTestService(t, repo, "", locker)

	err := svc.RecordPush(t.Context(), PushEvent{
		UserID:       "1",
		NamespaceID:  "2",
		RepositoryID: "3",
		ArtifactID:   "4",
		Digest:       "sha256:test",
	})
	require.NoError(t, err)

	err = svc.Flush(t.Context())
	require.NoError(t, err)

	require.Len(t, repo.userRows, 1)
	require.Len(t, repo.namespaceRows, 1)
	for _, row := range repo.userRows {
		require.Equal(t, int64(1), row.PushCount)
	}
	for _, row := range repo.namespaceRows {
		require.Equal(t, int64(1), row.PushCount)
		require.Zero(t, row.PullCount)
	}
	require.Empty(t, locker.keys)
}

func TestFlushRedisUsesHourLock(t *testing.T) {
	repo := newFakeAnalyticsRepository()
	locker := &fakeLocker{}
	svc := newTestService(t, repo, redisCounterBackend, locker)

	require.NoError(t, svc.RecordPush(t.Context(), PushEvent{
		UserID:      "1",
		NamespaceID: "2",
	}))
	require.NoError(t, svc.Flush(t.Context()))

	require.Equal(t, []string{analyticsLockKey(hourValue(time.Now().UTC()))}, locker.keys)
}

func TestFlushRedisSkipsLockedHourAndContinues(t *testing.T) {
	repo := newFakeAnalyticsRepository()
	firstHour := int64(2026072701)
	secondHour := int64(2026072702)
	locker := &fakeLocker{
		errs: map[string]error{
			analyticsLockKey(firstHour): context.DeadlineExceeded,
		},
	}
	svc := newTestService(t, repo, redisCounterBackend, locker)
	ctr := svc.ctr.(*fakeCounter)
	require.NoError(t, ctr.HIncrBy(t.Context(), userPushKey(firstHour), map[string]int64{"first": 1}))
	require.NoError(t, ctr.HIncrBy(t.Context(), userPushKey(secondHour), map[string]int64{"second": 2}))
	require.NoError(t, ctr.SAdd(t.Context(), dirtyHoursKey(), "2026072701", "2026072702"))

	require.NoError(t, svc.Flush(t.Context()))

	require.NotContains(t, repo.userRows, firstHour)
	require.Equal(t, int64(2), repo.userRows[secondHour].PushCount)
	require.ElementsMatch(t, []string{analyticsLockKey(firstHour), analyticsLockKey(secondHour)}, locker.keys)
}

func TestFlushRedisReturnsLockError(t *testing.T) {
	repo := newFakeAnalyticsRepository()
	hour := int64(2026072701)
	lockErr := errors.New("locker unavailable")
	locker := &fakeLocker{errs: map[string]error{analyticsLockKey(hour): lockErr}}
	svc := newTestService(t, repo, redisCounterBackend, locker)
	require.NoError(t, svc.ctr.SAdd(t.Context(), dirtyHoursKey(), "2026072701"))

	err := svc.Flush(t.Context())

	require.ErrorIs(t, err, lockErr)
}

func TestGetUserPushHeatmapFillsMissingDays(t *testing.T) {
	repo := newFakeAnalyticsRepository()
	svc := newTestService(t, repo, "", &fakeLocker{})
	today := time.Now().UTC()
	startDay := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -2)
	repo.userRows[hourValue(startDay)] = &models.UserActivityHourly{Hour: hourValue(startDay), UserID: "1", PushCount: 3}
	repo.userRows[hourValue(startDay.AddDate(0, 0, 2))] = &models.UserActivityHourly{Hour: hourValue(startDay.AddDate(0, 0, 2)), UserID: "1", PushCount: 5}

	items, err := svc.GetUserPushHeatmap(t.Context(), "1", 3)
	require.NoError(t, err)
	require.Len(t, items, 3)
	require.Equal(t, int64(3), items[0].PushCount)
	require.Zero(t, items[1].PushCount)
	require.Equal(t, int64(5), items[2].PushCount)
}

func TestGetNamespaceTrendsFillsMissingHours(t *testing.T) {
	repo := newFakeAnalyticsRepository()
	svc := newTestService(t, repo, "", &fakeLocker{})
	now := time.Now().UTC().Truncate(time.Hour)
	repo.namespaceRows[hourValue(now)] = &models.NamespaceActivityHourly{
		Hour:        hourValue(now),
		NamespaceID: "10",
		PushCount:   2,
		PullCount:   4,
		SizeDelta:   8,
		TagDelta:    1,
	}

	items, err := svc.GetNamespaceTrends(t.Context(), "10", 1)
	require.NoError(t, err)
	require.Len(t, items, 24)
	require.Equal(t, int64(2), items[23].PushCount)
	require.Equal(t, int64(4), items[23].PullCount)
	require.Equal(t, int64(8), items[23].SizeDelta)
	require.Equal(t, int64(1), items[23].TagDelta)
}

func newTestService(t *testing.T, repo *fakeAnalyticsRepository, backend string, locker lock.Locker) *service {
	t.Helper()

	digCon := dig.New()
	require.NoError(t, digCon.Provide(func() *config.Configuration {
		return &config.Configuration{
			Analytics: config.ConfigurationAnalytics{
				Enabled:        true,
				CounterBackend: backend,
				RetentionDays:  defaultRetentionDays,
			},
		}
	}))
	require.NoError(t, digCon.Provide(func() repoanalytics.AnalyticsRepository {
		return repo
	}))
	require.NoError(t, digCon.Provide(func() counter.Counter {
		return &fakeCounter{
			hashes: make(map[string]map[string]int64),
			sets:   make(map[string]map[string]struct{}),
		}
	}))
	require.NoError(t, digCon.Provide(func() lock.Locker {
		return locker
	}))
	require.NoError(t, NewService(digCon))

	var svc Service
	require.NoError(t, digCon.Invoke(func(service Service) {
		svc = service
	}))
	return svc.(*service)
}

type fakeLocker struct {
	keys []string
	errs map[string]error
}

func (l *fakeLocker) Acquire(_ context.Context, _ string, _, _ time.Duration) (lock.Lock, error) {
	return nil, errors.New("not implemented")
}

func (l *fakeLocker) AcquireWithRenew(_ context.Context, key string, _, _ time.Duration) error {
	l.keys = append(l.keys, key)
	return l.errs[key]
}

type fakeAnalyticsRepository struct {
	userRows      map[int64]*models.UserActivityHourly
	namespaceRows map[int64]*models.NamespaceActivityHourly
}

func newFakeAnalyticsRepository() *fakeAnalyticsRepository {
	return &fakeAnalyticsRepository{
		userRows:      make(map[int64]*models.UserActivityHourly),
		namespaceRows: make(map[int64]*models.NamespaceActivityHourly),
	}
}

func (repo *fakeAnalyticsRepository) IncrUserPush(_ context.Context, hour int64, userID string, delta int64) error {
	row := repo.userRows[hour]
	if row == nil {
		row = &models.UserActivityHourly{Hour: hour, UserID: userID}
		repo.userRows[hour] = row
	}
	row.PushCount += delta
	return nil
}

func (repo *fakeAnalyticsRepository) IncrNamespace(
	_ context.Context,
	hour int64,
	namespaceID string,
	pushDelta, pullDelta, sizeDelta, tagDelta int64,
) error {

	row := repo.namespaceRows[hour]
	if row == nil {
		row = &models.NamespaceActivityHourly{Hour: hour, NamespaceID: namespaceID}
		repo.namespaceRows[hour] = row
	}
	row.PushCount += pushDelta
	row.PullCount += pullDelta
	row.SizeDelta += sizeDelta
	row.TagDelta += tagDelta
	return nil
}

func (repo *fakeAnalyticsRepository) ListUserPush(
	_ context.Context,
	userID string,
	startHour, endHour int64,
) ([]*models.UserActivityHourly, error) {

	rows := make([]*models.UserActivityHourly, 0, len(repo.userRows))
	for _, row := range repo.userRows {
		if row.UserID == userID && row.Hour >= startHour && row.Hour <= endHour {
			rows = append(rows, row)
		}
	}
	return rows, nil
}

func (repo *fakeAnalyticsRepository) ListNamespaceActivity(
	_ context.Context,
	namespaceID string,
	startHour, endHour int64,
) ([]*models.NamespaceActivityHourly, error) {

	rows := make([]*models.NamespaceActivityHourly, 0, len(repo.namespaceRows))
	for _, row := range repo.namespaceRows {
		if row.NamespaceID == namespaceID && row.Hour >= startHour && row.Hour <= endHour {
			rows = append(rows, row)
		}
	}
	return rows, nil
}

func (repo *fakeAnalyticsRepository) DeleteBeforeHour(_ context.Context, _ int64) error {
	return nil
}

type fakeCounter struct {
	hashes map[string]map[string]int64
	sets   map[string]map[string]struct{}
}

func (c *fakeCounter) HIncrBy(_ context.Context, key string, fields map[string]int64) error {
	h := c.hashes[key]
	if h == nil {
		h = make(map[string]int64)
		c.hashes[key] = h
	}
	for field, delta := range fields {
		h[field] += delta
	}
	return nil
}

func (c *fakeCounter) HGetAll(_ context.Context, key string) (map[string]int64, error) {
	return c.hashes[key], nil
}

func (c *fakeCounter) Del(_ context.Context, keys ...string) error {
	for _, key := range keys {
		delete(c.hashes, key)
		delete(c.sets, key)
	}
	return nil
}

func (c *fakeCounter) SAdd(_ context.Context, key string, members ...string) error {
	s := c.sets[key]
	if s == nil {
		s = make(map[string]struct{})
		c.sets[key] = s
	}
	for _, m := range members {
		s[m] = struct{}{}
	}
	return nil
}

func (c *fakeCounter) SMembers(_ context.Context, key string) ([]string, error) {
	s := c.sets[key]
	result := make([]string, 0, len(s))
	for m := range s {
		result = append(result, m)
	}
	return result, nil
}

func (c *fakeCounter) SRem(_ context.Context, key string, members ...string) error {
	s := c.sets[key]
	if s == nil {
		return nil
	}
	for _, m := range members {
		delete(s, m)
	}
	return nil
}

func (c *fakeCounter) Rename(_ context.Context, oldKey, newKey string) error {
	if _, ok := c.hashes[newKey]; ok {
		return nil
	}
	if _, ok := c.sets[newKey]; ok {
		return nil
	}
	if h, ok := c.hashes[oldKey]; ok {
		c.hashes[newKey] = h
		delete(c.hashes, oldKey)
	}
	if s, ok := c.sets[oldKey]; ok {
		c.sets[newKey] = s
		delete(c.sets, oldKey)
	}
	return nil
}

func (c *fakeCounter) Exists(_ context.Context, keys ...string) (int64, error) {
	var count int64
	for _, key := range keys {
		if _, ok := c.hashes[key]; ok {
			count++
			continue
		}
		if _, ok := c.sets[key]; ok {
			count++
		}
	}
	return count, nil
}

func (c *fakeCounter) Shutdown(_ context.Context) error {
	return nil
}
