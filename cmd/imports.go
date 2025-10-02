// Copyright 2025 sigma
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

package cmd

// nolint: gci
import (
	_ "github.com/go-sigma/sigma/pkg/server/handlers/apidocs"
	_ "github.com/go-sigma/sigma/pkg/server/handlers/artifacts"
	_ "github.com/go-sigma/sigma/pkg/server/handlers/builders"
	_ "github.com/go-sigma/sigma/pkg/server/handlers/caches"
	_ "github.com/go-sigma/sigma/pkg/server/handlers/coderepos"
	_ "github.com/go-sigma/sigma/pkg/server/handlers/daemons"
	_ "github.com/go-sigma/sigma/pkg/server/handlers/namespaces"
	_ "github.com/go-sigma/sigma/pkg/server/handlers/oauth2"
	_ "github.com/go-sigma/sigma/pkg/server/handlers/repositories"
	_ "github.com/go-sigma/sigma/pkg/server/handlers/systems"
	_ "github.com/go-sigma/sigma/pkg/server/handlers/tags"
	_ "github.com/go-sigma/sigma/pkg/server/handlers/tokens"
	_ "github.com/go-sigma/sigma/pkg/server/handlers/users"
	_ "github.com/go-sigma/sigma/pkg/server/handlers/validators"
	_ "github.com/go-sigma/sigma/pkg/server/handlers/webhooks"

	_ "github.com/go-sigma/sigma/pkg/server/handlers/distribution/base"
	_ "github.com/go-sigma/sigma/pkg/server/handlers/distribution/blob"
	_ "github.com/go-sigma/sigma/pkg/server/handlers/distribution/manifest"
	_ "github.com/go-sigma/sigma/pkg/server/handlers/distribution/upload"

	_ "github.com/go-sigma/sigma/pkg/builder/docker"
	_ "github.com/go-sigma/sigma/pkg/builder/kubernetes"
	_ "github.com/go-sigma/sigma/pkg/builder/logger/database"
	_ "github.com/go-sigma/sigma/pkg/builder/logger/obs"
	_ "github.com/go-sigma/sigma/pkg/builder/podman"

	_ "github.com/go-sigma/sigma/pkg/cronjob/builder"

	_ "github.com/go-sigma/sigma/pkg/daemon/builder"
	_ "github.com/go-sigma/sigma/pkg/daemon/coderepo"
	_ "github.com/go-sigma/sigma/pkg/daemon/gc"
	_ "github.com/go-sigma/sigma/pkg/daemon/pushed"
	_ "github.com/go-sigma/sigma/pkg/daemon/scan"
	_ "github.com/go-sigma/sigma/pkg/daemon/webhook"

	_ "github.com/go-sigma/sigma/pkg/modules/locker/badger"
	_ "github.com/go-sigma/sigma/pkg/modules/locker/redis"

	_ "go.uber.org/mock/mockgen/model"

	_ "github.com/distribution/distribution/v3/manifest/manifestlist"
	_ "github.com/distribution/distribution/v3/manifest/ocischema"
	_ "github.com/distribution/distribution/v3/manifest/schema2"

	_ "github.com/go-sigma/sigma/pkg/signing/cosign"

	_ "github.com/go-sigma/sigma/pkg/storage/cos"
	_ "github.com/go-sigma/sigma/pkg/storage/filesystem"
	_ "github.com/go-sigma/sigma/pkg/storage/oss"
	_ "github.com/go-sigma/sigma/pkg/storage/s3"

	_ "github.com/go-sigma/sigma/pkg/modules/workq/database"
	_ "github.com/go-sigma/sigma/pkg/modules/workq/inmemory"
	_ "github.com/go-sigma/sigma/pkg/modules/workq/redis"
)
