// Copyright 2025 sigma
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

package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/spf13/cobra"

	"github.com/go-sigma/sigma/pkg/app/bootstrap"
	"github.com/go-sigma/sigma/pkg/app/worker"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/dal"
	"github.com/go-sigma/sigma/pkg/telemetry"
)

// workerCmd represents the worker command
func NewCmdWorker() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "worker",
		Short: "start the sigma worker",
		Run: func(_ *cobra.Command, _ []string) {
			digCon, err := bootstrap.NewDigContainer()
			if err != nil {
				slog.Error("new dig container failed", "err", err)
				return
			}

			shutdown, err := telemetry.Init(config.GetConfig())
			if err != nil {
				slog.Error("init trace failed", "err", err)
				return
			}
			defer func() {
				if shutdown != nil {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					_ = shutdown(ctx)
				}
			}()

			err = dal.Initialize(digCon)
			if err != nil {
				slog.Error("initialize database failed", "err", err)
				return
			}

			err = bootstrap.Initialize(digCon)
			if err != nil {
				slog.Error("initialize bootstrap failed", "err", err)
				return
			}

			err = worker.Worker(digCon)
			if err != nil {
				slog.Error("start worker failed", "err", err)
				return
			}
		},
	}
	return cmd
}
