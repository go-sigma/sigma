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
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"

	builderdriver "github.com/go-sigma/sigma/pkg/builder"
	"github.com/go-sigma/sigma/pkg/daemon"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
)

func init() {
	utils.PanicIf(daemon.RegisterTask(enums.DaemonBuilder, builderRunner))
}

func builderRunner(ctx context.Context, task *asynq.Task) error {
	var payload types.DaemonBuilderPayload
	err := json.Unmarshal(task.Payload(), &payload)
	if err != nil {
		return fmt.Errorf("Unmarshal payload failed: %v", err)
	}
	b := builder{
		builderServiceFactory: dao.NewBuilderServiceFactory(),
	}
	return b.runner(ctx, payload)
}

type builder struct {
	builderServiceFactory dao.BuilderServiceFactory
}

func (b builder) runner(ctx context.Context, payload types.DaemonBuilderPayload) error {
	if payload.Action == enums.DaemonBuilderActionStop {
		return builderdriver.Driver.Stop(ctx, payload.BuilderID, payload.RunnerID)
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

	buildConfig := builderdriver.BuilderConfig{
		Builder: types.Builder{
			BuilderID: payload.BuilderID,
			RunnerID:  runnerObj.ID,

			// ScmCredentialType: builderObj.ScmCredentialType,
			// ScmProvider:       enums.ScmProviderGithub,
			// ScmSshKey:         builderObj.ScmSshKey,
			// ScmToken:          builderObj.ScmToken,
			// ScmUsername:       builderObj.ScmUsername,
			// ScmPassword:       builderObj.ScmPassword,
			// ScmRepository:     builderObj.ScmRepository,
			ScmBranch:    runnerObj.ScmBranch,
			ScmDepth:     builderObj.ScmDepth,
			ScmSubmodule: builderObj.ScmSubmodule,

			OciRegistryDomain:   []string{"192.168.31.114:3000"},
			OciRegistryUsername: []string{"sigma"},
			OciRegistryPassword: []string{"sigma"},
			OciName:             "192.168.31.114:3000/library/test:dev",

			BuildkitInsecureRegistries: []string{"192.168.31.114:3000@http"},
		},
	}
	if payload.Action == enums.DaemonBuilderActionStart { // nolint: gocritic
		err = builderdriver.Driver.Start(ctx, buildConfig)
	} else if payload.Action == enums.DaemonBuilderActionRestart {
		err = builderdriver.Driver.Start(ctx, buildConfig)
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
