// Copyright 2023 sigma
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

package daemons

import (
	"time"

	"github.com/hako/durafmt"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

// gcRunnerBase represents the common fields of all GC runner models
type gcRunnerBase interface {
	GetID() int64
	GetStatus() enums.TaskCommonStatus
	GetMessage() []byte
	GetSuccessCount() *int64
	GetFailedCount() *int64
	GetStartedAt() *int64
	GetEndedAt() *int64
	GetDuration() *int64
	GetCreatedAt() int64
	GetUpdatedAt() int64
}

// Ensure all runner types implement gcRunnerBase
var (
	_ gcRunnerBase = (*models.DaemonGcTagRunner)(nil)
	_ gcRunnerBase = (*models.DaemonGcRepositoryRunner)(nil)
	_ gcRunnerBase = (*models.DaemonGcArtifactRunner)(nil)
	_ gcRunnerBase = (*models.DaemonGcBlobRunner)(nil)
)

// gcRunnerItemBase represents the common fields of all GC runner item response types
type gcRunnerItemBase struct {
	ID           int64                  `json:"id"`
	Status       enums.TaskCommonStatus `json:"status"`
	Message      string                 `json:"message"`
	SuccessCount *int64                 `json:"success_count"`
	FailedCount  *int64                 `json:"failed_count"`
	StartedAt    *string                `json:"started_at"`
	EndedAt      *string                `json:"ended_at"`
	RawDuration  *int64                 `json:"raw_duration"`
	Duration     *string                `json:"duration"`
	CreatedAt    string                 `json:"created_at"`
	UpdatedAt    string                 `json:"updated_at"`
}

// formatRunnerTimestamp converts a millisecond timestamp to a formatted string
func formatRunnerTimestamp(timestampMs *int64) *string {
	if timestampMs == nil {
		return nil
	}
	return ptr.Of(time.Unix(0, int64(time.Millisecond)*ptr.To(timestampMs)).UTC().Format(consts.DefaultTimePattern))
}

// formatRunnerDuration converts a millisecond duration to a human-readable string
func formatRunnerDuration(durationMs *int64) *string {
	if durationMs == nil {
		return nil
	}
	return ptr.Of(durafmt.ParseShort(time.Millisecond * time.Duration(ptr.To(durationMs))).String())
}

// buildRunnerItemBase creates a gcRunnerItemBase from a gcRunnerBase
func buildRunnerItemBase(runnerObj gcRunnerBase) gcRunnerItemBase {
	startedAt := formatRunnerTimestamp(runnerObj.GetStartedAt())
	endedAt := formatRunnerTimestamp(runnerObj.GetEndedAt())
	duration := formatRunnerDuration(runnerObj.GetDuration())

	return gcRunnerItemBase{
		ID:           runnerObj.GetID(),
		Status:       runnerObj.GetStatus(),
		Message:      string(runnerObj.GetMessage()),
		SuccessCount: runnerObj.GetSuccessCount(),
		FailedCount:  runnerObj.GetFailedCount(),
		RawDuration:  runnerObj.GetDuration(),
		Duration:     duration,
		StartedAt:    startedAt,
		EndedAt:      endedAt,
		CreatedAt:    time.Unix(0, int64(time.Millisecond)*runnerObj.GetCreatedAt()).UTC().Format(consts.DefaultTimePattern),
		UpdatedAt:    time.Unix(0, int64(time.Millisecond)*runnerObj.GetUpdatedAt()).UTC().Format(consts.DefaultTimePattern),
	}
}
