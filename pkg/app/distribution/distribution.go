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
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/app"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/graceful"
	"github.com/go-sigma/sigma/pkg/logger"
	"github.com/go-sigma/sigma/pkg/server/handlers"
)

func Serve(digCon *dig.Container) error {
	var ginServer *gin.Engine
	var distParams app.GinServerParams
	err := digCon.Invoke(func(params app.GinServerParams) error {
		var invokeErr error
		distParams = params
		ginServer, invokeErr = app.NewGinServer(params)
		return invokeErr
	})
	if err != nil {
		return fmt.Errorf("failed to new gin server: %v", err)
	}

	err = digCon.Provide(func() *gin.Engine { return ginServer })
	if err != nil {
		return fmt.Errorf("failed to provide gin engine: %v", err)
	}

	var cfg *config.Configuration
	err = digCon.Invoke(func(c *config.Configuration) { cfg = c })
	if err != nil {
		return err
	}

	err = handlers.InitializeDistribution(digCon)
	if err != nil {
		return fmt.Errorf("failed to initialize distribution handlers: %v", err)
	}

	httpServer := &http.Server{
		Addr:              consts.DistributionPort,
		Handler:           ginServer,
		ReadHeaderTimeout: cfg.HTTP.Timeout.ReadHeaderTimeout(),
		ReadTimeout:       cfg.HTTP.Timeout.ReadTimeout(),
		WriteTimeout:      cfg.HTTP.Timeout.WriteTimeout(),
		IdleTimeout:       cfg.HTTP.Timeout.IdleTimeout(),
	}

	// Verify all required dependencies are reachable before accepting traffic.
	startupCtx, startupCancel := context.WithTimeout(context.Background(), app.StartupTimeout)
	defer startupCancel()
	err = app.ValidateOnStartup(startupCtx, app.NewReadinessResources(distParams))
	if err != nil {
		return fmt.Errorf("startup dependency check failed: %w", err)
	}
	app.PrewarmMetadataCache(startupCtx, digCon)

	go func() {
		slog.Info("server listening", "addr", consts.DistributionPort)
		if cfg.HTTP.TLS.Enabled {
			crtBytes, e := os.ReadFile(cfg.HTTP.TLS.Certificate)
			if e != nil {
				logger.Fatal("read certificate failed", "err", e, "certificate", cfg.HTTP.TLS.Certificate)
				return
			}
			keyBytes, e := os.ReadFile(cfg.HTTP.TLS.Key)
			if e != nil {
				logger.Fatal("read key failed", "err", e, "key", cfg.HTTP.TLS.Key)
				return
			}
			cert, e := tls.X509KeyPair(crtBytes, keyBytes)
			if e != nil {
				logger.Fatal("parse X509 key pair failed", "err", e)
				return
			}
			httpServer.TLSConfig = &tls.Config{Certificates: []tls.Certificate{cert}}
			e = httpServer.ListenAndServeTLS("", "")
			if e != nil && e != http.ErrServerClosed {
				logger.Fatal("listening on interface failed", "err", e)
			}
		} else {
			e := httpServer.ListenAndServe()
			if e != nil && e != http.ErrServerClosed {
				logger.Fatal("listening on interface failed", "err", e)
			}
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a
	// single shared 30s budget covering HTTP drain + shutdown hooks.
	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	// Flip readiness to 503 and give the load balancer a moment to stop
	// routing new connections before we start rejecting in-flight requests.
	app.SetDraining(true)
	slog.Info("draining traffic before shutdown")
	time.Sleep(app.DrainDelay)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err = httpServer.Shutdown(ctx)
	if err != nil {
		slog.Error("server shutdown failed", "err", err)
	}

	if err := graceful.Shutdown(ctx); err != nil {
		slog.Error("graceful shutdown failed", "err", err)
	}

	return nil
}
