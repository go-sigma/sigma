// Copyright 2023 sigma
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
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/go-sigma/sigma/pkg/cmds/worker"
	"github.com/go-sigma/sigma/pkg/dal"
	"github.com/go-sigma/sigma/pkg/inits"
)

// workerCmd represents the worker command
func NewCmdWorker() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "worker",
		Short: "Start the sigma worker",
		Run: func(_ *cobra.Command, _ []string) {
			digCon, err := inits.NewDigContainer()
			if err != nil {
				log.Error().Err(err).Msg("new dig container failed")
				return
			}

			err = dal.Initialize(digCon)
			if err != nil {
				log.Error().Err(err).Msg("Initialize database with error")
				return
			}

			err = inits.Initialize(digCon)
			if err != nil {
				log.Error().Err(err).Msg("Initialize inits with error")
				return
			}

			err = worker.Worker(digCon)
			if err != nil {
				log.Error().Err(err).Msg("Start worker with error")
				return
			}
		},
	}
	return cmd
}
