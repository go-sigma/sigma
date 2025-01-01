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
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/middlewares/extractor"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// AuthMapper ...
var AuthMapper = make(map[*regexp.Regexp]*AuthzConfig)

// AuthzWithConfig returns a CasbinAuth middleware with config
func AuthzWithConfig(config Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				log.Trace().Msg("Skipping auth middleware, allowing request")
				return next(c)
			}

			requester := c.Request()
			requestUri := requester.RequestURI
			requestMethod := requester.Method

			var isDistribution bool
			if strings.HasPrefix(requestUri, "/v2") {
				isDistribution = true
			}

			echo := utils.MustGetObjFromDigCon[*echo.Echo](config.DigCon)

			authConfig := authMatch(echo, requestMethod, requestUri)
			if authConfig == nil || authConfig.Skip {
				return next(c)
			}
			for index, source := range authConfig.Sources {
				if len(ptr.To(source.Scope.ScopeValue.Position)) > 0 && len(ptr.To(source.Scope.ScopeValue.Key)) > 0 {
					var position = ptr.To(source.Scope.ScopeValue.Position)
					var key = ptr.To(source.Scope.ScopeValue.Key)
					extractors := extractor.MustCreateExtractors(fmt.Sprintf("%s:%s", position, key))
					var values []string
					for _, extractor := range extractors {
						vals, err := extractor(c)
						if err != nil {
							return nil // TODO
						}
						values = append(values, vals...)
					}
					authConfig.Sources[index].Resource.ResourceValue.Values = values
				}
			}

			authRuleFactory := utils.MustGetObjFromDigCon[dao.AuthRuleServiceFactory](config.DigCon)
			authRuleSvc := authRuleFactory.New()

			var scopes = make([]dao.ScopeItem, 0, 20)

			for _, source := range authConfig.Sources {
				scopes = append(scopes, dao.ScopeItem{
					ScopeType:  source.Scope.ScopeType,
					ScopeValue: source.Scope.ScopeValue.Value,
				})
			}

			ctx := context.Background()
			authRules, err := authRuleSvc.ListByScope(ctx, scopes)
			if err != nil {
				return nil // TODO
			}
			if len(authRules) == 0 {
				if isDistribution {
					return xerrors.NewDSError(c, xerrors.DSErrCodeUnauthorized)
				}
				return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized, "Authorization failed")
			}

			for _, rule := range authRules {
				for index, source := range authConfig.Sources {
					if rule.ScopeType == source.Scope.ScopeType && rule.ScopeValue == source.Scope.ScopeValue.Value {
						if strings.EqualFold(rule.Role.Action.String(), requestMethod) && rule.Role.Resource == source.Resource.ResourceType {
							if !source.Matched && source.Effect != enums.AuthEffectDeny {
								// for _, value := range source.Resource.ResourceValue.Values {
								// 	// rule.Role.
								// }
								authConfig.Sources[index].Matched = true
								source.Effect = rule.Role.Effect
							}
						}
					}
				}
			}
			return next(c)
		}
	}
}

func authMatch(echo *echo.Echo, method, uri string) *AuthzConfig {
	if strings.HasPrefix(uri, "/api/v1/") {
		ctx := echo.AcquireContext()
		defer echo.ReleaseContext(ctx)
		echo.Router().Find(method, uri, ctx)
		matchedPath := ctx.Path()
		if matchedPath == "" {
			return nil
		}
		return authMapperMatcher(uri)
	} else if strings.HasPrefix(uri, "/v2/") {
		return authMapperMatcher(uri)
	}
	return nil
}

func authMapperMatcher(uri string) *AuthzConfig {
	for reg, config := range AuthMapper {
		if !reg.MatchString(uri) {
			continue
		}
		if config == nil {
			continue
		}
		return config
	}
	return nil
}
