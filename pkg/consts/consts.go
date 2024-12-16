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
	// Blobs represents a blobs
	// file always represent like: blobs/{algo}/xx/xx/{digest}
	Blobs = "blobs"
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
	// WebhookSecretHeader ...
	WebhookSecretHeader = "X-Sigma-Signature-256" // nolint: gosec
	// InsertBatchSize ...
	InsertBatchSize = 10
	// MaxNamespaceMember ...
	MaxNamespaceMember = 10
	// MaxWebhooks ...
	MaxWebhooks = 5
	// ObsPresignMaxTtl
	ObsPresignMaxTtl = time.Minute * 30
	// PprofPath ...
	PprofPath = "/__debug/pprof"
	// BuilderImagePath ...
	BuilderImagePath = "/baseimages/"
)

const (
	// UserInternal used for build and push image to sigma
	UserInternal = "sigma-internal"
	// UserAnonymous used for anonymous login, just have read permission
	UserAnonymous = "sigma-anonymous"
)

// UserAgent represents the user agent
var UserAgent = fmt.Sprintf("sigma/%s (https://github.com/go-sigma/sigma)", version.Version)

const (
	// AuthModel represents the auth model
	// policy_effect: it means at least one matched policy rule of allow, and there is no matched policy rule of deny. So in this way, both the allow and deny authorizations are supported, and the deny overrides.
	AuthModel = `
	[request_definition]
	r = sub, ns, url, visibility, method

	[policy_definition]
	p = sub, ns, url, visibility, method, effect

	[role_definition]
	g = _, _, _

	[policy_effect]
	e = some(where (p.eft == allow)) && !some(where (p.eft == deny))

	[matchers]
	m = g(r.sub, p.sub, r.ns) && keyMatch(r.ns, p.ns) && urlMatch(r.url, p.url) && regexMatch(r.visibility, p.visibility) && regexMatch(r.method, p.method) && p.effect == "allow" || r.sub == "admin" || r.sub == "root"`
)

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
	// CacherBlob ...
	CacherBlob = "blob"
	// CacherManifest ...
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
	// RedisPid ...
	RedisPid = "/var/run/redis.pid"
)

const (
	// LockerMigration ...
	LockerMigration = "locker-migration"
	// LockerCronjobBuilder ...
	LockerCronjobBuilder = "locker-cronjob-builder"
	// LockerBaseimage ...
	LockerBaseimage = "locker-baseimage"
)

var (
	// KeepNamespaces namespace keep name, any name expect these names
	KeepNamespaces = []string{"api", "v2"}
)

const (
	// LockerRetryDelay ...
	LockerRetryDelay = time.Second * 1
	// LockerRetryMaxTimes ...
	LockerRetryMaxTimes = 6
)
