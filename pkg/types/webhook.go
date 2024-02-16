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
	NamespaceID     *int64  `json:"namespace_id,omitempty" query:"namespace_id" validate:"omitempty,numeric" example:"1" swaggerignore:"true"`
	URL             string  `json:"url" validate:"required,url,max=128" example:"http://example.com/webhook"`
	Secret          *string `json:"secret,omitempty" validate:"omitempty,max=63" example:"secret"`
	SslVerify       bool    `json:"ssl_verify" validate:"required" example:"true"`
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
	ID int64 `json:"id" param:"id" validate:"required,number" swaggerignore:"true"`

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
	ID int64 `json:"id" param:"id" example:"1"`
}

// GetWebhookRequest ...
type GetWebhookRequest struct {
	ID int64 `json:"id" param:"id" example:"1"`
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

// // WebhookItem ...
// type WebhookItem = GetWebhookResponse

// ListWebhookLogRequest ...
type ListWebhookLogRequest struct {
	Pagination
	Sortable

	WebhookID int64 `json:"webhook_id" param:"webhook_id" example:"1"`
}

// WebhookLogItem ...
type WebhookLogItem struct {
	ID         int64                     `json:"id" example:"1"`
	Event      enums.WebhookResourceType `json:"event" example:"event"`
	Action     enums.WebhookAction       `json:"action" example:"action"`
	StatusCode int                       `json:"status_code" example:"200"`
	ReqHeader  string                    `json:"req_header" example:""`
	ReqBody    string                    `json:"req_body" example:""`
	RespHeader string                    `json:"resp_header" example:""`
	RespBody   string                    `json:"resp_body" example:""`
	CreatedAt  string                    `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt  string                    `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// GetWebhookLogResponse ...
type GetWebhookLogResponse = WebhookLogItem

// GetWebhookLogRequest ...
type GetWebhookLogRequest struct {
	WebhookID    int64 `json:"webhook_id" param:"webhook_id" example:"1"`
	WebhookLogID int64 `json:"webhook_log_id" param:"webhook_log_id" example:"1"`
}

// DaemonWebhookPayloadPing ...
type DaemonWebhookPayloadPing struct {
	Event string `json:"event"`
}

// DaemonWebhookPayloadNamespace ...
type DaemonWebhookPayloadNamespace struct {
	Event     string        `json:"event"`
	Operate   string        `json:"operate"`
	Namespace NamespaceItem `json:"namespace"`
}

// DaemonWebhookPayloadRepository ...
type DaemonWebhookPayloadRepository struct {
	Event      string         `json:"event"`
	Operate    string         `json:"operate"`
	Namespace  NamespaceItem  `json:"namespace"`
	Repository RepositoryItem `json:"repository"`
}

// DaemonWebhookPayloadTag ...
type DaemonWebhookPayloadTag struct {
	Event      string         `json:"event"`
	Operate    string         `json:"operate"`
	Namespace  NamespaceItem  `json:"namespace"`
	Repository RepositoryItem `json:"repository"`
	Tag        TagItem        `json:"tag"`
}

// DaemonWebhookPayloadMember ...
type DaemonWebhookPayloadMember struct {
	Event     string         `json:"event"`
	Operate   string         `json:"operate"`
	Namespace *NamespaceItem `json:"namespace,omitempty"`
	User      UserItem       `json:"user"`
}

// DaemonWebhookPayloadPullPush ...
type DaemonWebhookPayloadPullPush struct {
	Event      string         `json:"event"`
	Operate    string         `json:"operate"`
	Namespace  NamespaceItem  `json:"namespace"`
	Repository RepositoryItem `json:"repository"`
	Tag        TagItem        `json:"tag"`
}
