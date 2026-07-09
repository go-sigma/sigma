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

package authn

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/consts"
	repouser "github.com/go-sigma/sigma/pkg/dal/repository/user"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/service/password"
	"github.com/go-sigma/sigma/pkg/service/token"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

// Skipper defines a function to skip middleware.
type Skipper func(c *gin.Context) bool

// Config is the configuration for the Auth middleware
type Config struct {
	// Skipper defines a function to skip middleware
	Skipper Skipper
	// Config is the application configuration.
	Config *config.Configuration
	// TokenSvc validates bearer tokens.
	TokenSvc token.Service
	// PasswordSvc verifies basic auth passwords.
	PasswordSvc password.Service
	// UserRepository creates user repositories.
	UserRepository repouser.UserRepository
}

// AuthnConfig ...
type AuthnConfig struct {
	Skip bool
}

// scheme returns the request scheme (http or https), honoring
// X-Forwarded-Proto like echo's c.Scheme() did.
func scheme(c *gin.Context) string {
	if proto := c.GetHeader("X-Forwarded-Proto"); proto != "" {
		return proto
	}
	if c.Request.TLS != nil {
		return "https"
	}
	return "http"
}

// AuthnWithConfig returns a middleware which authenticates requests.
func AuthnWithConfig(config Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if config.Skipper != nil && config.Skipper(c) {
			slog.Debug("skipping auth middleware, allowing request")
			c.Next()
			return
		}

		requestUri := c.Request.RequestURI

		var isDistribution bool
		if strings.HasPrefix(requestUri, "/v2/") {
			isDistribution = true
		}

		req := c.Request
		ctx := req.Context()
		authorization := req.Header.Get(consts.HeaderAuthorization)

		var uid string
		var jti = uuid.NewV7String()
		var err error

		userRepository := config.UserRepository

		switch {
		case strings.HasPrefix(authorization, "Basic"):
			username, pwd, ok := req.BasicAuth()
			if !ok {
				slog.Error("basic auth failed", "Authorization", req.Header.Get(consts.HeaderAuthorization))
				c.Header(consts.HeaderWWWAuthenticate, genWwwAuthenticate(config.Config, req.Host, scheme(c)))
				if isDistribution {
					errcode.NewDSError(c, errcode.DSErrCodeUnauthorized)
				} else {
					errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized, "Basic auth failed")
				}
				c.Abort()
				return
			}

			user, err := userRepository.GetByUsername(ctx, username)
			if err != nil {
				slog.Error("get user by username failed", "err", err)
				c.Header(consts.HeaderWWWAuthenticate, genWwwAuthenticate(config.Config, req.Host, scheme(c)))
				if isDistribution {
					errcode.NewDSError(c, errcode.DSErrCodeUnauthorized)
				} else {
					errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized, "Username or password is not correct")
				}
				c.Abort()
				return
			}
			uid = user.ID

			verify := config.PasswordSvc.Verify(pwd, ptr.To(user.Password))
			if !verify {
				slog.Error("verify password failed", "err", err)
				c.Header(consts.HeaderWWWAuthenticate, genWwwAuthenticate(config.Config, req.Host, scheme(c)))
				if isDistribution {
					errcode.NewDSError(c, errcode.DSErrCodeUnauthorized)
				} else {
					errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized, "Username or password is not correct")
				}
				c.Abort()
				return
			}
		case strings.HasPrefix(authorization, "Bearer"):
			jti, uid, err = config.TokenSvc.Validate(ctx, strings.TrimSpace(strings.TrimPrefix(authorization, "Bearer")))
			if err != nil {
				slog.Error("validate token failed", "err", err)
				c.Header(consts.HeaderWWWAuthenticate, genWwwAuthenticate(config.Config, req.Host, scheme(c)))
				if isDistribution {
					errcode.NewDSError(c, errcode.DSErrCodeUnauthorized)
				} else {
					errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized, err.Error())
				}
				c.Abort()
				return
			}
		default:
			uri := req.URL.Path
			if strings.HasPrefix(uri, "/v2") || uri == "/api/v1/users/self" {
				c.Header(consts.HeaderWWWAuthenticate, genWwwAuthenticate(config.Config, req.Host, scheme(c)))
				if isDistribution {
					errcode.NewDSError(c, errcode.DSErrCodeUnauthorized)
				} else {
					errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized)
				}
				c.Abort()
				return
			}
			userObj, err := userRepository.GetByUsername(ctx, consts.UserAnonymous)
			if err != nil {
				slog.Error("get anonymous user failed", "err", err)
				c.Header(consts.HeaderWWWAuthenticate, genWwwAuthenticate(config.Config, req.Host, scheme(c)))
				if isDistribution {
					errcode.NewDSError(c, errcode.DSErrCodeUnauthorized)
				} else {
					errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized)
				}
				c.Abort()
				return
			}
			uid = userObj.ID
		}

		userObj, err := userRepository.Get(ctx, uid)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				slog.Error("user not found", "err", err)
				if isDistribution {
					errcode.NewDSError(c, errcode.DSErrCodeUnknown)
				} else {
					errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized, err.Error())
				}
				c.Abort()
				return
			}
			slog.Error("get user failed", "err", err)
			if isDistribution {
				errcode.NewDSError(c, errcode.DSErrCodeUnknown)
			} else {
				errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, err.Error())
			}
			c.Abort()
			return
		}

		c.Set(consts.ContextUser, userObj)
		c.Set(consts.ContextJti, jti)

		c.Next()
	}
}

func genWwwAuthenticate(config *config.Configuration, host, schema string) string {
	realm := fmt.Sprintf("%s://%s%s/tokens", schema, host, consts.APIV1)
	if config.Auth.Token.Realm != "" {
		realm = config.Auth.Token.Realm
	}
	service := consts.AppName
	if config.Auth.Token.Service != "" {
		service = config.Auth.Token.Service
	}
	return fmt.Sprintf("Bearer realm=\"%s\",service=\"%s\"", realm, service)
}
