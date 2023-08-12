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

package coderepo

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/daemon"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
)

func init() {
	utils.PanicIf(daemon.RegisterTask(enums.DaemonCodeRepository, crRunner))
}

func crRunner(ctx context.Context, task *asynq.Task) error {
	ctx = log.Logger.WithContext(ctx)
	var payload types.DaemonCodeRepositoryPayload
	err := json.Unmarshal(task.Payload(), &payload)
	if err != nil {
		return fmt.Errorf("Code repository unmarshal payload failed: %v", err)
	}
	cr := codeRepository{
		userServiceFactory:           dao.NewUserServiceFactory(),
		codeRepositoryServiceFactory: dao.NewCodeRepositoryServiceFactory(),
	}
	return cr.runner(ctx, payload)
}

type codeRepository struct {
	userServiceFactory           dao.UserServiceFactory
	codeRepositoryServiceFactory dao.CodeRepositoryServiceFactory
}

func (cr codeRepository) runner(ctx context.Context, payload types.DaemonCodeRepositoryPayload) error {
	userService := cr.userServiceFactory.New()
	user3rdPartyObj, err := userService.GetUser3rdParty(ctx, payload.UserID)
	if err != nil {
		log.Error().Err(err).Msg("Get 3rdParty user failed")
		return fmt.Errorf("Get 3rdParty user failed: %v", err)
	}
	switch user3rdPartyObj.Provider {
	case enums.ProviderGithub:
		return cr.github(ctx, user3rdPartyObj)
	case enums.ProviderGitlab:
		return cr.gitlab(ctx, user3rdPartyObj)
	case enums.ProviderGitea:
		return cr.gitea(ctx, user3rdPartyObj)
	}
	return nil
}
