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

package cacher

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	backendMemory = "inmemory"
	backendRedis  = "redis"

	resultHit         = "hit"
	resultMiss        = "miss"
	resultNegativeHit = "negative_hit"
	resultFetchError  = "fetch_error"
	resultSetError    = "set_error"
)

var (
	cacheRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_requests_total",
			Help: "Total number of cache requests.",
		},
		[]string{"backend", "prefix", "result"},
	)
	cacheFetchDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cache_fetch_duration_seconds",
			Help:    "Read-through fetch duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"backend", "prefix", "result"},
	)
)

func init() {
	prometheus.MustRegister(cacheRequestsTotal, cacheFetchDuration)
}

func recordCacheRequest(backend, prefix, result string) {
	cacheRequestsTotal.WithLabelValues(backend, prefix, result).Inc()
}

func observeCacheFetch(backend, prefix, result string, startedAt time.Time) {
	cacheFetchDuration.WithLabelValues(backend, prefix, result).Observe(time.Since(startedAt).Seconds())
}
