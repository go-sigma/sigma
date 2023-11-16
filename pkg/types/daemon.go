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

package types

import (
	"github.com/go-sigma/sigma/pkg/types/enums"
)

// TaskSbom is the task sbom struct
type TaskSbom struct {
	ArtifactID int64 `json:"artifact_id"`
}

// TaskVulnerability is the task scan struct
type TaskVulnerability struct {
	ArtifactID int64 `json:"artifact_id"`
}

// TaskProxyArtifact is the task proxy artifact
type TaskProxyArtifact struct {
	BlobDigest string `json:"blob_digest"`
}

// DaemonGcPayload is the gc daemon payload
type DaemonGcPayload struct {
	RunnerID int64 `json:"runner_id"`
}

// DaemonGcRepositoryPayload ...
type DaemonGcRepositoryPayload struct {
	RunnerID int64 `json:"runner_id"`
}

// DaemonWebhookPayload ...
type DaemonWebhookPayload struct {
	NamespaceID  *int64                      `json:"namespace_id"`
	WebhookID    *int64                      `json:"webhook_id"`
	WebhookLogID *int64                      `json:"webhook_log_id"`
	Resend       bool                        `json:"resend"`
	Ping         bool                        `json:"ping"`
	Event        enums.WebhookResourceType   `json:"event"`
	Action       enums.WebhookResourceAction `json:"action"`
	Payload      []byte                      `json:"payload"`
}

// DaemonBuilderPayload ...
type DaemonBuilderPayload struct {
	Action       enums.DaemonBuilderAction `json:"action"`
	BuilderID    int64                     `json:"builder_id"`
	RunnerID     int64                     `json:"runner_id"`
	RepositoryID int64                     `json:"repository_id"`
}

// DaemonArtifactPushedPayload ...
type DaemonArtifactPushedPayload struct {
	RepositoryID int64 `json:"repository_id"`
}

// DaemonTagPushedPayload ...
type DaemonTagPushedPayload struct {
	RepositoryID int64  `json:"repository_id"`
	Tag          string `json:"tag"`
}

// DaemonCodeRepositoryPayload ...
type DaemonCodeRepositoryPayload struct {
	User3rdPartyID int64 `json:"user_3rdparty_id"`
}

// RetentionPatternPayload ...
type RetentionPatternPayload struct {
	Patterns []string `json:"patterns"`
}

// UpdateGcArtifactRuleRequest ...
type UpdateGcArtifactRuleRequest struct {
	NamespaceID int64 `json:"namespace_id" param:"namespace_id" validate:"number" example:"10"`

	RetentionDay int     `json:"retention_day" validate:"gte=0,lte=180" example:"10" minimum:"0" maximum:"180"`
	CronEnabled  bool    `json:"cron_enabled" example:"true"`
	CronRule     *string `json:"cron_rule,omitempty" validate:"omitempty,is_valid_cron_rule" example:"0 0 * * 6"`
}

// GetGcArtifactRuleRequest ...
type GetGcArtifactRuleRequest struct {
	NamespaceID int64 `json:"namespace_id" param:"namespace_id" validate:"number"`
}

