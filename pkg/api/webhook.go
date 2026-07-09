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

import "github.com/go-sigma/sigma/pkg/api/enums"

// PostWebhookRequest is the request for creating a webhook.
type PostWebhookRequest struct {
	NamespaceID       *string `json:"namespace_id,omitempty" validate:"omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	URL               string  `json:"url" validate:"required,url,max=128" example:"http://example.com/webhook"`
	Secret            *string `json:"secret,omitempty" validate:"omitempty,max=63" example:"secret"`
	SslVerify         bool    `json:"ssl_verify" example:"true"`
	RetryTimes        int     `json:"retry_times" validate:"required" example:"3"`
	RetryDuration     int     `json:"retry_duration" validate:"required" example:"5"`
	Enable            bool    `json:"enable" example:"true"`
	EventNamespace    *bool   `json:"event_namespace,omitempty" example:"true"`
	EventRepository   bool    `json:"event_repository" example:"true"`
	EventTag          bool    `json:"event_tag" example:"true"`
	EventArtifact     bool    `json:"event_artifact" example:"true"`
	EventMember       bool    `json:"event_member" example:"true"`
	EventDaemonTaskGc bool    `json:"event_daemon_task_gc" example:"true"`
}

// PutWebhookRequest is the request for updating a webhook.
type PutWebhookRequest struct {
	ID string `json:"webhook_id" param:"webhook_id" validate:"required" swaggerignore:"true"`

	Url               *string `json:"url,omitempty" validate:"omitempty,url,max=128" example:"http://example.com/webhook"`
	Secret            *string `json:"secret,omitempty" validate:"omitempty,max=63" example:"secret"`
	SslVerify         *bool   `json:"ssl_verify,omitempty" validate:"omitempty,boolean" example:"true"`
	RetryTimes        *int    `json:"retry_times,omitempty" validate:"omitempty,number" example:"3"`
	RetryDuration     *int    `json:"retry_duration,omitempty" validate:"omitempty,number" example:"5"`
	Enable            *bool   `json:"enable,omitempty" validate:"omitempty,boolean" example:"true"`
	EventNamespace    *bool   `json:"event_namespace,omitempty" validate:"omitempty,boolean" example:"true"`
	EventRepository   *bool   `json:"event_repository,omitempty" validate:"omitempty,boolean" example:"true"`
	EventTag          *bool   `json:"event_tag,omitempty" validate:"omitempty,boolean" example:"true"`
	EventArtifact     *bool   `json:"event_artifact,omitempty" validate:"omitempty,boolean" example:"true"`
	EventMember       *bool   `json:"event_member,omitempty" validate:"omitempty,boolean" example:"true"`
	EventDaemonTaskGc *bool   `json:"event_daemon_task_gc,omitempty" validate:"omitempty,boolean" example:"true"`
}

// DeleteWebhookRequest is the request for deleting a webhook.
type DeleteWebhookRequest struct {
	ID string `json:"webhook_id" param:"webhook_id" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// GetWebhookRequest is the request for getting a webhook.
type GetWebhookRequest struct {
	ID string `json:"webhook_id" param:"webhook_id" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// WebhookItem represents a webhook configuration returned by the API.
type WebhookItem struct {
	ID                string  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	NamespaceID       *string `json:"namespace_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	URL               string  `json:"url" example:"http://example.com/webhook"`
	Secret            *string `json:"secret,omitempty" example:"secret"`
	SslVerify         bool    `json:"ssl_verify" example:"true"`
	RetryTimes        int     `json:"retry_times" example:"3"`
	RetryDuration     int     `json:"retry_duration" example:"5"`
	Enable            bool    `json:"enable" example:"true"`
	EventNamespace    *bool   `json:"event_namespace,omitempty" example:"true"`
	EventRepository   bool    `json:"event_repository" example:"true"`
	EventTag          bool    `json:"event_tag" example:"true"`
	EventArtifact     bool    `json:"event_artifact" example:"true"`
	EventMember       bool    `json:"event_member" example:"true"`
	EventDaemonTaskGc bool    `json:"event_daemon_task_gc" example:"true"`
	CreatedAt         string  `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt         string  `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// ListWebhookRequest is the request for listing webhooks.
type ListWebhookRequest struct {
	Pagination
	Sortable

	NamespaceID *string `json:"namespace_id,omitempty" query:"namespace_id" validate:"omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// ListWebhookLogRequest is the request for listing webhook delivery logs.
type ListWebhookLogRequest struct {
	Pagination
	Sortable

	WebhookID string `json:"webhook_id" param:"webhook_id" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// WebhookLogItem represents a webhook delivery log entry.
type WebhookLogItem struct {
	ID           string                    `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Action       enums.WebhookAction       `json:"action" example:"action"`
	ResourceType enums.WebhookResourceType `json:"resource_type" example:"event"`
	StatusCode   int                       `json:"status_code" example:"200"`
	TraceID      string                    `json:"trace_id" example:"4bf92f3577b34da6a3ce929d0e0e4736"`
	TraceContext string                    `json:"trace_context" example:"{\"traceparent\":\"00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01\"}"`
	ReqHeader    string                    `json:"req_header" example:""`
	ReqBody      string                    `json:"req_body" example:""`
	RespHeader   string                    `json:"resp_header" example:""`
	RespBody     string                    `json:"resp_body" example:""`
	CreatedAt    string                    `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt    string                    `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// GetWebhookLogRequest is the request for getting a webhook delivery log.
type GetWebhookLogRequest struct {
	WebhookID    string `json:"webhook_id" param:"webhook_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	WebhookLogID string `json:"webhook_log_id" param:"webhook_log_id" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// GetWebhookLogResendRequest is the request for resending a webhook delivery log.
type GetWebhookLogResendRequest struct {
	WebhookID    string `json:"webhook_id" param:"webhook_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	WebhookLogID string `json:"webhook_log_id" param:"webhook_log_id" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// DeleteWebhookLogRequest is the request for deleting a webhook delivery log.
type DeleteWebhookLogRequest struct {
	WebhookID    string `json:"webhook_id" param:"webhook_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	WebhookLogID string `json:"webhook_log_id" param:"webhook_log_id" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// GetWebhookPingRequest is the request for sending a webhook ping event.
type GetWebhookPingRequest struct {
	WebhookID string `json:"webhook_id" param:"webhook_id" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// DaemonWebhookNamespace represents namespace data embedded in webhook payloads.
type DaemonWebhookNamespace struct {
	ID              string           `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name            string           `json:"name" example:"test"`
	Description     *string          `json:"description,omitempty" example:"i am just description"`
	Overview        *string          `json:"overview,omitempty" example:"i am just overview"`
	Visibility      enums.Visibility `json:"visibility" example:"private"`
	RepositoryLimit int64            `json:"repository_limit" example:"1000"`
	RepositoryCount int64            `json:"repository_count" example:"100"`
	TagLimit        int64            `json:"tag_limit" example:"1000"`
	TagCount        int64            `json:"tag_count" example:"100"`
	Size            int64            `json:"size" example:"10000"`
	SizeLimit       int64            `json:"size_limit" example:"10000"`

	CreatedAt string `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt string `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// DaemonWebhookRepository represents repository data embedded in webhook payloads.
type DaemonWebhookRepository struct {
	ID          string           `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	NamespaceID string           `json:"namespace_id" example:"550e8400-e29b-41d4-a716-446655440000"`
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

// DaemonWebhookTag represents tag data embedded in webhook payloads.
type DaemonWebhookTag struct {
}

// DaemonWebhookArtifact represents artifact data embedded in webhook payloads.
type DaemonWebhookArtifact struct {
}

// DaemonWebhookPayloadPing is the daemon payload for webhook ping events.
type DaemonWebhookPayloadPing struct {
	ResourceType enums.WebhookResourceType `json:"resource_type" example:"webhook"`
	Action       enums.WebhookAction       `json:"action" example:"ping"`
	Namespace    *DaemonWebhookNamespace   `json:"namespace"`
}

// DaemonWebhookPayloadNamespace is the daemon payload for namespace webhook events.
type DaemonWebhookPayloadNamespace struct {
	ResourceType enums.WebhookResourceType `json:"resource_type" example:"namespace"`
	Action       enums.WebhookAction       `json:"action" example:"create"`
	Namespace    DaemonWebhookNamespace    `json:"namespace"`
}

// DaemonWebhookPayloadRepository is the daemon payload for repository webhook events.
type DaemonWebhookPayloadRepository struct {
	ResourceType enums.WebhookResourceType `json:"resource_type" example:"repository"`
	Action       enums.WebhookAction       `json:"action" example:"create"`
	Namespace    DaemonWebhookNamespace    `json:"namespace"`
	Repository   DaemonWebhookRepository   `json:"repository"`
}

// DaemonWebhookPayloadTag is the daemon payload for tag webhook events.
type DaemonWebhookPayloadTag struct {
	ResourceType enums.WebhookResourceType `json:"resource_type" example:"tag"`
	Action       enums.WebhookAction       `json:"action" example:"create"`
	Namespace    DaemonWebhookNamespace    `json:"namespace"`
	Repository   DaemonWebhookRepository   `json:"repository"`
	Tag          DaemonWebhookTag          `json:"tag"`
}

// DaemonWebhookPayloadArtifact is the daemon payload for artifact webhook events.
type DaemonWebhookPayloadArtifact struct {
	ResourceType enums.WebhookResourceType `json:"resource_type" example:"artifact"`
	Action       enums.WebhookAction       `json:"action" example:"create"`
	Namespace    DaemonWebhookNamespace    `json:"namespace"`
	Repository   DaemonWebhookRepository   `json:"repository"`
	Artifact     DaemonWebhookArtifact     `json:"artifact"`
}

// DaemonWebhookPayloadMember is the daemon payload for namespace member webhook events.
type DaemonWebhookPayloadMember struct {
	ResourceType enums.WebhookResourceType `json:"resource_type" example:"member"`
	Action       enums.WebhookAction       `json:"action" example:"create"`
	Namespace    *DaemonWebhookNamespace   `json:"namespace"`
}

// WebhookPayload contains common webhook event metadata.
type WebhookPayload struct {
	ResourceType enums.WebhookResourceType `json:"resource_type" example:"Namespace"`
	Action       enums.WebhookAction       `json:"action" example:"Create"`
}

// WebhookPayloadUser represents a user embedded in webhook payloads.
type WebhookPayloadUser struct {
	ID        string           `json:"id"`
	Username  string           `json:"username"`
	Email     string           `json:"email"`
	Status    enums.UserStatus `json:"status"`
	LastLogin string           `json:"last_login"`

	CreatedAt string `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt string `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// WebhookPayloadGcBlob is the webhook payload for blob GC task events.
type WebhookPayloadGcBlob struct {
	WebhookPayload
	OperateType  enums.OperateType   `json:"operate_type" example:"Automatic"`
	OperateUser  *WebhookPayloadUser `json:"operate_user"`
	SuccessCount int64               `json:"success_count"`
	FailedCount  int64               `json:"failed_count"`
}

// WebhookPayloadGcArtifact is the webhook payload for artifact GC task events.
type WebhookPayloadGcArtifact struct {
	WebhookPayload
	NamespaceID  *string             `json:"namespace_id"`
	OperateType  enums.OperateType   `json:"operate_type" example:"Automatic"`
	OperateUser  *WebhookPayloadUser `json:"operate_user"`
	SuccessCount int64               `json:"success_count"`
	FailedCount  int64               `json:"failed_count"`
}

// WebhookPayloadGcRepository is the webhook payload for repository GC task events.
type WebhookPayloadGcRepository struct {
	WebhookPayload
	NamespaceID  *string             `json:"namespace_id"`
	OperateType  enums.OperateType   `json:"operate_type" example:"Automatic"`
	OperateUser  *WebhookPayloadUser `json:"operate_user"`
	SuccessCount int64               `json:"success_count"`
	FailedCount  int64               `json:"failed_count"`
}

// WebhookPayloadGcTag is the webhook payload for tag GC task events.
type WebhookPayloadGcTag struct {
	WebhookPayload
	NamespaceID  *string             `json:"namespace_id"`
	OperateType  enums.OperateType   `json:"operate_type" example:"Automatic"`
	OperateUser  *WebhookPayloadUser `json:"operate_user"`
	SuccessCount int64               `json:"success_count"`
	FailedCount  int64               `json:"failed_count"`
}
