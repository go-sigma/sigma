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

import "github.com/go-sigma/sigma/pkg/types/enums"

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
	Target enums.GcTarget `json:"target"`
	Scope  *string        `json:"scope,omitempty"`
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

// PostDaemonRunRequest ...
type PostDaemonRunRequest struct {
	NamespaceID int64        `json:"namespace_id,omitempty" query:"namespace_id" validate:"omitempty,number" example:"123"`
	Name        enums.Daemon `json:"name" param:"name" validate:"required" example:"Gc"`
}

// GetDaemonRunRequest ...
type GetDaemonRunRequest struct {
	NamespaceID int64        `json:"namespace_id,omitempty" query:"namespace_id" validate:"omitempty,number" example:"123"`
	Name        enums.Daemon `json:"name" param:"name" validate:"required" example:"Gc"`
}

// GetDaemonRunResponse ...
type GetDaemonRunResponse struct {
	Status enums.TaskCommonStatus `json:"status"`
}

// GetDaemonLogsRequest ...
type GetDaemonLogsRequest struct {
	Pagination
	Sortable

	NamespaceID int64        `json:"namespace_id,omitempty" query:"namespace_id" validate:"omitempty,number" example:"123"`
	Name        enums.Daemon `json:"name" param:"name" validate:"required" example:"Gc"`
}

// DaemonLogItem ...
type DaemonLogItem struct {
	ID       int64                  `json:"id" example:"1"`
	Resource string                 `json:"resource" example:"test"`
	Action   enums.AuditAction      `json:"action" example:"delete"`
	Status   enums.TaskCommonStatus `json:"status"`
	Message  *string                `json:"message" example:"something error occurred"`

	CreatedAt string `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt string `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// GcRepositoryRunRequest ...
type GcRepositoryRunRequest struct {
	NamespaceID *int64
}

// GcArtifactRunRequest ...
type GcArtifactRunRequest struct {
	NamespaceID *int64
}

// GcArtifactGetRequest ...
type GcArtifactGetRequest struct {
	NamespaceID *int64
}

// GcArtifactGetResponse ...
type GcArtifactGetResponse struct {
	ID      int64                  `json:"id" example:"1"`
	Status  enums.TaskCommonStatus `json:"status" example:"Pending"`
	Message string                 `json:"message" example:"log"`

	CreatedAt string `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt string `json:"updated_at" example:"2006-01-02 15:04:05"`
}

type GcArtifactRunnersRequest struct {
	Pagination
	Sortable

	NamespaceId *int64 `query:"namespace_id"`
}

type GcArtifactRecordsRequest struct {
	Pagination
	Sortable

	RunnerID int64 `param:"runner_id"`
}

type GcArtifactRecordsResponse struct {
}

type GcArtifactRecordRequest struct {
}

type GcArtifactRecordResponse struct {
}

type GcArtifactRecordItem struct {
	ID        int64  `json:"id" example:"1"`
	Digest    string `json:"digest" example:"sha256:87508bf3e050b975770b142e62db72eeb345a67d82d36ca166300d8b27e45744"`
	CreatedAt string `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt string `json:"updated_at" example:"2006-01-02 15:04:05"`
}

type GcArtifactRunnerItem struct {
	ID        int64                  `json:"id" example:"1"`
	Status    enums.TaskCommonStatus `json:"status" example:"Pending"`
	Message   string                 `json:"message" example:"log"`
	CreatedAt string                 `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt string                 `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// GcBlobRunRequest ...
type GcBlobRunRequest struct {
}
