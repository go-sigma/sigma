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
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/app"
	"github.com/go-sigma/sigma/pkg/background/buildrunner"
	"github.com/go-sigma/sigma/pkg/background/cronjob"
	"github.com/go-sigma/sigma/pkg/background/daemon"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/graceful"
	"github.com/go-sigma/sigma/pkg/infra/workq"
	"github.com/go-sigma/sigma/pkg/logger"
)

// Worker is the worker initialization
func Worker(digCon *dig.Container) error {
	cfg := config.GetConfig()
	err := buildrunner.Initialize(cfg)
	if err != nil {
		return err
	}

	var ginServer *gin.Engine
	var workerParams app.GinServerParams
	err = digCon.Invoke(func(params app.GinServerParams) error {
		var invokeErr error
		workerParams = params
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

	err = daemon.Initialize(digCon)
	if err != nil {
		return fmt.Errorf("failed to initialize daemon: %v", err)
	}

	err = digCon.Invoke(workq.InitConsumer)
	if err != nil {
		return fmt.Errorf("failed to initialize work queue consumer: %v", err)
	}
	if err = cronjob.Initialize(digCon); err != nil {
		return fmt.Errorf("failed to initialize cronjob: %v", err)
	}
	graceful.RunAtShutdown("cronjob", 40, cronjob.DeInitialize)

	httpServer := &http.Server{
		Addr:              consts.WorkerPort,
		Handler:           ginServer,
		ReadHeaderTimeout: 30 * time.Second,
	}

	// Verify all required dependencies are reachable before accepting traffic.
	startupCtx, startupCancel := context.WithTimeout(context.Background(), app.StartupTimeout)
	defer startupCancel()
	if err := app.ValidateOnStartup(startupCtx, app.NewReadinessResources(workerParams)); err != nil {
		return fmt.Errorf("startup dependency check failed: %w", err)
	}

	go func() {
		slog.Info("server listening", "addr", consts.WorkerPort)
		err = httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logger.Fatal("listening on interface failed", "err", err)
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
