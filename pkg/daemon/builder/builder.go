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
	"github.com/go-sigma/sigma/pkg/modules/workq"
	"github.com/go-sigma/sigma/pkg/modules/workq/definition"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

func init() {
	workq.TopicHandlers[enums.DaemonBuilder.String()] = definition.Consumer{
		Handler:     builderRunner,
		MaxRetry:    6,
		Concurrency: 10,
		Timeout:     time.Minute * 10,
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

	// runnerObj := &models.BuilderRunner{
	// 	BuilderID: payload.BuilderID,
	// 	Status:    enums.BuildStatusPending,
	// }
	// err = builderService.CreateRunner(ctx, runnerObj)
	// if err != nil {
	// 	log.Error().Err(err).Msg("Create builder runner record failed")
	// 	return fmt.Errorf("Create builder runner record failed: %v", err)
	// }

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
			// ScmProvider:       enums.ScmProviderGithub,
			// ScmSshKey:         builderObj.ScmSshKey,
			// ScmToken:          builderObj.ScmToken,
			// ScmUsername:       builderObj.ScmUsername,
			// ScmPassword:       builderObj.ScmPassword,
			// ScmRepository:     builderObj.ScmRepository,
			ScmBranch: runnerObj.ScmBranch,
			ScmDepth:  builderObj.ScmDepth,
			// ScmSubmodule: builderObj.ScmSubmodule,

			// OciRegistryDomain:   []string{"192.168.31.198:3000"},
			// OciRegistryUsername: []string{"sigma"},
			// OciRegistryPassword: []string{"sigma"},
			// OciName: "192.168.31.198:3000/library/test:dev",

			BuildkitPlatforms:          platforms,
			BuildkitInsecureRegistries: strings.Split(builderObj.BuildkitInsecureRegistries, ","), //  []string{"192.168.31.198:3000@http"},
		},
	}
	if payload.Action == enums.DaemonBuilderActionStart { // nolint: gocritic
		err = builder.Driver.Start(ctx, buildConfig)
	} else if payload.Action == enums.DaemonBuilderActionRestart {
		err = builder.Driver.Start(ctx, buildConfig)
	} else {
		log.Error().Err(err).Str("action", payload.Action.String()).Msg("Daemon builder action not found")
		return fmt.Errorf("Daemon builder action not found")
	}
	if err != nil {
		log.Error().Err(err).Msg("Start or restart builder failed")
		return fmt.Errorf("Start or restart builder failed: %v", err)
	}
	return nil
}
