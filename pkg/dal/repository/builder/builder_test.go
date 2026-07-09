// Copyright 2024 sigma
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

package builder_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	repobuilder "github.com/go-sigma/sigma/pkg/dal/repository/builder"
	"github.com/go-sigma/sigma/pkg/testkit"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

func TestNewBuilderRepository(t *testing.T) {
	require.NotNil(t, repobuilder.NewBuilderRepository())
	require.NotNil(t, repobuilder.NewBuilderRepository(query.Q))
}

func TestBuilderRepositoryClaimNextTrigger(t *testing.T) {
	require.NotNil(t, testkit.InitRepository(t))

	ctx := t.Context()
	now := time.Now().Truncate(time.Second)
	dueBefore := now.Add(time.Minute)
	next := now.Add(time.Hour)
	builderObj := &models.Builder{
		ID:              uuid.NewV7String(),
		RepositoryID:    uuid.NewV7String(),
		Source:          enums.BuilderSourceDockerfile,
		CronNextTrigger: new(now.UnixMilli()),
	}
	require.NoError(t, query.Q.Builder.WithContext(ctx).Create(builderObj))

	builderRepository := repobuilder.NewBuilderRepository()
	claimed, err := builderRepository.ClaimNextTrigger(ctx, builderObj.ID, dueBefore, next)
	require.NoError(t, err)
	require.True(t, claimed)

	claimed, err = builderRepository.ClaimNextTrigger(ctx, builderObj.ID, dueBefore, next.Add(time.Hour))
	require.NoError(t, err)
	require.False(t, claimed)

	futureBuilder := &models.Builder{
		ID:              uuid.NewV7String(),
		RepositoryID:    uuid.NewV7String(),
		Source:          enums.BuilderSourceDockerfile,
		CronNextTrigger: new(next.UnixMilli()),
	}
	require.NoError(t, query.Q.Builder.WithContext(ctx).Create(futureBuilder))
	claimed, err = builderRepository.ClaimNextTrigger(ctx, futureBuilder.ID, dueBefore, next.Add(time.Hour))
	require.NoError(t, err)
	require.False(t, claimed)
}

