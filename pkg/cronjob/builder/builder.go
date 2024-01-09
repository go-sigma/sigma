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
	"bytes"
	"context"
	"fmt"
	"html/template"
	"time"

	"github.com/Masterminds/sprig/v3"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/cronjob"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/modules/locker"
	"github.com/go-sigma/sigma/pkg/modules/timewheel"
	"github.com/go-sigma/sigma/pkg/modules/workq"
	"github.com/go-sigma/sigma/pkg/modules/workq/definition"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
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
		config:                ptr.To(configs.GetConfiguration()),
		builderServiceFactory: dao.NewBuilderServiceFactory(),
	}
	builderTw.AddRunner(runner.runner)
}

type builderRunner struct {
	config                configs.Configuration
	builderServiceFactory dao.BuilderServiceFactory
}

func (r builderRunner) runner(ctx context.Context, tw timewheel.TimeWheel) {
	locker, err := locker.New(r.config)
	if err != nil {
		log.Error().Err(err).Msg("New locker failed")
		return
	}
	lock, err := locker.Lock(context.Background(), consts.LockerCronjobBuilder, time.Second*30)
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
			tag, err := buildTag(ptr.To(builderObj.CronTagTemplate), buildTagOption{ScmBranch: ptr.To(builderObj.CronBranch)})
			if err != nil {
				return err
			}
			runner, err := buildRunner(builderObj, buildRunnerOption{
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

			err = workq.ProducerClient.Produce(ctx, string(enums.DaemonBuilder), types.DaemonBuilderPayload{
				Action:    enums.DaemonBuilderActionStart,
				BuilderID: builderObj.ID,
				RunnerID:  runner.ID,
			}, definition.ProducerOption{Tx: tx})
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

// buildRunnerOption ...
type buildRunnerOption struct {
	Tag       string
	ScmBranch *string
}

// buildRunner ...
// nolint: unparam
func buildRunner(builder *models.Builder, option buildRunnerOption) (*models.BuilderRunner, error) {
	runner := &models.BuilderRunner{
		BuilderID: builder.ID,
		Status:    enums.BuildStatusPending,

		// Tag:       option.Tag,
		ScmBranch: option.ScmBranch,
	}
	return runner, nil
}

// buildTagOption ...
type buildTagOption struct {
	ScmBranch string
	ScmTag    string
	ScmRef    string
}

// buildTag ...
func buildTag(tmpl string, option buildTagOption) (string, error) {
	t, err := template.New("tag").Funcs(sprig.FuncMap()).Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("Template parse failed: %v", err)
	}
	var buffer bytes.Buffer
	err = t.Execute(&buffer, option)
	if err != nil {
		return "", fmt.Errorf("Execute template failed: %v", err)
	}
	return buffer.String(), nil
}
