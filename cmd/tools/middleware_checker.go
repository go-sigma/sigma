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

package tools

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/spf13/cobra"

	"github.com/go-sigma/sigma/pkg/config"
)

func toolsMiddlewareCheckerCmd() *cobra.Command {
	var waitTimeout time.Duration
	cmd := &cobra.Command{
		Use:   "middleware-checker",
		Short: "Check all of middleware status all ready",
		RunE: func(_ *cobra.Command, _ []string) error {
			if waitTimeout == 0 {
				waitTimeout = time.Second * 120
			}

			ctx, cancel := context.WithTimeout(context.Background(), waitTimeout)
			defer cancel()

			for {
				select {
				case <-ctx.Done():
					return fmt.Errorf("middleware checker timeout, not all of middleware ready")
				case <-time.After(time.Second * 3):
					err := config.CheckMiddleware()
					if err != nil {
						slog.Error("check middleware failed", "err", err)
					} else {
						return nil
					}
				}
			}
		},
	}

	cmd.PersistentFlags().DurationVar(&waitTimeout, "wait-timeout", time.Second*120, "wait middleware timeout")

	return cmd
}
