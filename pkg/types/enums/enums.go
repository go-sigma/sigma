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

// BuildStatus x ENUM(
// Success,
// Failed,
// Pending,
// Scheduling,
// Building,
// )
type BuildStatus string

// Database x ENUM(
// postgresql,
// mysql,
// sqlite3,
// )
type Database string

// RedisType x ENUM(
// internal,
// external,
// )
type RedisType string

// CacheType x ENUM(
// memory,
// redis,
// )
type CacheType string

// CacheKey x ENUM(
// redis,
// )
type WorkQueueType string

// Daemon x ENUM(
// Vulnerability,
// Sbom,
// Gc,
// GcRepository,
// Webhook,
// Builder,
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
// )
type Provider string

// SortMethod x ENUM(
// asc,
// desc,
// )
type SortMethod string

// ArtifactType x ENUM(
// image,
// imageIndex,
// chart
// cnab,
// wasm,
// provenance,
// unknown,
// )
type ArtifactType string

// AuditAction x ENUM(
// create,
// update,
// delete,
// pull,
// push,
// )
type AuditAction string

// AuditResourceType x ENUM(
// namespace,
// repository,
// tag,
// webhook,
// builder,
// )
type AuditResourceType string

// WebhookResourceType x ENUM(
// ping,
// namespace,
// repository,
// tag,
// artifact,
// member,
// )
type WebhookResourceType string

// WebhookOperate x ENUM(
// create,
// update,
// delete,
// add,
// remove,
// pull,
// push,
// )
type WebhookResourceAction string

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
