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

package distribution

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/rs/zerolog/log"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/cmds"
	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/graceful"
	"github.com/go-sigma/sigma/pkg/server/handlers"
	"github.com/go-sigma/sigma/pkg/modules/workq"
	"github.com/go-sigma/sigma/pkg/storage"
	"github.com/go-sigma/sigma/pkg/utils"
)

func Serve(digCon *dig.Container) error {
	echoServer, err := cmds.NewEchoServer(digCon)
	if err != nil {
		return fmt.Errorf("failed to new echo server: %v", err)
	}

	config := utils.MustGetObjFromDigCon[configs.Configuration](digCon)
	err = workq.InitProducer(config)
	if err != nil {
		return err
	}

	handlers.InitializeDistribution(dig.New())
	err = storage.Initialize(config)
	if err != nil {
		return err
	}

	go func() {
		log.Info().Str("addr", consts.DistributionPort).Msg("Server listening")
		if config.HTTP.TLS.Enabled {
			crtBytes, err := os.ReadFile(config.HTTP.TLS.Certificate)
			if err != nil {
				log.Fatal().Err(err).Str("certificate", config.HTTP.TLS.Certificate).Msgf("Read certificate failed") // TODO: fatal error
				return
			}
			keyBytes, err := os.ReadFile(config.HTTP.TLS.Key)
			if err != nil {
				log.Fatal().Err(err).Str("key", config.HTTP.TLS.Key).Msgf("Read key failed")
				return
			}
			err = echoServer.StartTLS(consts.DistributionPort, crtBytes, keyBytes)
			if err != http.ErrServerClosed {
				log.Fatal().Err(err).Msg("Listening on interface failed")
			}
		} else {
			err = echoServer.Start(consts.DistributionPort)
			if err != http.ErrServerClosed {
				log.Fatal().Err(err).Msg("Listening on interface failed")
			}
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
