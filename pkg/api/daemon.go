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

package api

import (
	"github.com/go-sigma/sigma/pkg/api/enums"
)

// TaskSbom is the payload for an SBOM generation task.
type TaskSbom struct {
	ArtifactID string `json:"artifact_id"`
}

// TaskVulnerability is the payload for a vulnerability scanning task.
type TaskVulnerability struct {
	ArtifactID string `json:"artifact_id"`
}

// TaskProxyArtifact is the payload for a proxied artifact task.
type TaskProxyArtifact struct {
	BlobDigest string `json:"blob_digest"`
}

// DaemonGcPayload is the payload for a garbage collection daemon task.
type DaemonGcPayload struct {
	RunnerID string `json:"runner_id"`
}

// DaemonGcRepositoryPayload is the payload for a repository garbage collection task.
type DaemonGcRepositoryPayload struct {
	RunnerID string `json:"runner_id"`
}

// DaemonWebhookPayload is the payload used to dispatch a webhook task.
type DaemonWebhookPayload struct {
	NamespaceID  *string                   `json:"namespace_id"`
	WebhookID    string                    `json:"webhook_id"`
	WebhookLogID *string                   `json:"webhook_log_id"`
	Type         enums.WebhookType         `json:"type"`
	Action       enums.WebhookAction       `json:"action"`
	ResourceType enums.WebhookResourceType `json:"resource_type"`
	Payload      []byte                    `json:"payload"`
}

// DaemonBuilderPayload is the payload for a builder daemon task.
type DaemonBuilderPayload struct {
	Action       enums.DaemonBuilderAction `json:"action"`
	BuilderID    string                    `json:"builder_id"`
	RunnerID     string                    `json:"runner_id"`
	RepositoryID string                    `json:"repository_id"`
}

// DaemonArtifactPushedPayload is the payload emitted after an artifact is pushed.
type DaemonArtifactPushedPayload struct {
	RepositoryID string `json:"repository_id"`
}

// DaemonTagPushedPayload is the payload emitted after a tag is pushed.
type DaemonTagPushedPayload struct {
	RepositoryID string `json:"repository_id"`
	Tag          string `json:"tag"`
}

// DaemonCodeRepositoryPayload is the payload for source code repository synchronization.
type DaemonCodeRepositoryPayload struct {
	User3rdPartyID string `json:"user_3rdparty_id"`
}

// RetentionPatternPayload contains tag retention pattern rules.
type RetentionPatternPayload struct {
	Patterns []string `json:"patterns"`
}

// UpdateGcArtifactRuleRequest is the request for updating artifact GC rules.
type UpdateGcArtifactRuleRequest struct {
	NamespaceID string `json:"namespace_id" param:"namespace_id" validate:"" example:"550e8400-e29b-41d4-a716-446655440000"`

	RetentionDay int     `json:"retention_day" validate:"gte=0,lte=180" example:"7" minimum:"0" maximum:"180"`
	CronEnabled  bool    `json:"cron_enabled" example:"true"`
	CronRule     *string `json:"cron_rule,omitempty" validate:"omitempty,is_valid_cron_rule" example:"0 0 * * 6"`
}

// GetGcArtifactRuleRequest is the request for getting artifact GC rules.
type GetGcArtifactRuleRequest struct {
	NamespaceID string `json:"namespace_id" param:"namespace_id" validate:""`
}

