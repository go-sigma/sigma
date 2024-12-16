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

package worker

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/rs/zerolog/log"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/builder"
	"github.com/go-sigma/sigma/pkg/cmds"
	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/graceful"
	"github.com/go-sigma/sigma/pkg/modules/workq"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

// Worker is the worker initialization
func Worker(digCon *dig.Container) error {
	config := ptr.To(configs.GetConfiguration())
	err := builder.Initialize(config)
	if err != nil {
		return err
	}

	err = workq.Initialize(config)
	if err != nil {
		return err
	}

	// e := echo.New()
	// e.HideBanner = true
	// e.HidePort = true
	// e.Use(echoprometheus.NewMiddleware(consts.AppName))
	// e.GET("/metrics", echoprometheus.NewHandler())
	// e.Use(middlewares.Healthz())
	// if config.Log.Level == enums.LogLevelDebug || config.Log.Level == enums.LogLevelTrace {
	// 	pprof.Register(e, consts.PprofPath)
	// }

	echoServer, err := cmds.NewEchoServer(digCon)
	if err != nil {
		return fmt.Errorf("failed to new echo server: %v", err)
	}

	go func() {
		log.Info().Str("addr", consts.WorkerPort).Msg("Server listening")
		err = echoServer.Start(consts.WorkerPort)
		if err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Listening on interface failed")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = echoServer.Shutdown(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Server shutdown failed")
	}

	graceful.Shutdown()

	return nil
}
