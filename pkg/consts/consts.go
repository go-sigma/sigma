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

package consts

import (
	"fmt"
	"regexp"
	"time"

	pwdvalidate "github.com/wagslane/go-password-validator"

	"github.com/go-sigma/sigma/pkg/version"
)

const (
	// AppName represents the app name
	AppName = "sigma"
	// APIVersion represents the API version
	APIVersion = "v2"
	// APIVersionKey represents the API version key
	APIVersionKey = "Docker-Distribution-API-Version"
	// APIVersionValue represents the API version value
	APIVersionValue = "registry/2.0"
	// UploadUUID represents the upload uuid in header
	UploadUUID = "Docker-Upload-UUID"
	// ContentDigest represents the content digest in header
	ContentDigest = "Docker-Content-Digest"
	// Blobs represents blobs in storage.
	// file always represent like: sigma/blobs/xx/xx/{digest}
	Blobs = "sigma/blobs"
	// Manifests represents manifests in storage.
	// file always represent like: sigma/manifests/xx/xx/{digest}
	Manifests = "sigma/manifests"
	// Vulnerabilities represents vulnerability reports in storage.
	// file always represent like: sigma/vulnerabilities/{artifact_id}.json.gz
	Vulnerabilities = "sigma/vulnerabilities"
	// BlobUploads represent blob uploads
	// file always represent like: blob_uploads/{upload_id}
	BlobUploads = "blob_uploads"
	// DirCache for the image build cache
	// file always represent like: caches/{builder_id}/{runner_id}
	DirCache = "caches"
	// BlobUploadParts represent blob upload parts
	BlobUploadParts = "blob_upload_parts"
	// BuilderLogs represent builder logs
	BuilderLogs = "builder_logs"
	// DefaultTimePattern time pattern
	DefaultTimePattern = "2006-01-02 15:04:05"
	// ContextJti represents jti in context
	ContextJti = "jti"
	// ContextUser represents user in context
	ContextUser = "ctx-user"
	// HotNamespace top hot namespaces
	HotNamespace = 3
	// WebhookSecretHeader is the HTTP header carrying the hex-encoded HMAC-SHA256 signature of a webhook payload.
	WebhookSecretHeader = "X-Sigma-Signature-256" // nolint: gosec
	// InsertBatchSize is the number of records written per statement when inserting rows in bulk.
	InsertBatchSize = 10
	// MaxNamespaceMember is the maximum number of members a single namespace may have.
	MaxNamespaceMember = 10
	// MaxWebhooks is the maximum number of webhooks a single namespace may have.
	MaxWebhooks = 5
	// ObsPresignMaxTtl
	ObsPresignMaxTtl = time.Minute * 30
	// PprofPath is the URL path where the pprof profiling handlers are mounted when debug logging is enabled.
	PprofPath = "/__debug/pprof"
)

const (
	// UserInternal is used by internal background tasks
	UserInternal = "sigma-internal"
	// UserAnonymous used for anonymous login, just have read permission
	UserAnonymous = "sigma-anonymous"
)

// UserAgent represents the user agent
var UserAgent = fmt.Sprintf("sigma/%s (https://github.com/go-sigma/sigma)", version.Version)

var (
	// PwdStrength represents the password strength
	PwdStrength = pwdvalidate.GetEntropy("Admin@123")
	// Alphanum alphabet num
	Alphanum = "abcdefghijklmnopqrstuvwxyz0123456789"
)

var (
	// TagRegexp matches valid tag names. From [docker/docker:graph/tags.go].
	//
	// [docker/docker:graph/tags.go]: https://github.com/moby/moby/blob/v1.6.0/graph/tags.go#L26-L28
	TagRegexp = regexp.MustCompile(`^[\w][\w.-]{0,127}$`)
)

const (
	// CacherBlob is the cache prefix under which blob metadata entries are stored.
	CacherBlob = "blob"
	// CacherManifest is the cache prefix under which manifest entries are stored.
	CacherManifest = "manifest"
)

const (
	// APIV1 api v1 for api router
	APIV1 = "/api/v1"
)

const (
	// ServerPort server port
	ServerPort = "0.0.0.0:3000"
	// WorkerPort worker port
	WorkerPort = "0.0.0.0:3001"
	// DistributionPort distribution port
	DistributionPort = "0.0.0.0:3002"
)

const (
	// LockerMigration is the distributed lock key serializing database migration runs.
	LockerMigration = "locker-migration"
	// LockerCronjobBuilder is the distributed lock key prefix serializing builder cronjob runs; a builder ID is appended to it.
	LockerCronjobBuilder = "locker-cronjob-builder"
	// LockerAnalyticsFlush is the distributed lock key prefix serializing analytics flush runs; the target hour is appended to it.
	LockerAnalyticsFlush = "locker-analytics-flush"
	// LockerCronjobAudit is the distributed lock key serializing audit log cleanup runs.
	LockerCronjobAudit = "locker-cronjob-audit"
	// LockerCronjobSize is the distributed lock key serializing namespace size reconciliation runs.
	LockerCronjobSize = "locker-cronjob-size"
	// LockerVulnerabilityScan is the distributed lock key serializing vulnerability scan runs.
	LockerVulnerabilityScan = "locker-vulnerability-scan"
)

var (
	// KeepNamespaces namespace keep name, any name expect these names
	KeepNamespaces = []string{"api", "v2"}
)

const (
	// LockerRetryDelay is the delay between successive attempts to acquire a distributed lock.
	LockerRetryDelay = time.Second * 1
	// LockerRetryMaxTimes is the maximum number of attempts made to acquire a distributed lock before giving up.
	LockerRetryMaxTimes = 6
)
