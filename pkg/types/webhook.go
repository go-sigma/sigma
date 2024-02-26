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

// PostWebhookRequest ...
type PostWebhookRequest struct {
	NamespaceID     *int64  `json:"namespace_id,omitempty" validate:"omitempty,numeric" example:"1"`
	URL             string  `json:"url" validate:"required,url,max=128" example:"http://example.com/webhook"`
	Secret          *string `json:"secret,omitempty" validate:"omitempty,max=63" example:"secret"`
	SslVerify       bool    `json:"ssl_verify" example:"true"`
	RetryTimes      int     `json:"retry_times" validate:"required" example:"3"`
	RetryDuration   int     `json:"retry_duration" validate:"required" example:"5"`
	Enable          bool    `json:"enable" example:"true"`
	EventNamespace  *bool   `json:"event_namespace,omitempty" example:"true"`
	EventRepository bool    `json:"event_repository" example:"true"`
	EventTag        bool    `json:"event_tag" example:"true"`
	EventArtifact   bool    `json:"event_artifact" example:"true"`
	EventMember     bool    `json:"event_member" example:"true"`
}

type PutWebhookRequest struct {
	ID int64 `json:"webhook_id" param:"webhook_id" validate:"required,number" swaggerignore:"true"`

	Url             *string `json:"url,omitempty" validate:"omitempty,url,max=128" example:"http://example.com/webhook"`
	Secret          *string `json:"secret,omitempty" validate:"omitempty,max=63" example:"secret"`
	SslVerify       *bool   `json:"ssl_verify,omitempty" validate:"omitempty,boolean" example:"true"`
	RetryTimes      *int    `json:"retry_times,omitempty" validate:"omitempty,number" example:"3"`
	RetryDuration   *int    `json:"retry_duration,omitempty" validate:"omitempty,number" example:"5"`
	Enable          *bool   `json:"enable,omitempty" validate:"omitempty,boolean" example:"true"`
	EventNamespace  *bool   `json:"event_namespace,omitempty" validate:"omitempty,boolean" example:"true"`
	EventRepository *bool   `json:"event_repository,omitempty" validate:"omitempty,boolean" example:"true"`
	EventTag        *bool   `json:"event_tag,omitempty" validate:"omitempty,boolean" example:"true"`
	EventArtifact   *bool   `json:"event_artifact,omitempty" validate:"omitempty,boolean" example:"true"`
	EventMember     *bool   `json:"event_member,omitempty" validate:"omitempty,boolean" example:"true"`
}

// DeleteWebhookRequest ...
type DeleteWebhookRequest struct {
	ID int64 `json:"webhook_id" param:"webhook_id" example:"1"`
}

// GetWebhookRequest ...
type GetWebhookRequest struct {
	ID int64 `json:"webhook_id" param:"webhook_id" example:"1"`
}

