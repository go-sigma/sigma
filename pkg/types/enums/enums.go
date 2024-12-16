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

package enums

//go:generate go-enum --sql --mustparse

// LogLevel x ENUM(
// trace,
// debug,
// info,
// warn,
// error,
// fatal,
// panic,
// )
type LogLevel string

// Deploy x ENUM(
// single,
// replica,
// )
type Deploy string

// TaskCommonStatus x ENUM(
// Pending,
// Doing,
// Success,
// Failed,
// )
type TaskCommonStatus string

// GcRecordStatus x ENUM(
// Success,
// Failed,
// )
type GcRecordStatus string

// BuildStatus x ENUM(
// Success,
// Failed,
// Pending,
// Scheduling,
// Building,
// Stopping,
// Stopped,
// )
type BuildStatus string

// Database x ENUM(
// postgresql,
// mysql,
// sqlite3,
// )
type Database string

// RedisType x ENUM(
// none,
// external,
// )
type RedisType string

// CacherType x ENUM(
// inmemory,
// redis,
// badger,
// )
type CacherType string

// WorkQueueType x ENUM(
// redis,
// kafka,
// database,
// inmemory,
// )
type WorkQueueType string

// LockerType x ENUM(
// redis,
// badger,
// )
type LockerType string

// Daemon x ENUM(
// Vulnerability,
// Sbom,
// Gc,
// GcRepository,
// GcArtifact,
// GcBlob,
// GcTag,
// Webhook,
// Builder,
// CodeRepository,
// TagPushed,
// ArtifactPushed,
// )
type Daemon string

// Visibility x ENUM(
// private,
// public,
// )
type Visibility string

// GcTarget x ENUM(
// blobsAndArtifacts,
// artifacts,
// )
type GcTarget string

// Provider x ENUM(
// local,
// github,
// gitlab,
// gitea,
// )
type Provider string

// SortMethod x ENUM(
// asc,
// desc,
// )
type SortMethod string

// ArtifactType x ENUM(
// Image,
// ImageIndex,
// Chart
// Cnab,
// Wasm,
// Provenance,
// Cosign,
// Sif,
// Unknown,
// )
type ArtifactType string

// AuditAction x ENUM(
// Create,
// Update,
// Delete,
// Pull,
// Push,
// )
type AuditAction string

// AuditResourceType x ENUM(
// Namespace,
// NamespaceMember,
// Repository,
// Tag,
// Webhook,
// Builder,
// )
type AuditResourceType string

// WebhookResourceType x ENUM(
// Webhook,
// Namespace,
// Repository,
// Tag,
// Artifact,
// Member,
// DaemonTaskGcRepositoryRule,
// DaemonTaskGcTagRule,
// DaemonTaskGcArtifactRule,
// DaemonTaskGcBlobRule,
// DaemonTaskGcRepositoryRunner,
// DaemonTaskGcTagRunner,
// DaemonTaskGcArtifactRunner,
// DaemonTaskGcBlobRunner,
// )
type WebhookResourceType string

// WebhookAction x ENUM(
// Create,
// Update,
// Delete,
// Add,
// Remove,
// Ping,
// Started,
// Doing,
// Finished,
// )
type WebhookAction string

// WebhookType x ENUM(
// Ping,
// Send,
// Resend,
// )
type WebhookType string

// ScmCredentialType x ENUM(
// ssh,
// token,
// username,
// none,
// )
type ScmCredentialType string

// ScmProvider x ENUM(
// github,
// gitlab,
// gitea,
// none,
// )
type ScmProvider string

// OciPlatform x ENUM(
// linux/amd64,
// linux/amd64/v2,
// linux/amd64/v3,
// linux/arm64,
// linux/riscv64,
// linux/ppc64le,
// linux/s390x,
// linux/386,
// linux/mips64le,
// linux/mips64,
// linux/arm/v7,
// linux/arm/v6,
// )
type OciPlatform string

// DaemonBuilderAction x ENUM(
// Start,
// Restart,
// Stop,
// )
type DaemonBuilderAction string

// BuilderSource x ENUM(
// Dockerfile,
// CodeRepository,
// SelfCodeRepository,
// )
type BuilderSource string

// StorageType x ENUM(
// s3,
// filesystem,
// cos,
// oss,
// dummy,
// )
type StorageType string

// BuilderType x ENUM(
// docker,
// kubernetes,
// )
type BuilderType string

// SigningType x ENUM(
// cosign,
// )
type SigningType string

// UserStatus x ENUM(
// Active,
// Inactive,
// )
type UserStatus string

// Root is available for user role, but it just create by initialized data

// UserRole x ENUM(
// Root,
// Admin,
// User,
// Anonymous,
// )
type UserRole string

// RetentionRuleType x ENUM(
// Day,
// Quantity,
// )
type RetentionRuleType string

// NamespaceRole x ENUM(
// Admin="NamespaceAdmin",
// Manager="namespace_manager",
// Reader="NamespaceReader",
// )
type NamespaceRole string

// Auth x ENUM(
// Read,
// Manage,
// Admin,
// )
type Auth string

// OperateType x ENUM(
// Manual,
// Automatic,
// )
type OperateType string
