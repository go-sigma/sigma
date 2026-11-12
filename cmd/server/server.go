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

package server

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/go-sigma/sigma/pkg/app/bootstrap"
	"github.com/go-sigma/sigma/pkg/app/server"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/dal"
	"github.com/go-sigma/sigma/pkg/logger"
	"github.com/go-sigma/sigma/pkg/telemetry"
)

// NewCmdServer represents the server command
func NewCmdServer() *cobra.Command {
	var withoutDistribution bool
	var withoutWorker bool
	var withoutWeb bool
	cmd := &cobra.Command{
		Use:   "server",
		Short: "start the sigma server",
		// Startup failures must exit non-zero: Kubernetes reports an exit code
		// of 0 as "Completed" and hides the reason a container keeps
		// restarting.
		Run: func(_ *cobra.Command, _ []string) {
			digCon, err := bootstrap.NewDigContainer()
			if err != nil {
				logger.Fatal("new dig container failed", "err", err)
			}

			shutdown, err := telemetry.Init(config.GetConfig())
			if err != nil {
				logger.Fatal("init trace failed", "err", err)
			}
			defer func() {
				if shutdown != nil {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					_ = shutdown(ctx)
				}
			}()

			err = digCon.Provide(func() server.ServerConfig {
				return server.ServerConfig{
					WithoutDistribution: withoutDistribution,
					WithoutWorker:       withoutWorker,
					WithoutWeb:          withoutWeb,
				}
			})
			if err != nil {
				logger.Fatal("dig container provide server config failed", "err", err)
			}

			err = dal.Initialize(digCon)
			if err != nil {
				logger.Fatal("initialize database failed", "err", err)
			}

			err = bootstrap.Initialize(digCon)
			if err != nil {
				logger.Fatal("initialize bootstrap failed", "err", err)
			}

			err = server.Serve(digCon)
			if err != nil {
				logger.Fatal("start server failed", "err", err)
			}
		},
	}
	cmd.PersistentFlags().BoolVar(&withoutDistribution, "without-distribution", false, "server without distribution service")
	cmd.PersistentFlags().BoolVar(&withoutWorker, "without-worker", false, "server without worker service")
	cmd.PersistentFlags().BoolVar(&withoutWeb, "without-web", false, "server without web service")
	return cmd
}
