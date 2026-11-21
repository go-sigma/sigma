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

package web

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/logger"
	"github.com/go-sigma/sigma/pkg/telemetry"
	webapp "github.com/go-sigma/sigma/pkg/web"
)

// NewCmdWeb represents the web command, which serves the built UI without the
// API or the registry.
func NewCmdWeb() *cobra.Command {
	return &cobra.Command{
		Use:   "web",
		Short: "Start the sigma web front end",
		Run: func(_ *cobra.Command, _ []string) {
			cfg := config.GetConfig()

			shutdown, err := telemetry.Init(cfg)
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

			if err := webapp.Serve(cfg); err != nil {
				logger.Fatal("start web with error", "err", err)
			}
		},
	}
}
