// Copyright 2026 sigma
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

package analytics

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/authz"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/server"
	svcanalytics "github.com/go-sigma/sigma/pkg/service/analytics"
)

// Handler is the interface for analytics handlers.
type Handler interface {
	GetUserPushHeatmap(c *gin.Context)
	GetNamespaceTrends(c *gin.Context)
}

var _ Handler = &handler{}

type handler struct {
	dig.In

	AnalyticsSvc svcanalytics.Service
	Authorizer   authz.Authorizer
}

// Initialize registers the handler routes.
func Initialize(e *gin.Engine, h handler) error {
	group := e.Group(consts.APIV1)
	group.GET("/users/:user_id/activity/heatmap", server.Wrap(h.GetUserPushHeatmap))
	group.GET("/namespaces/:namespace_id/activity/trends", server.Wrap(h.GetNamespaceTrends))
	return nil
}
