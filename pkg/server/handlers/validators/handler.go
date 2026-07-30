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
	"github.com/gin-gonic/gin"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/server"
)

// Handler ...
type Handler interface {
	// GetReference handles the validate reference request
	GetReference(c *gin.Context, req *api.GetValidatorReferenceRequest)
	// GetTag handles the validate tag request
	GetTag(c *gin.Context, req *api.GetValidatorTagRequest)
	// GetPassword handles the validate password request
	GetPassword(c *gin.Context, req *api.ValidatePasswordRequest)
	// ValidateCron handles the validate cron request
	ValidateCron(c *gin.Context, req *api.ValidateCronRequest)
	// ValidateRegexp handles the validate regex request
	ValidateRegexp(c *gin.Context, req *api.ValidateRegexpRequest)
}

var _ Handler = &handler{}

type handler struct{}

// Initialize registers the handler routes.
func Initialize(e *gin.Engine) error {
	h := &handler{}
	group := e.Group(consts.APIV1 + "/validators")
	group.GET("/reference", server.WrapRequest(h.GetReference))
	group.GET("/tag", server.WrapRequest(h.GetTag))
	group.POST("/password", server.WrapRequest(h.GetPassword))
	group.POST("/cron", server.WrapRequest(h.ValidateCron))
	group.POST("/regexp", server.WrapRequest(h.ValidateRegexp))
	return nil
}
