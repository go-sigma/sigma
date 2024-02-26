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

package models

import (
	"gorm.io/plugin/soft_delete"

	"github.com/go-sigma/sigma/pkg/types/enums"
)

// Webhook ...
type Webhook struct {
	CreatedAt int64                 `gorm:"autoUpdateTime:milli"`
	UpdatedAt int64                 `gorm:"autoUpdateTime:milli"`
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	NamespaceID *int64
	Namespace   *Namespace

	URL             string
	Secret          *string
	SslVerify       bool
	RetryTimes      int
	RetryDuration   int
	Enable          bool
	EventNamespace  *bool
	EventRepository bool
	EventTag        bool
	EventArtifact   bool
	EventMember     bool
}

// WebhookLog ...
type WebhookLog struct {
	CreatedAt int64                 `gorm:"autoUpdateTime:milli"`
	UpdatedAt int64                 `gorm:"autoUpdateTime:milli"`
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	WebhookID int64
	Webhook   Webhook

	ResourceType enums.WebhookResourceType
	Action       enums.WebhookAction
	StatusCode   int
	ReqHeader    []byte
	ReqBody      []byte
	RespHeader   []byte
	RespBody     []byte
}
