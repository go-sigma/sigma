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

import "github.com/prometheus/client_golang/prometheus"

var (
	loginFailuresTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "login_auth_failures_total",
		Help: "Total number of recorded failed login attempts.",
	})
	loginDelayedTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "login_auth_delayed_total",
		Help: "Total number of login attempts delayed by throttling.",
	})
	loginCacheErrorsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "login_auth_cache_errors_total",
		Help: "Total number of failed-login cache operation errors.",
	})
)

func init() {
	prometheus.MustRegister(loginFailuresTotal, loginDelayedTotal, loginCacheErrorsTotal)
}

func recordFailure() {
	loginFailuresTotal.Inc()
}

func recordCacheError() {
	loginCacheErrorsTotal.Inc()
}

// RecordDelayed increments the counter of login attempts delayed by throttling.
func RecordDelayed() {
	loginDelayedTotal.Inc()
}
