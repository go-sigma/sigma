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

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/daemon"
	"github.com/go-sigma/sigma/pkg/dal"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/modules/timewheel"
	"github.com/go-sigma/sigma/pkg/service/builder"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

var builderTw timewheel.TimeWheel

func init() {
	starter = append(starter, builderJob)
	stopper = append(stopper, func() {
		if builderTw != nil {
			builderTw.Stop()
		}
	})
}

func builderJob() {
	builderTw = timewheel.NewTimeWheel(context.Background(), cronjobIterDuration)

	runner := builderRunner{
		builderServiceFactory: dao.NewBuilderServiceFactory(),
	}
	builderTw.AddRunner(runner.runner)
}

type builderRunner struct {
	builderServiceFactory dao.BuilderServiceFactory
}

func (r builderRunner) runner(ctx context.Context, tw timewheel.TimeWheel) {
	rs := redsync.New(goredis.NewPool(dal.RedisCli))
	mutex := rs.NewMutex(consts.LockerCronjobBuilder, redsync.WithRetryDelay(time.Second*3), redsync.WithTries(10), redsync.WithExpiry(time.Second*30))
	err := mutex.Lock()
	if err != nil {
		log.Error().Err(err).Msg("Require redis lock failed")
		return
	}
	defer func() {
		if ok, err := mutex.Unlock(); !ok || err != nil {
			log.Error().Err(err).Msg("Release redis lock failed")
		}
	}()
	ctx = log.Logger.WithContext(ctx)
	builderService := r.builderServiceFactory.New()
	builderObjs, err := builderService.GetByNextTrigger(ctx, time.Now(), maxJob)
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
				ScmBranch: ptr.To(builderObj.CronBranch),
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
	if len(builderObjs) >= maxJob {
		tw.TickNext(tickNextDuration)
	}
}
