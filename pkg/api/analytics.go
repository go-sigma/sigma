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

package api

// DailyCount represents an aggregated count for one day.
type DailyCount struct {
	Date      string `json:"date"`
	PushCount int64  `json:"push_count"`
}

// NamespaceHourlyMetric represents namespace activity metrics aggregated by hour.
type NamespaceHourlyMetric struct {
	Hour      string `json:"hour"`
	PushCount int64  `json:"push_count"`
	PullCount int64  `json:"pull_count"`
	SizeDelta int64  `json:"size_delta"`
	TagDelta  int64  `json:"tag_delta"`
}

// ListDailyCountResponse is the response for daily analytics counts.
type ListDailyCountResponse struct {
	Items []DailyCount `json:"items"`
}

// ListNamespaceHourlyMetricResponse is the response for namespace hourly metrics.
type ListNamespaceHourlyMetricResponse struct {
	Items []NamespaceHourlyMetric `json:"items"`
}
