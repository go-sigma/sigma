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

package oauth2

import (
	"fmt"
	"path"
	"reflect"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"go.uber.org/dig"
	"golang.org/x/exp/slices"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/handlers"
	"github.com/go-sigma/sigma/pkg/middlewares"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/token"
)

// Handler is the interface for the oauth2 handlers
type Handler interface {
	// Callback handles the callback request
	Callback(c echo.Context) error
	// ClientID handles the client id request
	ClientID(c echo.Context) error
	// RedirectCallback ...
	RedirectCallback(c echo.Context) error
}

var _ Handler = &handler{}

type handler struct {
	config             *configs.Configuration
	tokenService       token.Service
	userServiceFactory dao.UserServiceFactory
}

// handlerNew creates a new instance of the distribution handlers
func handlerNew(digCon *dig.Container) (Handler, error) {
	var tokenService token.Service
	userServiceFactory := dao.NewUserServiceFactory()
	config := configs.GetConfiguration()
	return &handler{
		config:             config,
		tokenService:       tokenService,
		userServiceFactory: userServiceFactory,
	}, nil
}

type factory struct{}

// Initialize initializes the namespace handlers
func (f factory) Initialize(digCon *dig.Container) error {
	e := utils.MustGetObjFromDigCon[*echo.Echo](digCon)
	oauth2Group := e.Group(consts.APIV1 + "/oauth2")
	repositoryHandler, err := handlerNew(digCon)
	if err != nil {
		return err
	}

	oauth2Mapper := viper.GetStringMap("auth.oauth2")
	var skipAuths = make([]string, 0, len(oauth2Mapper))
	for key := range oauth2Mapper {
		skipAuths = append(skipAuths, fmt.Sprintf("get:/api/v1/oauth2/%s/client_id", strings.ToLower(key)))
		skipAuths = append(skipAuths, fmt.Sprintf("get:/api/v1/oauth2/%s/callback", strings.ToLower(key)))
		skipAuths = append(skipAuths, fmt.Sprintf("get:/api/v1/oauth2/%s/redirect_callback", strings.ToLower(key)))
	}
	oauth2Group.Use(middlewares.AuthWithConfig(middlewares.AuthConfig{
		Skipper: func(c echo.Context) bool {
			authStr := strings.ToLower(fmt.Sprintf("%s:%s", c.Request().Method, c.Request().URL.Path))
			return slices.Contains(skipAuths, authStr)
		},
	}))
	oauth2Group.GET("/:provider/callback", repositoryHandler.Callback)
	oauth2Group.GET("/:provider/client_id", repositoryHandler.ClientID)
	oauth2Group.GET("/:provider/redirect_callback", repositoryHandler.RedirectCallback)
	return nil
}

func init() {
	utils.PanicIf(handlers.RegisterRouterFactory(path.Base(reflect.TypeOf(factory{}).PkgPath()), &factory{}))
}
