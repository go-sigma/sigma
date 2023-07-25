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
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo-contrib/pprof"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/daemon"
	"github.com/go-sigma/sigma/pkg/middlewares"

	_ "github.com/go-sigma/sigma/pkg/daemon/gc"
	_ "github.com/go-sigma/sigma/pkg/daemon/sbom"
	_ "github.com/go-sigma/sigma/pkg/daemon/vulnerability"
)

// Worker is the worker initialization
func Worker() error {
	err := daemon.InitializeServer()
	if err != nil {
		return err
	}

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Use(middlewares.Healthz())
	if viper.GetInt("log.level") < 1 {
		pprof.Register(e)
	}

	go func() {
		log.Info().Str("addr", consts.WorkerPort).Msg("Server listening")
		err = e.Start(consts.WorkerPort)
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
	err = e.Shutdown(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Server shutdown failed")
	}

	return nil
}
