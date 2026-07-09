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

package size

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/background/cronjob"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/consts"
	reponamespace "github.com/go-sigma/sigma/pkg/dal/repository/namespace"
	reporegistry "github.com/go-sigma/sigma/pkg/dal/repository/registry"
	"github.com/go-sigma/sigma/pkg/infra/lock"
	"github.com/go-sigma/sigma/pkg/infra/timewheel"
)

var sizeTw timewheel.TimeWheel

func init() {
	cronjob.Starter = append(cronjob.Starter, sizeReconcileJob)
	cronjob.Stopper = append(cronjob.Stopper, func() {
		if sizeTw != nil {
			sizeTw.Stop()
		}
	})
}

type reconcileParams struct {
	dig.In

	Config               *config.Configuration
	NamespaceRepository  reponamespace.NamespaceRepository
	RepositoryRepository reporegistry.RepositoryRepository
	ArtifactRepository   reporegistry.ArtifactRepository
	Locker               lock.Locker
}

func sizeReconcileJob(digCon *dig.Container) error {
	var params reconcileParams
	if err := digCon.Invoke(func(p reconcileParams) { params = p }); err != nil {
		return err
	}

	cfg := params.Config.Daemon.SizeReconcile
	sizeTw = timewheel.NewTimeWheel(context.Background(), cfg.Interval)
	sizeTw.AddRunner(reconcileRunner{
		namespaceRepository:  params.NamespaceRepository,
		repositoryRepository: params.RepositoryRepository,
		artifactRepository:   params.ArtifactRepository,
		locker:               params.Locker,
		interval:             cfg.Interval,
		batchSize:            cfg.BatchSize,
		lockExpire:           cfg.LockExpire,
		lockWaitTimeout:      cfg.LockWaitTimeout,
	}.runner)
	return nil
}

type reconcileRunner struct {
	namespaceRepository  reponamespace.NamespaceRepository
	repositoryRepository reporegistry.RepositoryRepository
	artifactRepository   reporegistry.ArtifactRepository
	locker               lock.Locker
	interval             time.Duration
	batchSize            int
	lockExpire           time.Duration
	lockWaitTimeout      time.Duration
}

func (r reconcileRunner) runner(ctx context.Context, _ timewheel.TimeWheel) {
	lockCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	if err := r.locker.AcquireWithRenew(lockCtx, consts.LockerCronjobSize, r.lockExpire, r.lockWaitTimeout); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			slog.Info("skip size reconcile because another runner holds the lock")
			return
		}
		slog.Error("acquire size reconcile lock failed", "err", err)
		return
	}

	r.reconcileNamespaces(lockCtx)
	r.reconcileRepositories(lockCtx)
}

func (r reconcileRunner) reconcileNamespaces(ctx context.Context) {
	var lastID string
	var reconciled int
	for {
		namespaces, err := r.namespaceRepository.FindWithCursorDirty(ctx, int64(r.batchSize), lastID)
		if err != nil {
			slog.Error("size reconcile: list dirty namespaces failed", "err", err)
			return
		}
		for _, ns := range namespaces {
			actualSize, err := r.artifactRepository.GetNamespaceSize(ctx, ns.ID)
			if err != nil {
				slog.Error("size reconcile: get namespace size failed", "err", err, "namespace_id", ns.ID)
				continue
			}
			if ns.Size != actualSize {
				slog.Info("size reconcile: namespace size drift detected", "namespace_id", ns.ID, "recorded", ns.Size, "actual", actualSize)
				err = r.namespaceRepository.UpdateSizeAndClearDirty(ctx, ns.ID, actualSize)
				if err != nil {
					slog.Error("size reconcile: update namespace size failed", "err", err, "namespace_id", ns.ID)
					continue
				}
				reconciled++
			} else {
				// Size is consistent, just clear the dirty flag
				err = r.namespaceRepository.ClearSizeDirty(ctx, ns.ID)
				if err != nil {
					slog.Error("size reconcile: clear namespace dirty failed", "err", err, "namespace_id", ns.ID)
					continue
				}
			}
		}
		if len(namespaces) < r.batchSize {
			slog.Info("size reconcile: namespaces finished", "reconciled", reconciled)
			return
		}
		lastID = namespaces[len(namespaces)-1].ID
	}
}

func (r reconcileRunner) reconcileRepositories(ctx context.Context) {
	var lastID string
	var reconciled int
	for {
		repositories, err := r.repositoryRepository.FindWithCursorDirty(ctx, int64(r.batchSize), lastID)
		if err != nil {
			slog.Error("size reconcile: list dirty repositories failed", "err", err)
			return
		}
		for _, repo := range repositories {
			actualSize, err := r.artifactRepository.GetRepositorySize(ctx, repo.ID)
			if err != nil {
				slog.Error("size reconcile: get repository size failed", "err", err, "repository_id", repo.ID)
				continue
			}
			if repo.Size != actualSize {
				slog.Info("size reconcile: repository size drift detected", "repository_id", repo.ID, "recorded", repo.Size, "actual", actualSize)
				err = r.repositoryRepository.UpdateSizeAndClearDirty(ctx, repo.ID, actualSize)
				if err != nil {
					slog.Error("size reconcile: update repository size failed", "err", err, "repository_id", repo.ID)
					continue
				}
				reconciled++
			} else {
				// Size is consistent, just clear the dirty flag
				err = r.repositoryRepository.ClearSizeDirty(ctx, repo.ID)
				if err != nil {
					slog.Error("size reconcile: clear repository dirty failed", "err", err, "repository_id", repo.ID)
					continue
				}
			}
		}
		if len(repositories) < r.batchSize {
			slog.Info("size reconcile: repositories finished", "reconciled", reconciled)
			return
		}
		lastID = repositories[len(repositories)-1].ID
	}
}
