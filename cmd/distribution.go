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

	"github.com/go-sigma/sigma/pkg/cmds/distribution"
	"github.com/go-sigma/sigma/pkg/dal"
	"github.com/go-sigma/sigma/pkg/inits"
	"github.com/go-sigma/sigma/pkg/logger"
)

// distributionCmd represents the distribution command
var distributionCmd = &cobra.Command{
	Use:     "distribution",
	Aliases: []string{"ds"},
	Short:   "Start the sigma distribution server",
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

		err = distribution.Serve(digCon)
		if err != nil {
			log.Error().Err(err).Msg("Start distribution with error")
			return
		}
	},
}

func init() {
	distributionCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/sigma/sigma.yaml)")
	rootCmd.AddCommand(distributionCmd)
}
