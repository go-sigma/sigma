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

package builder

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"path"
	"reflect"
	"strings"
	"time"

	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/background/build/runtime"
	"github.com/go-sigma/sigma/pkg/background/daemon"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/dal/query"
	repobuilder "github.com/go-sigma/sigma/pkg/dal/repository/builder"
	reporegistry "github.com/go-sigma/sigma/pkg/dal/repository/registry"
	"github.com/go-sigma/sigma/pkg/infra/workq"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

func init() {
	utils.PanicIf(daemon.Daemons.Register(path.Base(reflect.TypeFor[factory]().PkgPath()), &factory{}))
}

type factory struct{}

type params struct {
	dig.In

	HandlerRegistry      workq.HandlerRegistry
	BuilderRepository    repobuilder.BuilderRepository
	RepositoryRepository reporegistry.RepositoryRepository
}

// Initialize initializes the builder daemon, which is used to build the image.
func (f factory) Initialize(digCon *dig.Container) error {
	var p params
	if err := digCon.Invoke(func(deps params) { p = deps }); err != nil {
		return err
	}
	return p.HandlerRegistry.Register(enums.DaemonBuilder, workq.Consumer{
		Handler:     builderRunner(p),
		Concurrency: 10,
		Timeout:     time.Minute * 60,
	})
}

func builderRunner(p params) func(ctx context.Context, data []byte) error {
	return func(ctx context.Context, data []byte) error {
		var payload api.DaemonBuilderPayload
		err := json.Unmarshal(data, &payload)
		if err != nil {
			return fmt.Errorf("unmarshal payload failed: %v", err)
		}
		b := taskRunner{
			builderRepository:    p.BuilderRepository,
			repositoryRepository: p.RepositoryRepository,
		}
		return b.run(ctx, payload)
	}
}

type taskRunner struct {
	builderRepository    repobuilder.BuilderRepository
	repositoryRepository reporegistry.RepositoryRepository
}

