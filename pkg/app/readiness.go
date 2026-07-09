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

package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/gorm"

	dalredis "github.com/go-sigma/sigma/pkg/dal/redis"
	"github.com/go-sigma/sigma/pkg/server/middlewares/healthz"
	"github.com/go-sigma/sigma/pkg/storage"
)

const (
	readinessTimeout  = 2 * time.Second
	startupRetryDelay = 2 * time.Second
	// StartupTimeout is the maximum time ValidateOnStartup waits for all
	// dependencies to become reachable before failing the startup gate.
	StartupTimeout = 60 * time.Second
)

// drainReadyResource flips /readyz to 503 as soon as SetDraining(true) is
// called, so the load balancer stops routing new traffic before
// http.Server.Shutdown starts rejecting in-flight requests. /healthz
// (liveness) is unaffected and keeps returning 200.
type drainReadyResource struct{}

func (drainReadyResource) HealthCheck() error {
	if IsDraining() {
		return fmt.Errorf("shutting down")
	}
	return nil
}

type databaseReadyResource struct {
	db *gorm.DB
}

func (r databaseReadyResource) HealthCheck() error {
	if r.db == nil {
		return fmt.Errorf("database is not initialized")
	}
	rawDB, err := r.db.DB()
	if err != nil {
		return fmt.Errorf("get database instance: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), readinessTimeout)
	defer cancel()
	if err := rawDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}
	return nil
}

type redisReadyResource struct {
	redisClientFactory dalredis.ClientFactory
}

func (r redisReadyResource) HealthCheck() error {
	if r.redisClientFactory == nil {
		return fmt.Errorf("redis client factory is not initialized")
	}
	redisCli, err := r.redisClientFactory()
	if err != nil {
		return fmt.Errorf("create redis client: %w", err)
	}
	if redisCli == nil {
		return fmt.Errorf("redis client is not initialized")
	}
	ctx, cancel := context.WithTimeout(context.Background(), readinessTimeout)
	defer cancel()
	res, err := redisCli.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}
	if res != "PONG" {
		return fmt.Errorf("redis ping returned %q", res)
	}
	return nil
}

// storageReadyResource verifies the object storage backend is reachable.
type storageReadyResource struct {
	driver storage.StorageDriver
}

func (s storageReadyResource) HealthCheck() error {
	if s.driver == nil {
		return fmt.Errorf("storage driver is not initialized")
	}
	ctx, cancel := context.WithTimeout(context.Background(), readinessTimeout)
	defer cancel()
	if err := s.driver.Ping(ctx); err != nil {
		return fmt.Errorf("storage ping failed: %w", err)
	}
	return nil
}

// NewReadinessResources builds the list of healthz.Resource used by both the
// /readyz probe (via the healthz middleware) and ValidateOnStartup.
func NewReadinessResources(params GinServerParams) []healthz.Resource {
	// drainReadyResource is first: once draining, /readyz must fail fast
	// regardless of the state of downstream dependencies.
	resources := []healthz.Resource{
		drainReadyResource{},
		databaseReadyResource{db: params.Database},
		storageReadyResource{driver: params.StorageDriver},
	}
	if params.Config.Redis.Enabled() {
		resources = append(resources, redisReadyResource{
			redisClientFactory: params.RedisClientFactory,
		})
	}
	return resources
}

// ValidateOnStartup blocks until every resource passes its HealthCheck or
// ctx is done. It is invoked before the HTTP server starts listening so the
// process exits without serving traffic when a required dependency is
// unavailable. Each retry is logged via slog.
func ValidateOnStartup(ctx context.Context, resources []healthz.Resource) error {
	logger := slog.Default()
	deadline := time.Now().Add(StartupTimeout)
	if dl, ok := ctx.Deadline(); ok && dl.Before(deadline) {
		deadline = dl
	}
	for {
		var failed []string
		for _, r := range resources {
			if err := r.HealthCheck(); err != nil {
				failed = append(failed, err.Error())
			}
		}
		if len(failed) == 0 {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("startup dependency check failed: %v", failed)
		}
		logger.Warn("startup dependency check pending, retrying", "failures", failed)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(startupRetryDelay):
		}
	}
}