func TestBuilderRepositoryLifecycle(t *testing.T) {
	require.NotNil(t, testkit.InitRepository(t))

	ctx := t.Context()
	builderRepository := repobuilder.NewBuilderRepository()
	branch := "main"
	builderObj := &models.Builder{
		ID:           uuid.NewV7String(),
		RepositoryID: uuid.NewV7String(),
		Source:       enums.BuilderSourceDockerfile,
		Dockerfile:   []byte("FROM scratch"),
		ScmBranch:    &branch,
	}

	require.NoError(t, builderRepository.Create(ctx, builderObj))

	got, err := builderRepository.Get(ctx, builderObj.RepositoryID)
	require.NoError(t, err)
	require.Equal(t, builderObj.ID, got.ID)
	require.Equal(t, []byte("FROM scratch"), got.Dockerfile)

	byIDs, err := builderRepository.GetByRepositoryIDs(ctx, []string{builderObj.RepositoryID, uuid.NewV7String()})
	require.NoError(t, err)
	require.Len(t, byIDs, 1)
	require.Equal(t, builderObj.ID, byIDs[builderObj.RepositoryID].ID)

	empty, err := builderRepository.GetByRepositoryIDs(ctx, nil)
	require.NoError(t, err)
	require.Nil(t, empty)

	require.NoError(t, builderRepository.Update(ctx, builderObj.ID, map[string]any{
		query.Builder.ScmBranch.ColumnName().String(): "release",
	}))
	require.NoError(t, builderRepository.Update(ctx, builderObj.ID, map[string]any{}))

	got, err = builderRepository.Get(ctx, builderObj.RepositoryID)
	require.NoError(t, err)
	require.Equal(t, "release", ptr.To(got.ScmBranch))

	err = builderRepository.Update(ctx, uuid.NewV7String(), map[string]any{
		query.Builder.ScmBranch.ColumnName().String(): "missing",
	})
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestBuilderRunnerLifecycle(t *testing.T) {
	require.NotNil(t, testkit.InitRepository(t))

	ctx := t.Context()
	builderRepository := repobuilder.NewBuilderRepository()
	builderObj := &models.Builder{
		ID:           uuid.NewV7String(),
		RepositoryID: uuid.NewV7String(),
		Source:       enums.BuilderSourceDockerfile,
	}
	require.NoError(t, builderRepository.Create(ctx, builderObj))

	runnerOne := &models.BuilderRunner{
		ID:        uuid.NewV7String(),
		BuilderID: builderObj.ID,
		Status:    enums.BuildStatusPending,
		RawTag:    "v1",
	}
	runnerTwo := &models.BuilderRunner{
		ID:        uuid.NewV7String(),
		BuilderID: builderObj.ID,
		Status:    enums.BuildStatusBuilding,
		RawTag:    "v2",
	}
	require.NoError(t, builderRepository.CreateRunner(ctx, runnerOne))
	require.NoError(t, builderRepository.CreateRunner(ctx, runnerTwo))

	got, err := builderRepository.GetRunner(ctx, runnerOne.ID)
	require.NoError(t, err)
	require.Equal(t, builderObj.ID, got.Builder.ID)

	method := enums.SortMethodAsc
	page := 1
	limit := 10
	sortBy := "raw_tag"
	runners, total, err := builderRepository.ListRunners(ctx, builderObj.ID, api.Pagination{
		Page:  &page,
		Limit: &limit,
	}, api.Sortable{Sort: &sortBy, Method: &method})
	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, runners, 2)
	require.Equal(t, []string{"v1", "v2"}, []string{runners[0].RawTag, runners[1].RawTag})

	require.NoError(t, builderRepository.UpdateRunner(ctx, builderObj.ID, runnerOne.ID, map[string]any{
		query.BuilderRunner.Status.ColumnName().String(): enums.BuildStatusSuccess,
	}))
	require.NoError(t, builderRepository.UpdateRunner(ctx, builderObj.ID, runnerOne.ID, map[string]any{}))

	got, err = builderRepository.GetRunner(ctx, runnerOne.ID)
	require.NoError(t, err)
	require.Equal(t, enums.BuildStatusSuccess, got.Status)

	err = builderRepository.UpdateRunner(ctx, builderObj.ID, uuid.NewV7String(), map[string]any{
		query.BuilderRunner.Status.ColumnName().String(): enums.BuildStatusFailed,
	})
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestBuilderRepositoryNextTrigger(t *testing.T) {
	require.NotNil(t, testkit.InitRepository(t))

	ctx := t.Context()
	builderRepository := repobuilder.NewBuilderRepository()
	now := time.Now().Truncate(time.Second)
	dueAt := now.Add(-time.Minute)
	futureAt := now.Add(time.Minute)
	nextAt := now.Add(time.Hour)
	dueMillis := dueAt.UnixMilli()
	futureMillis := futureAt.UnixMilli()

	dueBuilder := &models.Builder{
		ID:              uuid.NewV7String(),
		RepositoryID:    uuid.NewV7String(),
		Source:          enums.BuilderSourceDockerfile,
		CronNextTrigger: &dueMillis,
	}
	futureBuilder := &models.Builder{
		ID:              uuid.NewV7String(),
		RepositoryID:    uuid.NewV7String(),
		Source:          enums.BuilderSourceDockerfile,
		CronNextTrigger: &futureMillis,
	}
	require.NoError(t, builderRepository.Create(ctx, dueBuilder))
	require.NoError(t, builderRepository.Create(ctx, futureBuilder))

	builders, err := builderRepository.GetByNextTrigger(ctx, now, 10)
	require.NoError(t, err)
	require.Len(t, builders, 1)
	require.Equal(t, dueBuilder.ID, builders[0].ID)

	require.NoError(t, builderRepository.UpdateNextTrigger(ctx, dueBuilder.ID, nextAt))
	got, err := query.Q.Builder.WithContext(ctx).Where(query.Q.Builder.ID.Eq(dueBuilder.ID)).First()
	require.NoError(t, err)
	require.Equal(t, nextAt.UnixMilli(), ptr.To(got.CronNextTrigger))

	err = builderRepository.UpdateNextTrigger(ctx, uuid.NewV7String(), nextAt)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}
