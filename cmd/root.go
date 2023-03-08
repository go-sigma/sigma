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
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ximager/ximager/pkg/configs"
	"github.com/ximager/ximager/pkg/utils"

	_ "github.com/ximager/ximager/pkg/storage/filesystem"
	_ "github.com/ximager/ximager/pkg/storage/s3"
	_ "github.com/ximager/ximager/pkg/utils/leader/k8s"
	_ "github.com/ximager/ximager/pkg/utils/leader/redis"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ximager",
	Short: "XImager is an OCI artifact storage and distribution system",
	Long: `XImager is an OCI artifact storage and distribution system,
which is designed to be a lightweight, easy-to-use, and easy-to-deploy,
and can be used as a private registry or a public registry.
XImager is a cloud-native, distributed, and highly available system,
which can be deployed on any cloud platform or on-premises.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		err := configs.Initialize()
		if err != nil {
			return err
		}
		return nil
	},
}

// Execute ...
func Execute() {
	rootCmd.PersistentPreRun = func(_ *cobra.Command, _ []string) {
		utils.SetLevel(viper.GetInt("log.level"))
	}
	err := rootCmd.Execute()
	if err != nil {
		log.Error().Err(err).Msg("Execute root command with error")
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ximager.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".ximager" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".ximager")
	}

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Sprintf("Fatal error config file: %s \n", err))
	}
}
