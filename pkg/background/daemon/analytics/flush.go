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

package analytics

import (
	"context"
	"log/slog"
	"path"
	"reflect"
	"time"

	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/background/daemon"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/graceful"
	svcanalytics "github.com/go-sigma/sigma/pkg/service/analytics"
	"github.com/go-sigma/sigma/pkg/utils"
)

const (
	counterBackendRedis    = "redis"
	counterBackendInmemory = "inmemory"
)

func init() {
	utils.PanicIf(daemon.Daemons.Register(path.Base(reflect.TypeFor[factory]().PkgPath()), &factory{}))
}

type factory struct{}

type params struct {
	dig.In

	Config       *config.Configuration
	AnalyticsSvc svcanalytics.Service
}

func (f factory) Initialize(digCon *dig.Container) error {
	var p params
	if err := digCon.Invoke(func(i params) { p = i }); err != nil {
		return err
	}
	if !p.Config.Analytics.Enabled {
		return nil
	}

	backend := p.Config.Analytics.CounterBackend
	if backend == "" {
		backend = counterBackendInmemory
	}

	switch backend {
	case counterBackendRedis:
		interval, err := time.ParseDuration(p.Config.Analytics.FlushInterval)
		if err != nil || interval <= 0 {
			interval = time.Minute
		}
		ctx, cancel := context.WithCancel(context.Background())
		go flushLoop(ctx, p.AnalyticsSvc, interval)
		graceful.RunAtShutdown("analytics", 50, func() {
			cancel()
			if err := p.AnalyticsSvc.Flush(context.Background()); err != nil {
				slog.Error("flush analytics failed", "err", err)
			}
		})
	case counterBackendInmemory:
		graceful.RunAtShutdown("analytics", 50, func() {
			if err := p.AnalyticsSvc.Flush(context.Background()); err != nil {
				slog.Error("flush analytics failed", "err", err)
			}
		})
	}
	return nil
}

func flushLoop(ctx context.Context, service svcanalytics.Service, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := service.Flush(ctx); err != nil {
				slog.Error("flush analytics failed", "err", err)
			}
		}
	}
}
