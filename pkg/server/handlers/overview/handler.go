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

package overview

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/consts"
	svcoverview "github.com/go-sigma/sigma/pkg/service/overview"
)

// Handler is the interface for the overview handlers.
type Handler interface {
	// GetOverview handles the home dashboard request.
	GetOverview(c *gin.Context)
	// GetDailyActivity handles the daily push and pull request.
	GetDailyActivity(c *gin.Context)
	// GetDailyBuilds handles the daily build request.
	GetDailyBuilds(c *gin.Context)
}

var _ Handler = &handler{}

type handler struct {
	dig.In

	OverviewSvc svcoverview.Service
}

// Initialize registers the handler routes.
func Initialize(e *gin.Engine, h handler) error {
	group := e.Group(consts.APIV1)
	group.GET("/overview", h.GetOverview)
	group.GET("/activity/daily", h.GetDailyActivity)
	group.GET("/builds/daily", h.GetDailyBuilds)
	return nil
}
