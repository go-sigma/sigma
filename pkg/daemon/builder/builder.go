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
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/builder"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/modules/workq"
	"github.com/go-sigma/sigma/pkg/modules/workq/definition"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

func init() {
	workq.TopicHandlers[enums.DaemonBuilder] = definition.Consumer{
		Handler:     builderRunner,
		MaxRetry:    1,
		Concurrency: 10,
		Timeout:     time.Minute * 60,
	}
}

func builderRunner(ctx context.Context, data []byte) error {
	var payload types.DaemonBuilderPayload
	err := json.Unmarshal(data, &payload)
	if err != nil {
		return fmt.Errorf("Unmarshal payload failed: %v", err)
	}
	b := runner{
		builderServiceFactory:    dao.NewBuilderServiceFactory(),
		repositoryServiceFactory: dao.NewRepositoryServiceFactory(),
	}
	return b.runner(ctx, payload)
}

type runner struct {
	builderServiceFactory    dao.BuilderServiceFactory
	repositoryServiceFactory dao.RepositoryServiceFactory
}

func (b runner) runner(ctx context.Context, payload types.DaemonBuilderPayload) error {
	ctx = log.Logger.WithContext(ctx)

	if payload.Action == enums.DaemonBuilderActionStop {
		return builder.Driver.Stop(ctx, payload.BuilderID, payload.RunnerID)
	}
	repositoryService := b.repositoryServiceFactory.New()
	repositoryObj, err := repositoryService.Get(ctx, payload.RepositoryID)
	if err != nil {
		log.Error().Err(err).Int64("id", payload.RepositoryID).Msg("Get repository record failed")
		return fmt.Errorf("Get repository record failed")
	}
	builderService := b.builderServiceFactory.New()
	builderObj, err := builderService.GetByRepositoryID(ctx, payload.RepositoryID)
	if err != nil {
		log.Error().Err(err).Int64("id", payload.RepositoryID).Msg("Get builder record failed")
		return fmt.Errorf("Get builder record failed")
	}

	runnerObj, err := builderService.GetRunner(ctx, payload.RunnerID)
	if err != nil {
		log.Error().Err(err).Msg("Get runner failed")
		return fmt.Errorf("Get runner failed: %v", err)
	}

	defer func() {
		var updates map[string]any
		if !(payload.Action == enums.DaemonBuilderActionStart || payload.Action == enums.DaemonBuilderActionRestart || payload.Action == enums.DaemonBuilderActionStop) {
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
			err = builderService.UpdateRunner(ctx, payload.BuilderID, payload.RunnerID, updates)
			if err != nil {
				log.Error().Err(err).Msg("Update runner after got error")
			}
		}
	}()

	if builder.Driver == nil {
		err = fmt.Errorf("Builder driver is not initialized")
		return fmt.Errorf("Builder driver is not initialized, or check config.daemon.builder.enabled is true or not")
	}

	platforms := []enums.OciPlatform{}
	for _, p := range strings.Split(builderObj.BuildkitPlatforms, ",") {
		platforms = append(platforms, enums.OciPlatform(p))
	}

	buildConfig := builder.BuilderConfig{
		Builder: types.Builder{
			BuilderID: payload.BuilderID,
			RunnerID:  runnerObj.ID,

			Repository: base64.StdEncoding.EncodeToString([]byte(repositoryObj.Name)),
			Tag:        base64.StdEncoding.EncodeToString([]byte(runnerObj.RawTag)),

			Source: runnerObj.Builder.Source,

			Dockerfile: ptr.Of(base64.StdEncoding.EncodeToString(runnerObj.Builder.Dockerfile)),

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
	}
	if builderObj.Source == enums.BuilderSourceCodeRepository {
		buildConfig.Builder.ScmCredentialType = builderObj.ScmCredentialType

		switch ptr.To(builderObj.ScmCredentialType) {
		case enums.ScmCredentialTypeSsh:
			buildConfig.Builder.ScmSshKey = builderObj.ScmSshKey
			if builderObj.CodeRepository != nil {
				buildConfig.Builder.ScmRepository = ptr.Of(builderObj.CodeRepository.SshUrl)
			}
		case enums.ScmCredentialTypeToken:
			buildConfig.Builder.ScmToken = builderObj.ScmToken
			if builderObj.CodeRepository != nil {
				buildConfig.Builder.ScmRepository = ptr.Of(builderObj.CodeRepository.CloneUrl)
			}
		case enums.ScmCredentialTypeUsername:
			buildConfig.Builder.ScmUsername = builderObj.ScmUsername
			buildConfig.Builder.ScmPassword = builderObj.ScmPassword

			if builderObj.CodeRepository != nil {
				buildConfig.Builder.ScmRepository = ptr.Of(builderObj.CodeRepository.CloneUrl)
			}
		}

		buildConfig.Builder.ScmProvider = (*enums.ScmProvider)(&builderObj.CodeRepository.User3rdParty.Provider) // TODO: change type
	}

	if payload.Action == enums.DaemonBuilderActionStart || payload.Action == enums.DaemonBuilderActionRestart {
		err = builder.Driver.Start(ctx, buildConfig)
		if err != nil {
			log.Error().Err(err).Msg("Start or restart builder failed")
			return fmt.Errorf("Start or restart builder failed: %v", err)
		}
	}

	return nil
}
