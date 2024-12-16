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

package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/go-sigma/sigma/pkg/cmds/server"
	"github.com/go-sigma/sigma/pkg/dal"
	"github.com/go-sigma/sigma/pkg/inits"
	"github.com/go-sigma/sigma/pkg/logger"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the sigma server",
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		initConfig()
		logger.SetLevel(viper.GetString("log.level"))
	},
	Run: func(_ *cobra.Command, _ []string) {
		// err := configs.Initialize()
		// if err != nil {
		// 	log.Error().Err(err).Msg("Initialize configs with error")
		// 	return
		// }

		digCon, err := inits.NewDigContainer()
		if err != nil {
			log.Error().Err(err).Msg("new dig container failed")
			return
		}

		err = digCon.Provide(func() server.ServerConfig {
			return server.ServerConfig{
				WithoutDistribution: withoutDistribution,
				WithoutWorker:       withoutWorker,
				WithoutWeb:          withoutWeb,
			}
		})
		if err != nil {
			log.Error().Err(err).Msg("dig container provide server config with error")
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

		err = server.Serve(digCon)
		if err != nil {
			log.Error().Err(err).Msg("Serve with error")
			return
		}
	},
}

var withoutDistribution bool
var withoutWorker bool
var withoutWeb bool

func init() {
	serverCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is /etc/sigma/config.yaml)")
	serverCmd.PersistentFlags().BoolVar(&withoutDistribution, "without-distribution", false, "server without distribution service")
	serverCmd.PersistentFlags().BoolVar(&withoutWorker, "without-worker", false, "server without worker service")
	serverCmd.PersistentFlags().BoolVar(&withoutWeb, "without-web", false, "server without web service")
	rootCmd.AddCommand(serverCmd)
}
