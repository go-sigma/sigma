package scan

import (
	"context"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"

	"github.com/ximager/ximager/pkg/consts"
	"github.com/ximager/ximager/pkg/daemon"
)

func init() {
	err := daemon.RegisterTask(consts.TopicScan, runner)
	if err != nil {
		log.Fatal().Err(err).Msg("RegisterTask error")
	}
}

func runner(context.Context, *asynq.Task) error {
	return nil
}
