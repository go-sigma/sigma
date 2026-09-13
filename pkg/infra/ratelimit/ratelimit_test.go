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

package ratelimit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/infra/cache"
)

func newTestLimiter(t *testing.T, window time.Duration) Limiter {
	t.Helper()
	cfg := &config.Configuration{}
	cfg.Cache.WithDefaults()
	cfg.Cache.Type = enums.CacherTypeInmemory
	cfg.Cache.Prefix = "test-cache"
	limiter, err := New(&config.ConfigurationAuthLoginRateLimit{Window: window}, cache.Params{Config: cfg})
	require.NoError(t, err)
	return limiter
}

func TestCacheLimiter(t *testing.T) {
	limiter := newTestLimiter(t, time.Minute)

	count, err := limiter.Count(t.Context(), "alpha")
	require.NoError(t, err)
	require.Zero(t, count)

	require.NoError(t, limiter.Incr(t.Context(), "alpha"))
	require.NoError(t, limiter.Incr(t.Context(), "alpha"))

	count, err = limiter.Count(t.Context(), "alpha")
	require.NoError(t, err)
	require.Equal(t, int64(2), count)

	require.NoError(t, limiter.Reset(t.Context(), "alpha"))
	count, err = limiter.Count(t.Context(), "alpha")
	require.NoError(t, err)
	require.Zero(t, count)

	count, err = limiter.Count(t.Context(), "beta")
	require.NoError(t, err)
	require.Zero(t, count)
}

func TestCacheLimiterExpiry(t *testing.T) {
	limiter := newTestLimiter(t, 50*time.Millisecond)

	require.NoError(t, limiter.Incr(t.Context(), "alpha"))
	count, err := limiter.Count(t.Context(), "alpha")
	require.NoError(t, err)
	require.Equal(t, int64(1), count)

	time.Sleep(60 * time.Millisecond)
	count, err = limiter.Count(t.Context(), "alpha")
	require.NoError(t, err)
	require.Zero(t, count)

	require.NoError(t, limiter.Incr(t.Context(), "alpha"))
	count, err = limiter.Count(t.Context(), "alpha")
	require.NoError(t, err)
	require.Equal(t, int64(1), count)
}
