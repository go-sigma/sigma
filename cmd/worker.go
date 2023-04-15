package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/ximager/ximager/pkg/cmds/worker"
	"github.com/ximager/ximager/pkg/dal"
)

// workerCmd represents the worker command
var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Start the XImager worker",
	Run: func(_ *cobra.Command, _ []string) {
		err := dal.Initialize()
		if err != nil {
			log.Error().Err(err).Msg("Initialize database with error")
			return
		}

		err = worker.Worker()
		if err != nil {
			log.Error().Err(err).Msg("Start worker with error")
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(workerCmd)
}