// GetGcArtifactRuleResponse ...
type GetGcArtifactRuleResponse struct {
	IsRunning       bool    `json:"is_running" example:"true"`
	RetentionDay    int     `json:"retention_day" example:"10"`
	CronEnabled     bool    `json:"cron_enabled" example:"true"`
	CronRule        *string `json:"cron_rule,omitempty" example:"0 0 * * 6"`
	CronNextTrigger *string `json:"cron_next_trigger,omitempty" example:"2021-01-01 00:00:00"`
	CreatedAt       string  `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt       string  `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// GetGcArtifactLatestRunnerRequest ...
type GetGcArtifactLatestRunnerRequest struct {
	NamespaceID int64 `json:"namespace_id" param:"namespace_id" validate:"number"`
}

// GcArtifactRunnerItem ...
type GcArtifactRunnerItem struct {
	ID           int64                  `json:"id" example:"1"`
	Status       enums.TaskCommonStatus `json:"status" example:"Pending"`
	Message      string                 `json:"message" example:"log"`
	SuccessCount *int64                 `json:"success_count" example:"1"`
	FailedCount  *int64                 `json:"failed_count" example:"1"`
	StartedAt    *string                `json:"started_at" example:"2006-01-02 15:04:05"`
	EndedAt      *string                `json:"ended_at" example:"2006-01-02 15:04:05"`
	RawDuration  *int64                 `json:"raw_duration" example:"10"`
	Duration     *string                `json:"duration" example:"1h"`
	CreatedAt    string                 `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt    string                 `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// CreateGcArtifactRunnerRequest ...
type CreateGcArtifactRunnerRequest struct {
	NamespaceID int64 `json:"namespace_id" param:"namespace_id" validate:"number"`
}

// ListGcArtifactRunnersRequest ...
type ListGcArtifactRunnersRequest struct {
	NamespaceID int64 `json:"namespace_id" param:"namespace_id" validate:"number"`

	Pagination
	Sortable
}

// GetGcArtifactRunnerRequest ...
type GetGcArtifactRunnerRequest struct {
	NamespaceID int64 `json:"namespace_id" param:"namespace_id" validate:"number"`
	RunnerID    int64 `json:"runner_id" param:"runner_id" validate:"required,number"`
}

// ListGcArtifactRecordsRequest ...
type ListGcArtifactRecordsRequest struct {
	NamespaceID int64 `json:"namespace_id" param:"namespace_id" validate:"number"`
	RunnerID    int64 `json:"runner_id" param:"runner_id" validate:"required,number"`

	Pagination
	Sortable
}

// GcArtifactRecordItem ...
type GcArtifactRecordItem struct {
	ID        int64                `json:"id" example:"1"`
	Digest    string               `json:"digest" example:"sha256:87508bf3e050b975770b142e62db72eeb345a67d82d36ca166300d8b27e45744"`
	Status    enums.GcRecordStatus `json:"status" example:"Success"`
	Message   string               `json:"message" example:"log"`
	CreatedAt string               `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt string               `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// GetGcArtifactRecordRequest ...
type GetGcArtifactRecordRequest struct {
	NamespaceID int64 `json:"namespace_id" param:"namespace_id" validate:"number"`
	RunnerID    int64 `json:"runner_id" param:"runner_id" validate:"required,number"`
	RecordID    int64 `json:"record_id" param:"record_id" validate:"required,number"`
}

// UpdateGcBlobRuleRequest ...
type UpdateGcBlobRuleRequest struct {
	NamespaceID int64 `json:"namespace_id" param:"namespace_id" validate:"number" example:"10"`

	RetentionDay int     `json:"retention_day" validate:"gte=0,lte=180" example:"10" minimum:"0" maximum:"180"`
	CronEnabled  bool    `json:"cron_enabled" example:"true"`
	CronRule     *string `json:"cron_rule,omitempty" validate:"omitempty,is_valid_cron_rule" example:"0 0 * * 6"`
}

// GetGcBlobRuleRequest ...
type GetGcBlobRuleRequest struct {
	NamespaceID int64 `json:"namespace_id" param:"namespace_id" validate:"number"`
}

// GetGcBlobRuleResponse ...
type GetGcBlobRuleResponse struct {
	RetentionDay    int     `json:"retention_day" example:"10"`
	CronEnabled     bool    `json:"cron_enabled" example:"true"`
	CronRule        *string `json:"cron_rule,omitempty" example:"0 0 * * 6"`
	CronNextTrigger *string `json:"cron_next_trigger,omitempty" example:"2021-01-01 00:00:00"`
	CreatedAt       string  `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt       string  `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// GetGcBlobLatestRunnerRequest ...
type GetGcBlobLatestRunnerRequest struct {
	NamespaceID int64 `json:"namespace_id" param:"namespace_id" validate:"number"`
}

// GcBlobRunnerItem ...
type GcBlobRunnerItem struct {
	ID           int64                  `json:"id" example:"1"`
	Status       enums.TaskCommonStatus `json:"status" example:"Pending"`
	Message      string                 `json:"message" example:"log"`
	SuccessCount *int64                 `json:"success_count" example:"1"`
	FailedCount  *int64                 `json:"failed_count" example:"1"`
	StartedAt    *string                `json:"started_at" example:"2006-01-02 15:04:05"`
	EndedAt      *string                `json:"ended_at" example:"2006-01-02 15:04:05"`
	RawDuration  *int64                 `json:"raw_duration" example:"10"`
	Duration     *string                `json:"duration" example:"1h"`
	CreatedAt    string                 `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt    string                 `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// CreateGcBlobRunnerRequest ...
type CreateGcBlobRunnerRequest struct {
	NamespaceID int64 `json:"namespace_id" param:"namespace_id" validate:"number"`
}

// ListGcBlobRunnersRequest ...
type ListGcBlobRunnersRequest struct {
	NamespaceID int64 `json:"namespace_id" param:"namespace_id" validate:"number"`

	Pagination
	Sortable
}

// GetGcBlobRunnerRequest ...
type GetGcBlobRunnerRequest struct {
	NamespaceID int64 `json:"namespace_id" param:"namespace_id" validate:"number"`
	RunnerID    int64 `json:"runner_id" param:"runner_id" validate:"required,number"`
}

// ListGcBlobRecordsRequest ...
type ListGcBlobRecordsRequest struct {
	NamespaceID int64 `json:"namespace_id" param:"namespace_id" validate:"number"`
	RunnerID    int64 `json:"runner_id" param:"runner_id" validate:"required,number"`

	Pagination
	Sortable
}

// GcBlobRecordItem ...
type GcBlobRecordItem struct {
	ID        int64                `json:"id" example:"1"`
	Digest    string               `json:"digest" example:"sha256:87508bf3e050b975770b142e62db72eeb345a67d82d36ca166300d8b27e45744"`
	Status    enums.GcRecordStatus `json:"status" example:"Success"`
	Message   string               `json:"message" example:"log"`
	CreatedAt string               `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt string               `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// GetGcBlobRecordRequest ...
type GetGcBlobRecordRequest struct {
	NamespaceID int64 `json:"namespace_id" param:"namespace_id" validate:"number"`
	RunnerID    int64 `json:"runner_id" param:"runner_id" validate:"required,number"`
	RecordID    int64 `json:"record_id" param:"record_id" validate:"required,number"`
}

// UpdateGcRepositoryRuleRequest ...
type UpdateGcRepositoryRuleRequest struct {
	NamespaceID int64 `json:"namespace_id" param:"namespace_id" validate:"number" example:"10"`

	RetentionDay int     `json:"retention_day" validate:"gte=0,lte=180" example:"10" minimum:"0" maximum:"180"`
	CronEnabled  *bool   `json:"cron_enabled,omitempty" example:"true"`
	CronRule     *string `json:"cron_rule,omitempty" validate:"omitempty,is_valid_cron_rule" example:"0 0 * * 6"`
}

// GetGcRepositoryRuleRequest ...
type GetGcRepositoryRuleRequest struct {
	NamespaceID int64 `json:"namespace_id" param:"namespace_id" validate:"number"`
}

// GetGcRepositoryRuleResponse ...
type GetGcRepositoryRuleResponse struct {
	RetentionDay    int     `json:"retention_day" example:"10"`
	CronEnabled     bool    `json:"cron_enabled" example:"true"`
	CronRule        *string `json:"cron_rule,omitempty" example:"0 0 * * 6"`
	CronNextTrigger *string `json:"cron_next_trigger,omitempty" example:"2021-01-01 00:00:00"`
	CreatedAt       string  `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt       string  `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// GetGcRepositoryLatestRunnerRequest ...
type GetGcRepositoryLatestRunnerRequest struct {
	NamespaceID int64 `json:"namespace_id" param:"namespace_id" validate:"number"`
}

// GcRepositoryRunnerItem ...
type GcRepositoryRunnerItem struct {
	ID           int64                  `json:"id" example:"1"`
	Status       enums.TaskCommonStatus `json:"status" example:"Pending"`
	Message      string                 `json:"message" example:"log"`
	SuccessCount *int64                 `json:"success_count" example:"1"`
	FailedCount  *int64                 `json:"failed_count" example:"1"`
	StartedAt    *string                `json:"started_at" example:"2006-01-02 15:04:05"`
	EndedAt      *string                `json:"ended_at" example:"2006-01-02 15:04:05"`
	RawDuration  *int64                 `json:"raw_duration" example:"10"`
	Duration     *string                `json:"duration" example:"1h"`
	CreatedAt    string                 `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt    string                 `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// CreateGcRepositoryRunnerRequest ...
type CreateGcRepositoryRunnerRequest struct {
	NamespaceID int64 `json:"namespace_id" param:"namespace_id" validate:"number"`
}

// ListGcRepositoryRunnersRequest ...
type ListGcRepositoryRunnersRequest struct {
	NamespaceID int64 `json:"namespace_id" param:"namespace_id" validate:"number"`

	Pagination
	Sortable
}

// GetGcRepositoryRunnerRequest ...
type GetGcRepositoryRunnerRequest struct {
	NamespaceID int64 `json:"namespace_id" param:"namespace_id" validate:"number"`
	RunnerID    int64 `json:"runner_id" param:"runner_id" validate:"required,number"`
}

// ListGcRepositoryRecordsRequest ...
type ListGcRepositoryRecordsRequest struct {
	NamespaceID int64 `json:"namespace_id" param:"namespace_id" validate:"number"`
	RunnerID    int64 `json:"runner_id" param:"runner_id" validate:"required,number"`

	Pagination
	Sortable
}

// GcRepositoryRecordItem ...
type GcRepositoryRecordItem struct {
	ID         int64                `json:"id" example:"1"`
	Repository string               `json:"repository" example:"library/busybox"`
	Status     enums.GcRecordStatus `json:"status" example:"Success"`
	Message    string               `json:"message" example:"log"`
	CreatedAt  string               `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt  string               `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// GetGcRepositoryRecordRequest ...
type GetGcRepositoryRecordRequest struct {
	NamespaceID int64 `json:"namespace_id" param:"namespace_id" validate:"number"`
	RunnerID    int64 `json:"runner_id" param:"runner_id" validate:"required,number"`
	RecordID    int64 `json:"record_id" param:"record_id" validate:"required,number"`
}

// UpdateGcTagRuleRequest ...
type UpdateGcTagRuleRequest struct {
	NamespaceID int64 `json:"namespace_id" param:"namespace_id" validate:"number" example:"10"`

	CronEnabled         bool                    `json:"cron_enabled" example:"true"`
	CronRule            *string                 `json:"cron_rule,omitempty" validate:"omitempty,is_valid_cron_rule" example:"0 0 * * 6"`
	RetentionRuleType   enums.RetentionRuleType `json:"retention_rule_type" validate:"is_valid_retention_rule_type" example:"Day"`
	RetentionRuleAmount int64                   `json:"retention_rule_amount" validate:"number,gte=1,lte=180" example:"1"  minimum:"1" maximum:"180"`
	RetentionPattern    *string                 `json:"retention_pattern,omitempty" validate:"omitempty,is_valid_retention_pattern" example:"v*,1.*"`
}

// GetGcTagRuleRequest ...
type GetGcTagRuleRequest struct {
	NamespaceID int64 `json:"namespace_id" param:"namespace_id" validate:"number"`
}

// GetGcTagRuleResponse ...
type GetGcTagRuleResponse struct {
	CronEnabled         bool                    `json:"cron_enabled" example:"true"`
	CronRule            *string                 `json:"cron_rule,omitempty" example:"0 0 * * 6"`
	CronNextTrigger     *string                 `json:"cron_next_trigger,omitempty" example:"2021-01-01 00:00:00"`
	RetentionRuleType   enums.RetentionRuleType `json:"retention_rule_type,omitempty" example:"Day"`
	RetentionRuleAmount int64                   `json:"retention_rule_amount,omitempty"  example:"1"`
	RetentionPattern    *string                 `json:"retention_pattern,omitempty" example:"v*,1.*"`
	CreatedAt           string                  `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt           string                  `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// GetGcTagLatestRunnerRequest ...
type GetGcTagLatestRunnerRequest struct {
	NamespaceID int64 `json:"namespace_id" param:"namespace_id" validate:"number"`
}

// GcTagRunnerItem ...
type GcTagRunnerItem struct {
	ID           int64                  `json:"id" example:"1"`
	Status       enums.TaskCommonStatus `json:"status" example:"Pending"`
	Message      string                 `json:"message" example:"log"`
	SuccessCount *int64                 `json:"success_count" example:"1"`
	FailedCount  *int64                 `json:"failed_count" example:"1"`
	StartedAt    *string                `json:"started_at" example:"2006-01-02 15:04:05"`
	EndedAt      *string                `json:"ended_at" example:"2006-01-02 15:04:05"`
	RawDuration  *int64                 `json:"raw_duration" example:"10"`
	Duration     *string                `json:"duration" example:"1h"`
	CreatedAt    string                 `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt    string                 `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// CreateGcTagRunnerRequest ...
type CreateGcTagRunnerRequest struct {
	NamespaceID int64 `json:"namespace_id" param:"namespace_id" validate:"number"`
}

// ListGcTagRunnersRequest ...
type ListGcTagRunnersRequest struct {
	NamespaceID int64 `json:"namespace_id" param:"namespace_id" validate:"number"`

	Pagination
	Sortable
}

// GetGcTagRunnerRequest ...
type GetGcTagRunnerRequest struct {
	NamespaceID int64 `json:"namespace_id" param:"namespace_id" validate:"number"`
	RunnerID    int64 `json:"runner_id" param:"runner_id" validate:"required,number"`
}

// ListGcTagRecordsRequest ...
type ListGcTagRecordsRequest struct {
	NamespaceID int64 `json:"namespace_id" param:"namespace_id" validate:"number"`
	RunnerID    int64 `json:"runner_id" param:"runner_id" validate:"required,number"`

	Pagination
	Sortable
}

// GcTagRecordItem ...
type GcTagRecordItem struct {
	ID        int64                `json:"id" example:"1"`
	Tag       string               `json:"digest" example:"sha256:87508bf3e050b975770b142e62db72eeb345a67d82d36ca166300d8b27e45744"`
	Status    enums.GcRecordStatus `json:"status" example:"Success"`
	Message   string               `json:"message" example:"log"`
	CreatedAt string               `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt string               `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// GetGcTagRecordRequest ...
type GetGcTagRecordRequest struct {
	NamespaceID int64 `json:"namespace_id" param:"namespace_id" validate:"number"`
	RunnerID    int64 `json:"runner_id" param:"runner_id" validate:"required,number"`
	RecordID    int64 `json:"record_id" param:"record_id" validate:"required,number"`
}
