package worker

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/ximager/ximager/pkg/daemon"
	"github.com/ximager/ximager/pkg/middlewares"
)

// Worker is the worker initialization
func Worker() error {
	err := daemon.Initialize()
	if err != nil {
		return err
	}

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Use(middlewares.Healthz())

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