// WebhookItem ...
type WebhookItem struct {
	ID              int64   `json:"id" example:"1"`
	NamespaceID     *int64  `json:"namespace_id,omitempty" example:"1"`
	URL             string  `json:"url" example:"http://example.com/webhook"`
	Secret          *string `json:"secret,omitempty" example:"secret"`
	SslVerify       bool    `json:"ssl_verify" example:"true"`
	RetryTimes      int     `json:"retry_times" example:"3"`
	RetryDuration   int     `json:"retry_duration" example:"5"`
	Enable          bool    `json:"enable" example:"true"`
	EventNamespace  *bool   `json:"event_namespace,omitempty" example:"true"`
	EventRepository bool    `json:"event_repository" example:"true"`
	EventTag        bool    `json:"event_tag" example:"true"`
	EventArtifact   bool    `json:"event_artifact" example:"true"`
	EventMember     bool    `json:"event_member" example:"true"`
	CreatedAt       string  `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt       string  `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// ListWebhookRequest ...
type ListWebhookRequest struct {
	Pagination
	Sortable

	NamespaceID *int64 `json:"namespace_id,omitempty" query:"namespace_id" validate:"omitempty,numeric" example:"1"`
}

// ListWebhookLogRequest ...
type ListWebhookLogRequest struct {
	Pagination
	Sortable

	WebhookID int64 `json:"webhook_id" param:"webhook_id" example:"1"`
}

// WebhookLogItem ...
type WebhookLogItem struct {
	ID           int64                     `json:"id" example:"1"`
	Action       enums.WebhookAction       `json:"action" example:"action"`
	ResourceType enums.WebhookResourceType `json:"resource_type" example:"event"`
	StatusCode   int                       `json:"status_code" example:"200"`
	ReqHeader    string                    `json:"req_header" example:""`
	ReqBody      string                    `json:"req_body" example:""`
	RespHeader   string                    `json:"resp_header" example:""`
	RespBody     string                    `json:"resp_body" example:""`
	CreatedAt    string                    `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt    string                    `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// GetWebhookLogRequest ...
type GetWebhookLogRequest struct {
	WebhookID    int64 `json:"webhook_id" param:"webhook_id" example:"1"`
	WebhookLogID int64 `json:"webhook_log_id" param:"webhook_log_id" example:"1"`
}

// GetWebhookLogResendRequest ...
type GetWebhookLogResendRequest struct {
	WebhookID    int64 `json:"webhook_id" param:"webhook_id" example:"1"`
	WebhookLogID int64 `json:"webhook_log_id" param:"webhook_log_id" example:"1"`
}

// DeleteWebhookLogRequest ...
type DeleteWebhookLogRequest struct {
	WebhookID    int64 `json:"webhook_id" param:"webhook_id" example:"1"`
	WebhookLogID int64 `json:"webhook_log_id" param:"webhook_log_id" example:"1"`
}

// GetWebhookPingRequest ...
type GetWebhookPingRequest struct {
	WebhookID int64 `json:"webhook_id" param:"webhook_id" example:"1"`
}

// DaemonWebhookNamespace ...
type DaemonWebhookNamespace struct {
	ID              int64            `json:"id" example:"1"`
	Name            string           `json:"name" example:"test"`
	Description     *string          `json:"description,omitempty" example:"i am just description"`
	Overview        *string          `json:"overview,omitempty" example:"i am just overview"`
	Visibility      enums.Visibility `json:"visibility" example:"private"`
	RepositoryLimit int64            `json:"repository_limit" example:"10"`
	RepositoryCount int64            `json:"repository_count" example:"10"`
	TagLimit        int64            `json:"tag_limit" example:"10"`
	TagCount        int64            `json:"tag_count" example:"10"`
	Size            int64            `json:"size" example:"10000"`
	SizeLimit       int64            `json:"size_limit" example:"10000"`

	CreatedAt string `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt string `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// DaemonWebhookRepository ...
type DaemonWebhookRepository struct {
	ID          int64            `json:"id" example:"1"`
	NamespaceID int64            `json:"namespace_id" example:"1"`
	Name        string           `json:"name" example:"busybox"`
	Description *string          `json:"description,omitempty" example:"i am just description"`
	Overview    *string          `json:"overview,omitempty" example:"i am just overview"`
	Visibility  enums.Visibility `json:"visibility" example:"private"`
	TagCount    int64            `json:"tag_count" example:"100"`
	TagLimit    *int64           `json:"tag_limit" example:"1000"`
	SizeLimit   *int64           `json:"size_limit" example:"10000"`
	Size        *int64           `json:"size" example:"10000"`

	CreatedAt string `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt string `json:"updated_at" example:"2006-01-02 15:04:05"`
}

type DaemonWebhookTag struct {
}

type DaemonWebhookArtifact struct {
}

// DaemonWebhookPayloadPing ...
type DaemonWebhookPayloadPing struct {
	ResourceType enums.WebhookResourceType `json:"resource_type" example:"webhook"`
	Action       enums.WebhookAction       `json:"action" example:"ping"`
	Namespace    *DaemonWebhookNamespace   `json:"namespace"`
}

// DaemonWebhookPayloadNamespace ...
type DaemonWebhookPayloadNamespace struct {
	ResourceType enums.WebhookResourceType `json:"resource_type" example:"namespace"`
	Action       enums.WebhookAction       `json:"action" example:"create"`
	Namespace    DaemonWebhookNamespace    `json:"namespace"`
}

// DaemonWebhookPayloadRepository ...
type DaemonWebhookPayloadRepository struct {
	ResourceType enums.WebhookResourceType `json:"resource_type" example:"repository"`
	Action       enums.WebhookAction       `json:"action" example:"create"`
	Namespace    DaemonWebhookNamespace    `json:"namespace"`
	Repository   DaemonWebhookRepository   `json:"repository"`
}

// DaemonWebhookPayloadTag ...
type DaemonWebhookPayloadTag struct {
	ResourceType enums.WebhookResourceType `json:"resource_type" example:"tag"`
	Action       enums.WebhookAction       `json:"action" example:"create"`
	Namespace    DaemonWebhookNamespace    `json:"namespace"`
	Repository   DaemonWebhookRepository   `json:"repository"`
	Tag          DaemonWebhookTag          `json:"tag"`
}

// DaemonWebhookPayloadArtifact ...
type DaemonWebhookPayloadArtifact struct {
	ResourceType enums.WebhookResourceType `json:"resource_type" example:"artifact"`
	Action       enums.WebhookAction       `json:"action" example:"create"`
	Namespace    DaemonWebhookNamespace    `json:"namespace"`
	Repository   DaemonWebhookRepository   `json:"repository"`
	Artifact     DaemonWebhookArtifact     `json:"artifact"`
}

// DaemonWebhookPayloadMember ...
type DaemonWebhookPayloadMember struct {
	ResourceType enums.WebhookResourceType `json:"resource_type" example:"member"`
	Action       enums.WebhookAction       `json:"action" example:"create"`
	Namespace    *DaemonWebhookNamespace   `json:"namespace"`
}
