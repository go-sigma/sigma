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

package mcpserver

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	toolResultOK    = "ok"
	toolResultError = "error"
)

var (
	mcpToolCallsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sigma_mcp_tool_calls_total",
			Help: "Total number of MCP tool calls.",
		},
		[]string{"tool", "result"},
	)
	mcpToolDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "sigma_mcp_tool_duration_seconds",
			Help:    "MCP tool call duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"tool"},
	)
	mcpAuthFailuresTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sigma_mcp_auth_failures_total",
			Help: "Total number of MCP authentication failures.",
		},
		[]string{"reason"},
	)
)

func init() {
	prometheus.MustRegister(mcpToolCallsTotal, mcpToolDuration, mcpAuthFailuresTotal)
}

func recordToolCall(toolName, result string, startedAt time.Time) {
	mcpToolCallsTotal.WithLabelValues(toolName, result).Inc()
	mcpToolDuration.WithLabelValues(toolName).Observe(time.Since(startedAt).Seconds())
}

func recordAuthFailure(reason string) {
	mcpAuthFailuresTotal.WithLabelValues(reason).Inc()
}
