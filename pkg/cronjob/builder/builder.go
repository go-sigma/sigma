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

package cronjob

import (
	"context"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/cronjob"
	"github.com/go-sigma/sigma/pkg/daemon"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/modules/locker"
	"github.com/go-sigma/sigma/pkg/modules/timewheel"
	"github.com/go-sigma/sigma/pkg/service/builder"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

var builderTw timewheel.TimeWheel

func init() {
	cronjob.Starter = append(cronjob.Starter, builderJob)
	cronjob.Stopper = append(cronjob.Stopper, func() {
		if builderTw != nil {
			builderTw.Stop()
		}
	})
}

func builderJob() {
	builderTw = timewheel.NewTimeWheel(context.Background(), cronjob.CronjobIterDuration)

	runner := builderRunner{
		builderServiceFactory: dao.NewBuilderServiceFactory(),
	}
	builderTw.AddRunner(runner.runner)
}

type builderRunner struct {
	builderServiceFactory dao.BuilderServiceFactory
}

func (r builderRunner) runner(ctx context.Context, tw timewheel.TimeWheel) {
	locker, err := locker.New()
	if err != nil {
		log.Error().Err(err).Msg("New locker failed")
		return
	}
	lock, err := locker.Lock(context.Background(), consts.LockerMigration, time.Second*30)
	if err != nil {
		log.Error().Err(err).Msg("Cronjob builder get locker failed")
		return
	}
	defer func() {
		err := lock.Unlock()
		if err != nil {
			log.Error().Err(err).Msg("Migrate locker release failed")
		}
	}()

	ctx = log.Logger.WithContext(ctx)
	builderService := r.builderServiceFactory.New()
	builderObjs, err := builderService.GetByNextTrigger(ctx, time.Now(), cronjob.MaxJob)
	if err != nil {
		log.Error().Err(err).Msg("Get builders by next trigger failed")
		return
	}
	cronParser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	for _, builderObj := range builderObjs {
		// do transaction:
		// 1. update the next trigger time
		// 2. publish the job
		err = query.Q.Transaction(func(tx *query.Query) error {
			builderService := r.builderServiceFactory.New(tx)
			schedule, err := cronParser.Parse(ptr.To(builderObj.CronRule))
			if err != nil {
				return err
			}
			err = builderService.UpdateNextTrigger(ctx, builderObj.ID, schedule.Next(time.Now()))
			if err != nil {
				return err
			}
			tag, err := builder.BuildTag(ptr.To(builderObj.CronTagTemplate), builder.BuildTagOption{ScmBranch: ptr.To(builderObj.CronBranch)})
			if err != nil {
				return err
			}
			runner, err := builder.BuildRunner(builderObj, builder.BuildRunnerOption{
				Tag:       tag,
				ScmBranch: builderObj.CronBranch,
			})
			if err != nil {
				return err
			}
			err = builderService.CreateRunner(ctx, runner)
			if err != nil {
				return err
			}
			builderJob := &types.DaemonBuilderPayload{
				Action:    enums.DaemonBuilderActionStart,
				BuilderID: builderObj.ID,
				RunnerID:  runner.ID,
			}
			err = daemon.Enqueue(consts.TopicBuilder, utils.MustMarshal(builderJob))
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			log.Error().Interface("builder", builderObj).Err(err).Msg("Cronjob create builder runner failed")
		}
	}
	if len(builderObjs) >= cronjob.MaxJob {
		tw.TickNext(cronjob.TickNextDuration)
	}
}
