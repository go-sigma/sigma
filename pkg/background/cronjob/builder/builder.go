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
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"time"

	"github.com/Masterminds/sprig/v3"
	"github.com/robfig/cron/v3"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/background/cronjob"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	repobuilder "github.com/go-sigma/sigma/pkg/dal/repository/builder"
	"github.com/go-sigma/sigma/pkg/infra/lock"
	"github.com/go-sigma/sigma/pkg/infra/timewheel"
	"github.com/go-sigma/sigma/pkg/infra/workq"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

var builderTw timewheel.TimeWheel

const (
	builderLockExpire      = 30 * time.Second
	builderLockWaitTimeout = 100 * time.Millisecond
)

func init() {
	cronjob.Starter = append(cronjob.Starter, builderJob)
	cronjob.Stopper = append(cronjob.Stopper, func() {
		if builderTw != nil {
			builderTw.Stop()
		}
	})
}

type builderParams struct {
	dig.In

	BuilderRepository repobuilder.BuilderRepository
	Producer          workq.Producer
	Locker            lock.Locker
}

func builderJob(digCon *dig.Container) error {
	var params builderParams
	if err := digCon.Invoke(func(p builderParams) { params = p }); err != nil {
		return err
	}

	builderTw = timewheel.NewTimeWheel(context.Background(), cronjob.CronjobIterDuration)

	runner := builderRunner{
		builderRepository: params.BuilderRepository,
		producer:          params.Producer,
		locker:            params.Locker,
	}
	builderTw.AddRunner(runner.runner)
	return nil
}

type builderRunner struct {
	builderRepository repobuilder.BuilderRepository
	producer          workq.Producer
	locker            lock.Locker
}

func (r builderRunner) runner(ctx context.Context, tw timewheel.TimeWheel) {
	scanTime := time.Now()
	builderRepository := r.builderRepository
	builderObjs, err := builderRepository.GetByNextTrigger(ctx, scanTime, cronjob.MaxJob)
	if err != nil {
		slog.Error("get builders by next trigger failed", "err", err)
		return
	}
	cronParser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	for _, builderObj := range builderObjs {
		lockCtx, cancel := context.WithCancel(ctx)
		err = r.locker.AcquireWithRenew(lockCtx, builderLockKey(builderObj.ID), builderLockExpire, builderLockWaitTimeout)
		if err != nil {
			cancel()
			if errors.Is(err, context.DeadlineExceeded) {
				slog.Info("skip cron builder because another runner holds the lock", "builder_id", builderObj.ID)
			} else {
				slog.Error("acquire cron builder lock failed", "builder_id", builderObj.ID, "err", err)
			}
			continue
		}

		err = r.runBuilder(lockCtx, scanTime, builderObj, cronParser)
		cancel()
		if err != nil {
			slog.Error("cronjob create builder runner failed", "builder", builderObj, "err", err)
		}
	}
	if len(builderObjs) >= cronjob.MaxJob {
		tw.TickNext(cronjob.TickNextDuration)
	}
}

func (r builderRunner) runBuilder(ctx context.Context, scanTime time.Time, builderObj *models.Builder, cronParser cron.Parser) error {
	schedule, err := cronParser.Parse(ptr.To(builderObj.CronRule))
	if err != nil {
		return err
	}
	nextTrigger := schedule.Next(scanTime)

	return query.Q.Transaction(func(tx *query.Query) error {
		builderRepository := repobuilder.NewBuilderRepository(tx)
		claimed, err := builderRepository.ClaimNextTrigger(ctx, builderObj.ID, scanTime, nextTrigger)
		if err != nil {
			return err
		}
		if !claimed {
			return nil
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
		if err = builderRepository.CreateRunner(ctx, runner); err != nil {
			return err
		}
		return r.producer.Produce(ctx, enums.DaemonBuilder, api.DaemonBuilderPayload{
			Action:       enums.DaemonBuilderActionStart,
			BuilderID:    builderObj.ID,
			RunnerID:     runner.ID,
			RepositoryID: builderObj.RepositoryID,
		})
	})
}

func builderLockKey(builderID string) string {
	return fmt.Sprintf("%s:%s", consts.LockerCronjobBuilder, builderID)
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
		ID:        uuid.NewV7String(),
		BuilderID: builder.ID,
		Status:    enums.BuildStatusPending,

		RawTag:    option.Tag,
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
		return "", fmt.Errorf("template parse failed: %v", err)
	}
	var buffer bytes.Buffer
	err = t.Execute(&buffer, option)
	if err != nil {
		return "", fmt.Errorf("execute template failed: %v", err)
	}
	return buffer.String(), nil
}
