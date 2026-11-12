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

package distribution

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/go-sigma/sigma/pkg/app/bootstrap"
	"github.com/go-sigma/sigma/pkg/app/distribution"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/dal"
	"github.com/go-sigma/sigma/pkg/logger"
	"github.com/go-sigma/sigma/pkg/telemetry"
)

// distributionCmd represents the distribution command
func NewCmdDistribution() *cobra.Command {
	return &cobra.Command{
		Use:     "distribution",
		Aliases: []string{"ds"},
		Short:   "Start the sigma distribution server",
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

			err = dal.Initialize(digCon)
			if err != nil {
				logger.Fatal("initialize database with error", "err", err)
			}

			err = bootstrap.Initialize(digCon)
			if err != nil {
				logger.Fatal("initialize bootstrap with error", "err", err)
			}

			err = distribution.Serve(digCon)
			if err != nil {
				logger.Fatal("start distribution with error", "err", err)
			}
		},
	}
}
