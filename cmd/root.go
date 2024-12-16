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

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/consts"

	_ "github.com/go-sigma/sigma/cmd/imports"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "sigma",
	Short: "sigma is an OCI artifact storage and distribution system",
	Long: `sigma is an OCI artifact storage and distribution system,
which is designed to be a lightweight, easy-to-use, and easy-to-deploy,
and can be used as a private registry or a public registry.
sigma is a cloud-native, distributed, and highly available system,
which can be deployed on any cloud platform or on-premises.`,
}

// Execute ...
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Error().Err(err).Msg("Execute root command with error")
		os.Exit(1)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath("/etc/sigma")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config.yaml")
	}

	viper.AutomaticEnv()
	viper.SetEnvPrefix(consts.AppName)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	cobra.CheckErr(viper.ReadInConfig())
	cobra.CheckErr(viper.Unmarshal(configs.GetConfiguration()))
}
