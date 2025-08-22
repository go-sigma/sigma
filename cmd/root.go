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
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"github.com/go-sigma/sigma/cmd/builder"
	"github.com/go-sigma/sigma/cmd/distribution"
	"github.com/go-sigma/sigma/cmd/server"
	"github.com/go-sigma/sigma/cmd/tools"
	"github.com/go-sigma/sigma/cmd/version"
	"github.com/go-sigma/sigma/cmd/worker"
	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/logger"
	"github.com/go-sigma/sigma/pkg/utils"
)

var cfgFile string
var logLevel string

// rootCmd represents the base command when called without any subcommands
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sigma",
		Short: "sigma is an OCI artifact storage and distribution system",
		Long: `sigma is an OCI artifact storage and distribution system,
which is designed to be a lightweight, easy-to-use, and easy-to-deploy,
and can be used as a private registry or a public registry.
sigma is a cloud-native, distributed, and highly available system,
which can be deployed on any cloud platform or on-premises.`,
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			initConfig()
			logger.SetLevel(viper.GetString("log.level"))
		},
	}
	cmd.PersistentFlags().StringVarP(&cfgFile, "config", "c",
		"/etc/sigma/config.yaml", "config file (default is /etc/sigma/config.yaml)")
	cmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "info", "log level")
	cmd.AddCommand(server.NewCmdServer())
	cmd.AddCommand(worker.NewCmdWorker())
	cmd.AddCommand(version.NewCmdVersion())
	cmd.AddCommand(tools.NewCmdTools())
	cmd.AddCommand(distribution.NewCmdDistribution())
	cmd.AddCommand(builder.NewCmdBuilder())
	return cmd
}

// Execute ...
func Execute() {
	err := NewRootCmd().Execute()
	if err != nil {
		log.Error().Err(err).Msg("Execute root command with error")
		os.Exit(1)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if strings.TrimSpace(cfgFile) == "" {
		log.Fatal().Msg("config file not found")
	}
	if !utils.IsExist(cfgFile) {
		log.Fatal().Str("config", cfgFile).Msg("config file not exist")
	}
	data, err := os.ReadFile(cfgFile)
	if err != nil {
		log.Fatal().Err(err).Msg("read config file failed")
	}
	err = yaml.Unmarshal(data, configs.GetConfig())
	if err != nil {
		log.Fatal().Err(err).Msg("unmarshal failed")
	}
}
