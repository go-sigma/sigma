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

package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/labstack/echo-contrib/pprof"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/spf13/viper"

	"github.com/go-sigma/sigma/pkg/builder"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/daemon"
	"github.com/go-sigma/sigma/pkg/handlers"
	"github.com/go-sigma/sigma/pkg/middlewares"
	"github.com/go-sigma/sigma/pkg/storage"
	"github.com/go-sigma/sigma/pkg/utils/serializer"
	"github.com/go-sigma/sigma/web"
)

// ServerConfig ...
type ServerConfig struct {
	WithoutDistribution bool
	WithoutWorker       bool
	WithoutWeb          bool
}

// Serve starts the server
func Serve(config ServerConfig) error {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{Level: 5}))
	e.Use(echo.MiddlewareFunc(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			n := next(c)
			log.Debug().
				Str("method", c.Request().Method).
				Str("path", c.Request().URL.Path).
				Str("query", c.Request().URL.RawQuery).
				Interface("req-header", c.Request().Header).
				Interface("resp-header", c.Response().Header()).
				Int("status", c.Response().Status).
				Msg("Request debugger")
			return n
		}
	}))

	e.Use(middleware.CORS())
	e.Use(middlewares.Healthz())
	e.JSONSerializer = new(serializer.DefaultJSONSerializer)

	if viper.GetInt("log.level") < 1 {
		pprof.Register(e)
	}

	if !config.WithoutDistribution {
		handlers.InitializeDistribution(e)
	}
	if !config.WithoutWorker {
		err := builder.Initialize()
		if err != nil {
			return err
		}
		err = daemon.InitializeServer()
		if err != nil {
			return err
		}
	}
	if !config.WithoutWeb {
		web.RegisterHandlers(e)
	}

	err := handlers.Initialize(e)
	if err != nil {
		return err
	}

	err = storage.Initialize()
	if err != nil {
		return err
	}

	go func() {
		log.Info().Str("addr", consts.ServerPort).Msg("Server listening")
		if viper.GetBool("http.tls.enabled") {
			crtBytes, err := os.ReadFile(viper.GetString("http.tls.certificate"))
			if err != nil {
				log.Fatal().Err(err).Str("certificate", viper.GetString("http.tls.certificate")).Msgf("Read certificate failed")
				return
			}
			keyBytes, err := os.ReadFile(viper.GetString("http.tls.key"))
			if err != nil {
				log.Fatal().Err(err).Str("key", viper.GetString("http.tls.key")).Msgf("Read key failed")
				return
			}
			err = e.StartTLS(consts.ServerPort, crtBytes, keyBytes)
			if err != http.ErrServerClosed {
				log.Fatal().Err(err).Msg("Listening on interface failed")
			}
		} else {
			err = e.Start(consts.ServerPort)
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
	err = e.Shutdown(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Server shutdown failed")
	}

	if viper.GetString("deploy") == "single" && viper.GetString("redis.type") == "internal" {
		_, err := os.Stat(consts.RedisPid)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		pidBytes, err := os.ReadFile(consts.RedisPid)
		if err != nil {
			return err
		}
		pid, err := strconv.ParseInt(string(pidBytes), 10, 0)
		if err != nil {
			return err
		}
		exist, err := process.PidExists(int32(pid))
		if err != nil {
			return err
		}
		if exist {
			ps, err := process.NewProcess(int32(pid))
			if err != nil {
				return err
			}
			err = ps.SendSignal(syscall.SIGTERM)
			if err != nil {
				return err
			}
			maxTimes := 10
			for i := 0; i < maxTimes; i++ {
				exist, err := process.PidExists(int32(pid))
				if err != nil {
					return err
				}
				if !exist {
					break
				}
				log.Info().Msg("Redis process is still here, wait for a moment")
			}
		}
	}

	return nil
}
