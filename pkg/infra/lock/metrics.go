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

package lock

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// MetricResultSuccess records successful lock operations.
	MetricResultSuccess = "success"
	// MetricResultFailure records failed lock operations.
	MetricResultFailure = "failure"
)

var (
	lockAcquireDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "sigma_lock_acquire_duration_seconds",
			Help:    "Lock acquire duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"backend", "result"},
	)
	lockHeldDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "sigma_lock_held_duration_seconds",
			Help:    "Lock held duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"backend", "result"},
	)
	lockRenewTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sigma_lock_renew_total",
			Help: "Total number of lock renew operations.",
		},
		[]string{"backend", "result"},
	)
)

func init() {
	prometheus.MustRegister(lockAcquireDuration, lockHeldDuration, lockRenewTotal)
}

// RecordLockAcquire records the time spent acquiring a lock.
func RecordLockAcquire(backend, result string, duration time.Duration) {
	lockAcquireDuration.WithLabelValues(backend, result).Observe(duration.Seconds())
}

// RecordLockHeld records the time between lock acquisition and unlock.
func RecordLockHeld(backend, result string, duration time.Duration) {
	lockHeldDuration.WithLabelValues(backend, result).Observe(duration.Seconds())
}

// RecordLockRenew records a lock renewal attempt.
func RecordLockRenew(backend, result string) {
	lockRenewTotal.WithLabelValues(backend, result).Inc()
}
