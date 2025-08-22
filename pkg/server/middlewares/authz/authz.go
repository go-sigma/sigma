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

package authz

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
)

// Config defines the config for CasbinAuth middleware
type Config struct {
	// Skipper defines a function to skip middleware
	Skipper middleware.Skipper
	// DigCon is the dig container
	DigCon *dig.Container
}

// AuthzWithConfig returns a CasbinAuth middleware with config
func AuthzWithConfig(config Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				log.Debug().Msg("skipping auth middleware, allowing request")
				return next(c)
			}

			requester := c.Request()
			requestUri := strings.TrimSpace(requester.RequestURI)
			requestMethod := requester.Method

			var isDistribution bool
			if strings.HasPrefix(requestUri, "/v2") {
				isDistribution = true
			}

			user, ok := utils.GetFromCtx[*models.User](c, consts.ContextUser)
			if !ok {
				log.Error().Msg("get user from header failed")
				return errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized)
			}
			// admin or root can be access all of resources
			if user.Role == enums.UserRoleAdmin || user.Role == enums.UserRoleRoot {
				log.Debug().Msg("skipping auth middleware, allowing admin and root can access all of resources")
				return next(c)
			}
			// anonymous can only access GET request
			// if user.Role == enums.UserRoleAnonymous {
			// 	if requestMethod == http.MethodGet {
			// 		return next(c)
			// 	} else {
			// 		if isDistribution {
			// 			return errcode.NewDSError(c, errcode.DSErrCodeUnauthorized)
			// 		}
			// 		return errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized, "Authorization failed")
			// 	}
			// }

			nsSvcFactory := utils.MustGetObjFromDigCon[dao.NamespaceServiceFactory](config.DigCon)
			nsSvc := nsSvcFactory.New()

			switch {
			case strings.HasPrefix(requestUri, "/v2"):

			case strings.HasPrefix(requestUri, consts.APIV1):
				switch {
				case strings.HasPrefix(requestUri, fmt.Sprintf("%s/namespaces/", consts.APIV1)):
					if requestUri == fmt.Sprintf("%s/namespaces/", consts.APIV1) {
						return next(c)
					} else {
						namespaceID := strings.TrimSpace(c.Param("namespace_id"))
						if namespaceID == "" {
							return errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized, "Authorization failed")
						}
						nsID, _ := strconv.ParseInt(namespaceID, 10, 64)
						namespace, err := nsSvc.Get(c.Request().Context(), nsID)
						if err != nil {
							log.Error().Err(err).Msg("get namespace failed")
							return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("get namespace failed: %v", err))
						}
						// all of user can get public namespace
						if namespace.Visibility == enums.VisibilityPublic && requestMethod == http.MethodGet {
							return next(c)
						}
						passed, matched, err := dal.AuthEnforcer.Enforcer.EnforceEx(strconv.FormatInt(user.ID, 10), namespace.Name, requestUri, requestMethod)
						if err != nil {
							log.Error().Err(err).Msg("casbin auth failed")
							if isDistribution {
								return errcode.NewDSError(c, errcode.DSErrCodeUnknown)
							}
							return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("get scope from database failed: %v", err))
						}
						log.Debug().Strs("matched", matched).Bool("result", passed).Msg("matched")
						if !passed {
							if isDistribution {
								return errcode.NewDSError(c, errcode.DSErrCodeUnauthorized)
							}
							return errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized, "Authorization failed")
						}
						return next(c)
					}
				default:
					log.Error().Msg("url not match any rule")
					return errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized, "Authorization failed")
				}
			default:
				return next(c)
			}

			fmt.Println(user.ID, "namespace", requestUri, "public", requestMethod, isDistribution, c.Param("id"))
			passed, matched, err := dal.AuthEnforcer.Enforcer.EnforceEx(strconv.FormatInt(user.ID, 10), "namespace", requestUri, "public", requestMethod)
			if err != nil {
				if isDistribution {
					return errcode.NewDSError(c, errcode.DSErrCodeUnknown)
				}
				return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("get scope from database failed: %v", err))
			}
			log.Debug().Strs("matched", matched).Bool("result", passed).Msg("matched")
			if !passed {
				if isDistribution {
					return errcode.NewDSError(c, errcode.DSErrCodeUnauthorized)
				}
				return errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized, "Authorization failed")
			}
			return next(c)
		}
	}
}