// GetGcArtifactRuleResponse is the response containing artifact GC rules.
type GetGcArtifactRuleResponse struct {
	IsRunning       bool    `json:"is_running" example:"true"`
	RetentionDay    int     `json:"retention_day" example:"7"`
	CronEnabled     bool    `json:"cron_enabled" example:"true"`
	CronRule        *string `json:"cron_rule,omitempty" example:"0 0 * * 6"`
	CronNextTrigger *string `json:"cron_next_trigger,omitempty" example:"2021-01-01 00:00:00"`
	CreatedAt       string  `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt       string  `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// GetGcArtifactLatestRunnerRequest is the request for getting the latest artifact GC runner.
type GetGcArtifactLatestRunnerRequest struct {
	NamespaceID string `json:"namespace_id" param:"namespace_id" validate:""`
}

// GcArtifactRunnerItem represents an artifact GC runner execution.
type GcArtifactRunnerItem struct {
	ID           string                 `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Status       enums.TaskCommonStatus `json:"status" example:"Pending"`
	Message      string                 `json:"message" example:"log"`
	SuccessCount *int64                 `json:"success_count" example:"10"`
	FailedCount  *int64                 `json:"failed_count" example:"1"`
	StartedAt    *string                `json:"started_at" example:"2006-01-02 15:04:05"`
	EndedAt      *string                `json:"ended_at" example:"2006-01-02 15:04:05"`
	RawDuration  *int64                 `json:"raw_duration" example:"1000"`
	Duration     *string                `json:"duration" example:"1h"`
	CreatedAt    string                 `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt    string                 `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// CreateGcArtifactRunnerRequest is the request for creating an artifact GC runner.
type CreateGcArtifactRunnerRequest struct {
	NamespaceID string `json:"namespace_id" param:"namespace_id" validate:""`
}

// ListGcArtifactRunnersRequest is the request for listing artifact GC runners.
type ListGcArtifactRunnersRequest struct {
	NamespaceID string `json:"namespace_id" param:"namespace_id" validate:""`

	Pagination
	Sortable
}

// GetGcArtifactRunnerRequest is the request for getting an artifact GC runner.
type GetGcArtifactRunnerRequest struct {
	NamespaceID string `json:"namespace_id" param:"namespace_id" validate:""`
	RunnerID    string `json:"runner_id" param:"runner_id" validate:"required"`
}

// ListGcArtifactRecordsRequest is the request for listing artifact GC records.
type ListGcArtifactRecordsRequest struct {
	NamespaceID string `json:"namespace_id" param:"namespace_id" validate:""`
	RunnerID    string `json:"runner_id" param:"runner_id" validate:"required"`

	Pagination
	Sortable
}

// GcArtifactRecordItem represents a single artifact GC record.
type GcArtifactRecordItem struct {
	ID        string               `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Digest    string               `json:"digest" example:"sha256:87508bf3e050b975770b142e62db72eeb345a67d82d36ca166300d8b27e45744"`
	Status    enums.GcRecordStatus `json:"status" example:"Success"`
	Message   string               `json:"message" example:"log"`
	CreatedAt string               `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt string               `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// GetGcArtifactRecordRequest is the request for getting an artifact GC record.
type GetGcArtifactRecordRequest struct {
	NamespaceID string `json:"namespace_id" param:"namespace_id" validate:""`
	RunnerID    string `json:"runner_id" param:"runner_id" validate:"required"`
	RecordID    string `json:"record_id" param:"record_id" validate:"required"`
}

// UpdateGcBlobRuleRequest is the request for updating blob GC rules.
type UpdateGcBlobRuleRequest struct {
	NamespaceID string `json:"namespace_id" param:"namespace_id" validate:"" example:"550e8400-e29b-41d4-a716-446655440000"`

	RetentionDay int     `json:"retention_day" validate:"gte=0,lte=180" example:"7" minimum:"0" maximum:"180"`
	CronEnabled  bool    `json:"cron_enabled" example:"true"`
	CronRule     *string `json:"cron_rule,omitempty" validate:"omitempty,is_valid_cron_rule" example:"0 0 * * 6"`
}

// GetGcBlobRuleRequest is the request for getting blob GC rules.
type GetGcBlobRuleRequest struct {
	NamespaceID string `json:"namespace_id" param:"namespace_id" validate:""`
}

// GetGcBlobRuleResponse is the response containing blob GC rules.
type GetGcBlobRuleResponse struct {
	RetentionDay    int     `json:"retention_day" example:"7"`
	CronEnabled     bool    `json:"cron_enabled" example:"true"`
	CronRule        *string `json:"cron_rule,omitempty" example:"0 0 * * 6"`
	CronNextTrigger *string `json:"cron_next_trigger,omitempty" example:"2021-01-01 00:00:00"`
	CreatedAt       string  `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt       string  `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// GetGcBlobLatestRunnerRequest is the request for getting the latest blob GC runner.
type GetGcBlobLatestRunnerRequest struct {
	NamespaceID string `json:"namespace_id" param:"namespace_id" validate:""`
}

// GcBlobRunnerItem represents a blob GC runner execution.
type GcBlobRunnerItem struct {
	ID           string                 `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Status       enums.TaskCommonStatus `json:"status" example:"Pending"`
	Message      string                 `json:"message" example:"log"`
	SuccessCount *int64                 `json:"success_count" example:"10"`
	FailedCount  *int64                 `json:"failed_count" example:"1"`
	StartedAt    *string                `json:"started_at" example:"2006-01-02 15:04:05"`
	EndedAt      *string                `json:"ended_at" example:"2006-01-02 15:04:05"`
	RawDuration  *int64                 `json:"raw_duration" example:"1000"`
	Duration     *string                `json:"duration" example:"1h"`
	CreatedAt    string                 `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt    string                 `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// CreateGcBlobRunnerRequest is the request for creating a blob GC runner.
type CreateGcBlobRunnerRequest struct {
	NamespaceID string `json:"namespace_id" param:"namespace_id" validate:""`
}

// ListGcBlobRunnersRequest is the request for listing blob GC runners.
type ListGcBlobRunnersRequest struct {
	NamespaceID string `json:"namespace_id" param:"namespace_id" validate:""`

	Pagination
	Sortable
}

// GetGcBlobRunnerRequest is the request for getting a blob GC runner.
type GetGcBlobRunnerRequest struct {
	NamespaceID string `json:"namespace_id" param:"namespace_id" validate:""`
	RunnerID    string `json:"runner_id" param:"runner_id" validate:"required"`
}

// ListGcBlobRecordsRequest is the request for listing blob GC records.
type ListGcBlobRecordsRequest struct {
	NamespaceID string `json:"namespace_id" param:"namespace_id" validate:""`
	RunnerID    string `json:"runner_id" param:"runner_id" validate:"required"`

	Pagination
	Sortable
}

// GcBlobRecordItem represents a single blob GC record.
type GcBlobRecordItem struct {
	ID        string               `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Digest    string               `json:"digest" example:"sha256:87508bf3e050b975770b142e62db72eeb345a67d82d36ca166300d8b27e45744"`
	Status    enums.GcRecordStatus `json:"status" example:"Success"`
	Message   string               `json:"message" example:"log"`
	CreatedAt string               `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt string               `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// GetGcBlobRecordRequest is the request for getting a blob GC record.
type GetGcBlobRecordRequest struct {
	NamespaceID string `json:"namespace_id" param:"namespace_id" validate:""`
	RunnerID    string `json:"runner_id" param:"runner_id" validate:"required"`
	RecordID    string `json:"record_id" param:"record_id" validate:"required"`
}

// UpdateGcRepositoryRuleRequest is the request for updating repository GC rules.
type UpdateGcRepositoryRuleRequest struct {
	NamespaceID string `json:"namespace_id" param:"namespace_id" validate:"" example:"550e8400-e29b-41d4-a716-446655440000"`

	RetentionDay int     `json:"retention_day" validate:"gte=0,lte=180" example:"7" minimum:"0" maximum:"180"`
	CronEnabled  *bool   `json:"cron_enabled,omitempty" example:"true"`
	CronRule     *string `json:"cron_rule,omitempty" validate:"omitempty,is_valid_cron_rule" example:"0 0 * * 6"`
}

// GetGcRepositoryRuleRequest is the request for getting repository GC rules.
type GetGcRepositoryRuleRequest struct {
	NamespaceID string `json:"namespace_id" param:"namespace_id" validate:""`
}

// GetGcRepositoryRuleResponse is the response containing repository GC rules.
type GetGcRepositoryRuleResponse struct {
	RetentionDay    int     `json:"retention_day" example:"7"`
	CronEnabled     bool    `json:"cron_enabled" example:"true"`
	CronRule        *string `json:"cron_rule,omitempty" example:"0 0 * * 6"`
	CronNextTrigger *string `json:"cron_next_trigger,omitempty" example:"2021-01-01 00:00:00"`
	CreatedAt       string  `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt       string  `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// GetGcRepositoryLatestRunnerRequest is the request for getting the latest repository GC runner.
type GetGcRepositoryLatestRunnerRequest struct {
	NamespaceID string `json:"namespace_id" param:"namespace_id" validate:""`
}

// GcRepositoryRunnerItem represents a repository GC runner execution.
type GcRepositoryRunnerItem struct {
	ID           string                 `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Status       enums.TaskCommonStatus `json:"status" example:"Pending"`
	Message      string                 `json:"message" example:"log"`
	SuccessCount *int64                 `json:"success_count" example:"10"`
	FailedCount  *int64                 `json:"failed_count" example:"1"`
	StartedAt    *string                `json:"started_at" example:"2006-01-02 15:04:05"`
	EndedAt      *string                `json:"ended_at" example:"2006-01-02 15:04:05"`
	RawDuration  *int64                 `json:"raw_duration" example:"1000"`
	Duration     *string                `json:"duration" example:"1h"`
	CreatedAt    string                 `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt    string                 `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// CreateGcRepositoryRunnerRequest is the request for creating a repository GC runner.
type CreateGcRepositoryRunnerRequest struct {
	NamespaceID string `json:"namespace_id" param:"namespace_id" validate:""`
}

// ListGcRepositoryRunnersRequest is the request for listing repository GC runners.
type ListGcRepositoryRunnersRequest struct {
	NamespaceID string `json:"namespace_id" param:"namespace_id" validate:""`

	Pagination
	Sortable
}

// GetGcRepositoryRunnerRequest is the request for getting a repository GC runner.
type GetGcRepositoryRunnerRequest struct {
	NamespaceID string `json:"namespace_id" param:"namespace_id" validate:""`
	RunnerID    string `json:"runner_id" param:"runner_id" validate:"required"`
}

// ListGcRepositoryRecordsRequest is the request for listing repository GC records.
type ListGcRepositoryRecordsRequest struct {
	NamespaceID string `json:"namespace_id" param:"namespace_id" validate:""`
	RunnerID    string `json:"runner_id" param:"runner_id" validate:"required"`

	Pagination
	Sortable
}

// GcRepositoryRecordItem represents a single repository GC record.
type GcRepositoryRecordItem struct {
	ID         string               `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Repository string               `json:"repository" example:"library/busybox"`
	Status     enums.GcRecordStatus `json:"status" example:"Success"`
	Message    string               `json:"message" example:"log"`
	CreatedAt  string               `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt  string               `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// GetGcRepositoryRecordRequest is the request for getting a repository GC record.
type GetGcRepositoryRecordRequest struct {
	NamespaceID string `json:"namespace_id" param:"namespace_id" validate:""`
	RunnerID    string `json:"runner_id" param:"runner_id" validate:"required"`
	RecordID    string `json:"record_id" param:"record_id" validate:"required"`
}

// UpdateGcTagRuleRequest is the request for updating tag GC rules.
type UpdateGcTagRuleRequest struct {
	NamespaceID string `json:"namespace_id" param:"namespace_id" validate:"" example:"550e8400-e29b-41d4-a716-446655440000"`

	CronEnabled         bool                    `json:"cron_enabled" example:"true"`
	CronRule            *string                 `json:"cron_rule,omitempty" validate:"omitempty,is_valid_cron_rule" example:"0 0 * * 6"`
	RetentionRuleType   enums.RetentionRuleType `json:"retention_rule_type" validate:"is_valid_retention_rule_type" example:"Day"`
	RetentionRuleAmount int64                   `json:"retention_rule_amount" validate:"number,gte=1,lte=180" example:"30"  minimum:"1" maximum:"180"`
	RetentionPattern    *string                 `json:"retention_pattern,omitempty" validate:"omitempty,is_valid_retention_pattern" example:"v*,1.*"`
}

// GetGcTagRuleRequest is the request for getting tag GC rules.
type GetGcTagRuleRequest struct {
	NamespaceID string `json:"namespace_id" param:"namespace_id" validate:""`
}

// GetGcTagRuleResponse is the response containing tag GC rules.
type GetGcTagRuleResponse struct {
	CronEnabled         bool                    `json:"cron_enabled" example:"true"`
	CronRule            *string                 `json:"cron_rule,omitempty" example:"0 0 * * 6"`
	CronNextTrigger     *string                 `json:"cron_next_trigger,omitempty" example:"2021-01-01 00:00:00"`
	RetentionRuleType   enums.RetentionRuleType `json:"retention_rule_type,omitempty" example:"Day"`
	RetentionRuleAmount int64                   `json:"retention_rule_amount,omitempty"  example:"30"`
	RetentionPattern    *string                 `json:"retention_pattern,omitempty" example:"v*,1.*"`
	CreatedAt           string                  `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt           string                  `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// GetGcTagLatestRunnerRequest is the request for getting the latest tag GC runner.
type GetGcTagLatestRunnerRequest struct {
	NamespaceID string `json:"namespace_id" param:"namespace_id" validate:""`
}

// GcTagRunnerItem represents a tag GC runner execution.
type GcTagRunnerItem struct {
	ID           string                 `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Status       enums.TaskCommonStatus `json:"status" example:"Pending"`
	Message      string                 `json:"message" example:"log"`
	SuccessCount *int64                 `json:"success_count" example:"10"`
	FailedCount  *int64                 `json:"failed_count" example:"1"`
	StartedAt    *string                `json:"started_at" example:"2006-01-02 15:04:05"`
	EndedAt      *string                `json:"ended_at" example:"2006-01-02 15:04:05"`
	RawDuration  *int64                 `json:"raw_duration" example:"1000"`
	Duration     *string                `json:"duration" example:"1h"`
	CreatedAt    string                 `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt    string                 `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// CreateGcTagRunnerRequest is the request for creating a tag GC runner.
type CreateGcTagRunnerRequest struct {
	NamespaceID string `json:"namespace_id" param:"namespace_id" validate:""`
}

// ListGcTagRunnersRequest is the request for listing tag GC runners.
type ListGcTagRunnersRequest struct {
	NamespaceID string `json:"namespace_id" param:"namespace_id" validate:""`

	Pagination
	Sortable
}

// GetGcTagRunnerRequest is the request for getting a tag GC runner.
type GetGcTagRunnerRequest struct {
	NamespaceID string `json:"namespace_id" param:"namespace_id" validate:""`
	RunnerID    string `json:"runner_id" param:"runner_id" validate:"required"`
}

// ListGcTagRecordsRequest is the request for listing tag GC records.
type ListGcTagRecordsRequest struct {
	NamespaceID string `json:"namespace_id" param:"namespace_id" validate:""`
	RunnerID    string `json:"runner_id" param:"runner_id" validate:"required"`

	Pagination
	Sortable
}

// GcTagRecordItem represents a single tag GC record.
type GcTagRecordItem struct {
	ID        string               `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Tag       string               `json:"digest" example:"sha256:87508bf3e050b975770b142e62db72eeb345a67d82d36ca166300d8b27e45744"`
	Status    enums.GcRecordStatus `json:"status" example:"Success"`
	Message   string               `json:"message" example:"log"`
	CreatedAt string               `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt string               `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// GetGcTagRecordRequest is the request for getting a tag GC record.
type GetGcTagRecordRequest struct {
	NamespaceID string `json:"namespace_id" param:"namespace_id" validate:""`
	RunnerID    string `json:"runner_id" param:"runner_id" validate:"required"`
	RecordID    string `json:"record_id" param:"record_id" validate:"required"`
}
