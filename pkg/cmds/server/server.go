/*
Copyright © 2023 Tosone <i@tosone.cn>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
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
	"github.com/ximager/ximager/pkg/storage"
)

// Serve starts the server
func Serve() error {
	debugAddr := viper.GetString("http.debug.addr")
	if debugAddr != "" {
		go func(addr string) {
			log.Info().Str("addr", addr).Msg("Debug server listening")
			err := http.ListenAndServe(addr, nil)
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
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	return nil
}
