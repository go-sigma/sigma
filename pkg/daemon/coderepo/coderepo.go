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
	"time"

	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/modules/workq"
	"github.com/go-sigma/sigma/pkg/modules/workq/definition"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
)

const (
	perPage = 100
)

func init() {
	workq.TopicHandlers[enums.DaemonCodeRepository.String()] = definition.Consumer{
		Handler:     crRunner,
		MaxRetry:    6,
		Concurrency: 10,
		Timeout:     time.Minute * 10,
	}
}

func crRunner(ctx context.Context, payload []byte) error {
	ctx = log.Logger.WithContext(ctx)
	var task types.DaemonCodeRepositoryPayload
	err := json.Unmarshal(payload, &task)
	if err != nil {
		return fmt.Errorf("Code repository unmarshal payload failed: %v", err)
	}
	cr := codeRepository{
		userServiceFactory:           dao.NewUserServiceFactory(),
		codeRepositoryServiceFactory: dao.NewCodeRepositoryServiceFactory(),
	}

	status := enums.TaskCommonStatusSuccess
	statusMessage := ""
	err = cr.runner(ctx, task)
	if err != nil {
		status = enums.TaskCommonStatusFailed
		statusMessage = err.Error()
	}
	userService := dao.NewUserServiceFactory().New()
	err = userService.UpdateUser3rdParty(ctx, task.User3rdPartyID, map[string]any{
		query.User3rdParty.CrLastUpdateTimestamp.ColumnName().String(): time.Now(),
		query.User3rdParty.CrLastUpdateStatus.ColumnName().String():    status,
		query.User3rdParty.CrLastUpdateMessage.ColumnName().String():   statusMessage,
	})
	if err != nil {
		return err
	}
	return nil
}

type codeRepository struct {
	userServiceFactory           dao.UserServiceFactory
	codeRepositoryServiceFactory dao.CodeRepositoryServiceFactory
}

func (cr codeRepository) runner(ctx context.Context, payload types.DaemonCodeRepositoryPayload) error {
	userService := cr.userServiceFactory.New()
	// TODO: fix get user 3rdparty
	user3rdPartyObj, err := userService.GetUser3rdParty(ctx, payload.User3rdPartyID)
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
