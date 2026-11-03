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

// OverviewScope is the data scope of the home dashboard: the whole instance for
// administrators, the current user otherwise.
type OverviewScope string

const (
	// OverviewScopeGlobal covers every namespace of the instance.
	OverviewScopeGlobal OverviewScope = "global"
	// OverviewScopeSelf covers the resources of the current user only.
	OverviewScopeSelf OverviewScope = "self"
)

// NamespaceUsage is the aggregated resource usage of a set of namespaces.
type NamespaceUsage struct {
	Namespaces   int64 `json:"namespaces"`
	Repositories int64 `json:"repositories"`
	Tags         int64 `json:"tags"`
	Size         int64 `json:"size"`
}

// OverviewStorage is the storage usage of the dashboard scope. Used is the sum of the
// distinct blobs of the scope, which is smaller than the sum of the namespace quotas because
// images share layers. Capacity is zero for non administrators, who have no meaningful total.
type OverviewStorage struct {
	Used     int64 `json:"used" example:"12884901888"`
	Capacity int64 `json:"capacity" example:"109951162777600"`
}

// OverviewCounts is the resource count summary of the dashboard scope.
type OverviewCounts struct {
	Namespaces   int64 `json:"namespaces" example:"10"`
	Repositories int64 `json:"repositories" example:"32"`
	Tags         int64 `json:"tags" example:"177"`
}

// OverviewActivity is the push and pull summary of the last 30 days, with the change
// ratio against the previous 30 days.
type OverviewActivity struct {
	Push      int64   `json:"push" example:"166"`
	Pull      int64   `json:"pull" example:"7874"`
	PushDelta float64 `json:"push_delta" example:"12.4"`
	PullDelta float64 `json:"pull_delta" example:"-3.1"`
}

// OverviewNamespace is one row of the namespace usage table.
type OverviewNamespace struct {
	ID           string `json:"id"`
	Name         string `json:"name" example:"library"`
	Visibility   string `json:"visibility" example:"public"`
	Repositories int64  `json:"repositories" example:"8"`
	Tags         int64  `json:"tags" example:"96"`
	Size         int64  `json:"size" example:"5430000000"`
	SizeLimit    int64  `json:"size_limit" example:"21474836480"`
}

// OverviewResponse is the payload of the home dashboard.
type OverviewResponse struct {
	Scope      OverviewScope       `json:"scope"`
	Storage    OverviewStorage     `json:"storage"`
	Counts     OverviewCounts      `json:"counts"`
	Activity   OverviewActivity    `json:"activity"`
	Namespaces []OverviewNamespace `json:"namespaces"`
}

// ListDailyActivityResponse is the response for the daily activity endpoint.
type ListDailyActivityResponse struct {
	Items []DailyActivity `json:"items"`
}

// DailyBuildCount is the number of successful and failed builds of one day.
type DailyBuildCount struct {
	Date    string `json:"date" example:"2026-11-03"`
	Success int64  `json:"success" example:"7"`
	Failed  int64  `json:"failed" example:"1"`
}

// BuildSummary summarizes the builds of the requested window. Fastest and slowest are
// build durations in seconds.
type BuildSummary struct {
	Success     int64   `json:"success" example:"112"`
	Failed      int64   `json:"failed" example:"11"`
	SuccessRate float64 `json:"success_rate" example:"91.1"`
	Fastest     int64   `json:"fastest" example:"120"`
	Slowest     int64   `json:"slowest" example:"389"`
}

// ListDailyBuildResponse is the response for the daily build endpoint.
type ListDailyBuildResponse struct {
	Items   []DailyBuildCount `json:"items"`
	Summary BuildSummary      `json:"summary"`
}
