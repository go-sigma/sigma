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

package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/ximager/ximager/pkg/handlers"
	"github.com/ximager/ximager/pkg/middlewares"
	"github.com/ximager/ximager/pkg/storage"
)

// Serve starts the server
func Serve() error {
	debugAddr := viper.GetString("http.debug.addr")
	if debugAddr != "" {
		go func(addr string) {
			log.Info().Str("addr", addr).Msg("Debug server listening")
			server := &http.Server{
				Addr:              addr,
				ReadHeaderTimeout: 3 * time.Second,
			}
			err := server.ListenAndServe()
			if err != nil {
				log.Fatal().Err(err).Msg("Listening on debug interface failed")
			}
		}(debugAddr)
	}
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Use(echo.MiddlewareFunc(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			log.Info().Str("method", c.Request().Method).Str("path", c.Request().URL.Path).Str("query", c.Request().URL.RawQuery).Msg("Request")
			return next(c)
		}
	}))
	e.Use(middleware.CORS())
	e.Use(middlewares.Healthz())

	err := handlers.Initialize(e)
	if err != nil {
		return err
	}

	err = storage.Initialize()
	if err != nil {
		return err
	}

	go func() {
		log.Info().Str("addr", viper.GetString("http.addr")).Msg("Server listening")
		err = e.Start(viper.GetString("http.addr"))
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
