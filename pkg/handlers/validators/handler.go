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

package validators

import (
	"path"
	"reflect"

	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/handlers"
	"github.com/go-sigma/sigma/pkg/middlewares/authn"
	"github.com/go-sigma/sigma/pkg/middlewares/authz"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/echoplus"
)

// Handler ...
type Handler interface {
	// GetReference handles the validate reference request
	GetReference(c echo.Context) error
	// GetTag handles the validate tag request
	GetTag(c echo.Context) error
	// GetPassword handles the validate password request
	GetPassword(c echo.Context) error
	// ValidateCron handles the validate cron request
	ValidateCron(c echo.Context) error
	// ValidateRegexp handles the validate regex request
	ValidateRegexp(c echo.Context) error
}

var _ Handler = &handler{}

type handler struct{}

// handlerNew creates a new instance of the distribution handlers
func handlerNew() Handler {
	return &handler{}
}

type factory struct{}

// Initialize initializes the namespace handlers
func (f factory) Initialize(digCon *dig.Container) error {
	handler := handlerNew()
	echo := utils.MustGetObjFromDigCon[*echo.Echo](digCon)
	plus := echoplus.New(echo.Group(consts.APIV1 + "/validators"))
	plus.Get("/reference", &authn.AuthnConfig{Skip: true}, &authz.AuthzConfig{Skip: true}, handler.GetReference)
	plus.Get("/tag", &authn.AuthnConfig{Skip: true}, &authz.AuthzConfig{Skip: true}, handler.GetTag)
	plus.Post("/password", &authn.AuthnConfig{Skip: true}, &authz.AuthzConfig{Skip: true}, handler.GetPassword)
	plus.Post("/cron", &authn.AuthnConfig{Skip: true}, &authz.AuthzConfig{Skip: true}, handler.ValidateCron)
	plus.Post("/regexp", &authn.AuthnConfig{Skip: true}, &authz.AuthzConfig{Skip: true}, handler.ValidateRegexp)
	return nil
}

func init() {
	utils.PanicIf(handlers.RegisterRouterFactory(path.Base(reflect.TypeOf(factory{}).PkgPath()), &factory{}))
}
