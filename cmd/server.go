// Copyright 2023 XImager
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

package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/ximager/ximager/pkg/cmds/server"
	"github.com/ximager/ximager/pkg/daemon"
	"github.com/ximager/ximager/pkg/dal"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the XImager server",
	Run: func(_ *cobra.Command, _ []string) {
		err := dal.Initialize()
		if err != nil {
			log.Error().Err(err).Msg("Initialize database with error")
			return
		}

		err = daemon.InitializeClient()
		if err != nil {
			log.Error().Err(err).Msg("Initialize daemon client with error")
			return
		}

		err = server.Serve()
		if err != nil {
			log.Error().Err(err).Msg("Serve with error")
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
