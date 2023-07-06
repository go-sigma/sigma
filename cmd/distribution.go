package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ximager/ximager/pkg/cmds/distribution"
	"github.com/ximager/ximager/pkg/configs"
	"github.com/ximager/ximager/pkg/daemon"
	"github.com/ximager/ximager/pkg/dal"
	"github.com/ximager/ximager/pkg/inits"
	"github.com/ximager/ximager/pkg/logger"
)

// distributionCmd represents the distribution command
var distributionCmd = &cobra.Command{
	Use:     "distribution",
	Aliases: []string{"ds"},
	Short:   "Start the XImager distribution server",
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

		err = distribution.Serve()
		if err != nil {
			log.Error().Err(err).Msg("Start distribution with error")
			return
		}
	},
}

func init() {
	distributionCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/ximager/ximager.yaml)")
	rootCmd.AddCommand(distributionCmd)
}
