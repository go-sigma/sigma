// Copyright 2026 sigma
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
	"errors"
	"testing"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/background/cronjob"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	repobuilder "github.com/go-sigma/sigma/pkg/dal/repository/builder"
	"github.com/go-sigma/sigma/pkg/infra/lock"
	"github.com/go-sigma/sigma/pkg/infra/timewheel"
	"github.com/go-sigma/sigma/pkg/testkit"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

func TestBuilderLockKey(t *testing.T) {
	require.Equal(t, consts.LockerCronjobBuilder+":builder-id", builderLockKey("builder-id"))
}

func TestBuilderRunnerContinuesAfterLockFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	builderRepository := repobuilder.NewMockBuilderRepository(ctrl)
	builders := []*models.Builder{{ID: "first"}, {ID: "second"}}
	builderRepository.EXPECT().
		GetByNextTrigger(gomock.Any(), gomock.Any(), cronjob.MaxJob).
		Return(builders, nil)
	locker := &recordingLocker{
		errs: map[string]error{
			builderLockKey("first"):  context.DeadlineExceeded,
			builderLockKey("second"): errors.New("redis unavailable"),
		},
	}

	builderRunner{
		builderRepository: builderRepository,
		locker:            locker,
	}.runner(t.Context(), fakeTimeWheel{})

	require.Equal(t, []string{builderLockKey("first"), builderLockKey("second")}, locker.keys)
}

func TestRunBuilderClaimsScheduleOnce(t *testing.T) {
	require.NotNil(t, testkit.InitRepository(t))

	scanTime := time.Now().Truncate(time.Second)
	cronRule := "* * * * *"
	tagTemplate := "latest"
	branch := "main"
	builderObj := &models.Builder{
		ID:              uuid.NewV7String(),
		RepositoryID:    uuid.NewV7String(),
		Source:          enums.BuilderSourceDockerfile,
		CronRule:        &cronRule,
		CronBranch:      &branch,
		CronTagTemplate: &tagTemplate,
		CronNextTrigger: new(scanTime.Add(-time.Minute).UnixMilli()),
	}
	require.NoError(t, query.Q.Builder.WithContext(t.Context()).Create(builderObj))
	producer := &recordingProducer{}
	runner := builderRunner{producer: producer}
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

	require.NoError(t, runner.runBuilder(t.Context(), scanTime, builderObj, parser))
	require.NoError(t, runner.runBuilder(t.Context(), scanTime, builderObj, parser))

	runners, err := query.Q.BuilderRunner.WithContext(t.Context()).
		Where(query.Q.BuilderRunner.BuilderID.Eq(builderObj.ID)).
		Find()
	require.NoError(t, err)
	require.Len(t, runners, 1)
	require.Equal(t, "latest", runners[0].RawTag)
	require.Len(t, producer.payloads, 1)
}

type recordingLocker struct {
	keys []string
	errs map[string]error
}

func (l *recordingLocker) Acquire(_ context.Context, _ string, _, _ time.Duration) (lock.Lock, error) {
	return nil, errors.New("not implemented")
}

func (l *recordingLocker) AcquireWithRenew(_ context.Context, key string, _, _ time.Duration) error {
	l.keys = append(l.keys, key)
	return l.errs[key]
}

type recordingProducer struct {
	payloads []any
}

func (p *recordingProducer) Produce(_ context.Context, _ enums.Daemon, payload any) error {
	p.payloads = append(p.payloads, payload)
	return nil
}

type fakeTimeWheel struct{}

func (fakeTimeWheel) TickNext(time.Duration)     {}
func (fakeTimeWheel) AddRunner(timewheel.Notify) {}
func (fakeTimeWheel) Stop()                      {}