func (b taskRunner) run(ctx context.Context, payload api.DaemonBuilderPayload) error {
	if payload.Action == enums.DaemonBuilderActionStop {
		return runtime.Driver.Stop(ctx, payload.BuilderID, payload.RunnerID)
	}
	repositoryRepository := b.repositoryRepository
	repositoryObj, err := repositoryRepository.Get(ctx, payload.RepositoryID)
	if err != nil {
		slog.Error("get repository record failed", "err", err, "id", payload.RepositoryID)
		return fmt.Errorf("get repository record failed")
	}
	builderRepository := b.builderRepository
	builderObj, err := builderRepository.GetByRepositoryID(ctx, payload.RepositoryID)
	if err != nil {
		slog.Error("get builder record failed", "err", err, "id", payload.RepositoryID)
		return fmt.Errorf("get builder record failed")
	}

	runnerObj, err := builderRepository.GetRunner(ctx, payload.RunnerID)
	if err != nil {
		slog.Error("get runner failed", "err", err)
		return fmt.Errorf("get runner failed: %v", err)
	}

	defer func() {
		var updates map[string]any
		if !(payload.Action == enums.DaemonBuilderActionStart || payload.Action == enums.DaemonBuilderActionRestart || payload.Action == enums.DaemonBuilderActionStop) { // nolint: staticcheck
			updates = map[string]any{
				query.BuilderRunner.Status.ColumnName().String():        enums.BuildStatusFailed,
				query.BuilderRunner.StatusMessage.ColumnName().String(): fmt.Sprintf("Daemon builder action(%s) is not support", payload.Action),
				query.BuilderRunner.EndedAt.ColumnName().String():       time.Now().UnixMilli(),
			}
		}
		if err != nil {
			updates = map[string]any{
				query.BuilderRunner.Status.ColumnName().String():        enums.BuildStatusFailed,
				query.BuilderRunner.StatusMessage.ColumnName().String(): err.Error(),
				query.BuilderRunner.EndedAt.ColumnName().String():       time.Now().UnixMilli(),
			}
		}
		if len(updates) > 0 {
			err = builderRepository.UpdateRunner(ctx, payload.BuilderID, payload.RunnerID, updates)
			if err != nil {
				slog.Error("update runner after got error", "err", err)
			}
		}
	}()

	if runtime.Driver == nil {
		err = fmt.Errorf("buildrunner driver is not initialized")
		return fmt.Errorf("buildrunner driver is not initialized, or check config.daemon.builder.enabled is true or not")
	}

	platforms := []enums.OciPlatform{}
	for p := range strings.SplitSeq(builderObj.BuildkitPlatforms, ",") {
		platforms = append(platforms, enums.OciPlatform(p))
	}

	buildConfig := runtime.Config{
		Builder: api.Builder{
			BuilderID: payload.BuilderID,
			RunnerID:  runnerObj.ID,

			Repository: base64.StdEncoding.EncodeToString([]byte(repositoryObj.Name)),
			Tag:        base64.StdEncoding.EncodeToString([]byte(runnerObj.RawTag)),

			Source: runnerObj.Builder.Source,

			Dockerfile: new(base64.StdEncoding.EncodeToString(runnerObj.Builder.Dockerfile)),

			// ScmCredentialType: builderObj.ScmCredentialType,
			// ScmProvider: enums.ScmProviderGithub,
			// ScmSshKey:         builderObj.ScmSshKey,
			// ScmToken:          builderObj.ScmToken,
			// ScmUsername:       builderObj.ScmUsername,
			// ScmPassword:       builderObj.ScmPassword,
			// ScmRepository:     builderObj.ScmRepository,
			ScmBranch:    runnerObj.ScmBranch,
			ScmDepth:     builderObj.ScmDepth,
			ScmSubmodule: builderObj.ScmSubmodule,

			// OciRegistryDomain:   []string{"192.168.31.198:3000"},
			// OciRegistryUsername: []string{"sigma"},
			// OciRegistryPassword: []string{"sigma"},
			// OciName: "192.168.31.198:3000/library/test:dev",

			BuildkitPlatforms:          platforms,
			BuildkitInsecureRegistries: strings.Split(builderObj.BuildkitInsecureRegistries, ","), //  []string{"192.168.31.198:3000@http"},
		},
		ExtraHosts: config.GetConfig().Daemon.Builder.ExtraHosts,
	}
	if builderObj.Source == enums.BuilderSourceCodeRepository {
		buildConfig.Builder.ScmCredentialType = builderObj.ScmCredentialType // nolint: staticcheck

		switch ptr.To(builderObj.ScmCredentialType) {
		case enums.ScmCredentialTypeSsh:
			buildConfig.Builder.ScmSshKey = builderObj.ScmSshKey // nolint: staticcheck
			if builderObj.CodeRepository != nil {
				buildConfig.Builder.ScmRepository = new(builderObj.CodeRepository.SshUrl) // nolint: staticcheck
			}
		case enums.ScmCredentialTypeToken:
			buildConfig.Builder.ScmToken = builderObj.ScmToken // nolint: staticcheck
			if builderObj.CodeRepository != nil {
				buildConfig.Builder.ScmRepository = new(builderObj.CodeRepository.CloneUrl) // nolint: staticcheck
			}
		case enums.ScmCredentialTypeUsername:
			buildConfig.Builder.ScmUsername = builderObj.ScmUsername // nolint: staticcheck
			buildConfig.Builder.ScmPassword = builderObj.ScmPassword // nolint: staticcheck

			if builderObj.CodeRepository != nil {
				buildConfig.Builder.ScmRepository = new(builderObj.CodeRepository.CloneUrl) // nolint: staticcheck
			}
		}

		// nolint: staticcheck
		buildConfig.Builder.ScmProvider = (*enums.ScmProvider)(&builderObj.CodeRepository.User3rdParty.Provider) // TODO: change type
	}

	if payload.Action == enums.DaemonBuilderActionStart || payload.Action == enums.DaemonBuilderActionRestart {
		err = runtime.Driver.Start(ctx, buildConfig)
		if err != nil {
			slog.Error("start or restart builder failed", "err", err)
			return fmt.Errorf("start or restart builder failed: %v", err)
		}
	}

	return nil
}
