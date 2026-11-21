// Copyright 2025 sigma
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

package web

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/logger"
)

// Serve runs the standalone web front end until SIGINT or SIGTERM. It only
// serves the built UI and its health endpoint; the API and the registry stay
// on the server and distribution services.
func Serve(cfg *config.Configuration) error {
	static := NewStatic(cfg)
	if static == nil {
		return fmt.Errorf("web root %q is not usable, check http.static.dir", cfg.HTTP.Static.Dir)
	}

	gin.SetMode(gin.ReleaseMode)
	e := gin.New()
	e.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})
	e.NoRoute(static.Serve)

	httpServer := &http.Server{
		Addr:              consts.WebPort,
		Handler:           e,
		ReadHeaderTimeout: cfg.HTTP.Timeout.ReadHeaderTimeout(),
		ReadTimeout:       cfg.HTTP.Timeout.ReadTimeout(),
		WriteTimeout:      cfg.HTTP.Timeout.WriteTimeout(),
		IdleTimeout:       cfg.HTTP.Timeout.IdleTimeout(),
	}

	go func() {
		slog.Info("web listening", "addr", httpServer.Addr, "dir", cfg.HTTP.Static.Dir)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("listening on interface failed", "err", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		slog.Error("web server shutdown failed", "err", err)
	}
	return nil
}
