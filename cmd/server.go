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
	"github.com/spf13/viper"

	"github.com/ximager/ximager/pkg/cmds/server"
	"github.com/ximager/ximager/pkg/configs"
	"github.com/ximager/ximager/pkg/daemon"
	"github.com/ximager/ximager/pkg/dal"
	"github.com/ximager/ximager/pkg/inits"
	"github.com/ximager/ximager/pkg/logger"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the XImager server",
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		initConfig()
		logger.SetLevel(viper.GetString("log.level"))
	},
	Run: func(_ *cobra.Command, _ []string) {
		err := configs.Initialize()
		if err != nil {
			log.Error().Err(err).Msg("Initialize configs with error")
			return
		}

		err = dal.Initialize()
		if err != nil {
			log.Error().Err(err).Msg("Initialize database with error")
			return
		}

		err = inits.Initialize()
		if err != nil {
			log.Error().Err(err).Msg("Initialize inits with error")
			return
		}

		err = daemon.InitializeClient()
		if err != nil {
			log.Error().Err(err).Msg("Initialize daemon client with error")
			return
		}

		err = server.Serve(server.ServerConfig{
			WithoutDistribution: withoutDistribution,
			WithoutWorker:       withoutWorker,
		})
		if err != nil {
			log.Error().Err(err).Msg("Serve with error")
			return
		}
	},
}

var withoutDistribution bool
var withoutWorker bool

func init() {
	serverCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/ximager/ximager.yaml)")
	serverCmd.PersistentFlags().BoolVar(&withoutDistribution, "without-distribution", false, "server without distribution service")
	serverCmd.PersistentFlags().BoolVar(&withoutWorker, "without-worker", false, "server without worker service")
	rootCmd.AddCommand(serverCmd)
}
