// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2017 LabStack and Echo contributors

// Copyright 2024 sigma
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

package middlewares

import (
	"errors"
	"net/http"

	"github.com/casbin/casbin/v2"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type (
	// Config defines the config for CasbinAuth middleware.
	Config struct {
		// Skipper defines a function to skip middleware.
		Skipper middleware.Skipper

		// Enforcer CasbinAuth main rule.
		// One of Enforcer or EnforceHandler fields is required.
		Enforcer *casbin.SyncedEnforcer

		// EnforceHandler is custom callback to handle enforcing.
		// One of Enforcer or EnforceHandler fields is required.
		EnforceHandler func(c echo.Context, user string) (bool, error)

		// Method to get the username - defaults to using basic auth
		UserGetter func(c echo.Context) (string, error)

		// Method to handle errors
		ErrorHandler func(c echo.Context, internal error, proposedStatus int) error
	}
)

var (
	// DefaultConfig is the default CasbinAuth middleware config.
	DefaultConfig = Config{
		Skipper: middleware.DefaultSkipper,
		UserGetter: func(c echo.Context) (string, error) {
			username, _, _ := c.Request().BasicAuth()
			return username, nil
		},
		ErrorHandler: func(c echo.Context, internal error, proposedStatus int) error {
			err := echo.NewHTTPError(proposedStatus, internal.Error())
			err.Internal = internal
			return err
		},
	}
)

// AuthzWithConfig returns a CasbinAuth middleware with config.
func AuthzWithConfig(config Config) echo.MiddlewareFunc {
	if config.Enforcer == nil {
		panic("casbin middleware Enforcer field must be set")
	}
	if config.Skipper == nil {
		config.Skipper = DefaultConfig.Skipper
	}
	if config.UserGetter == nil {
		config.UserGetter = DefaultConfig.UserGetter
	}
	if config.ErrorHandler == nil {
		config.ErrorHandler = DefaultConfig.ErrorHandler
	}
	if config.EnforceHandler == nil {
		config.EnforceHandler = func(c echo.Context, user string) (bool, error) {
			return config.Enforcer.Enforce(user, c.Request().URL.Path, c.Request().Method)
		}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}
			user, err := config.UserGetter(c)
			if err != nil {
				return config.ErrorHandler(c, err, http.StatusForbidden)
			}
			pass, err := config.EnforceHandler(c, user)
			if err != nil {
				return config.ErrorHandler(c, err, http.StatusInternalServerError)
			}
			if !pass {
				return config.ErrorHandler(c, errors.New("enforce did not pass"), http.StatusForbidden)
			}
			return next(c)
		}
	}
}
