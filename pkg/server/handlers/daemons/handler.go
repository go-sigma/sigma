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

package daemons

import (
	"path"
	"reflect"

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/server"
	"github.com/go-sigma/sigma/pkg/server/handlers"
	"github.com/go-sigma/sigma/pkg/service/daemons"
	"github.com/go-sigma/sigma/pkg/utils"
)

// Handler is the interface for the gc handlers
type Handler interface {
	// UpdateGcTagRule ...
	UpdateGcTagRule(c *gin.Context, req *api.UpdateGcTagRuleRequest)
	// GetGcTagRule ...
	GetGcTagRule(c *gin.Context, req *api.GetGcTagRuleRequest)
	// GetGcTagLatestRunner ...
	GetGcTagLatestRunner(c *gin.Context, req *api.GetGcTagLatestRunnerRequest)
	// CreateGcTagRunner ...
	CreateGcTagRunner(c *gin.Context, req *api.CreateGcTagRunnerRequest)
	// ListGcTagRunners ...
	ListGcTagRunners(c *gin.Context, req *api.ListGcTagRunnersRequest)
	// GetGcTagRunner ...
	GetGcTagRunner(c *gin.Context, req *api.GetGcTagRunnerRequest)
	// ListGcTagRecords ...
	ListGcTagRecords(c *gin.Context, req *api.ListGcTagRecordsRequest)
	// GetGcTagRecord ...
	GetGcTagRecord(c *gin.Context, req *api.GetGcTagRecordRequest)

	// UpdateGcRepositoryRule ...
	UpdateGcRepositoryRule(c *gin.Context, req *api.UpdateGcRepositoryRuleRequest)
	// GetGcRepositoryRule ...
	GetGcRepositoryRule(c *gin.Context, req *api.GetGcRepositoryRuleRequest)
	// GetGcRepositoryLatestRunner ...
	GetGcRepositoryLatestRunner(c *gin.Context, req *api.GetGcRepositoryLatestRunnerRequest)
	// CreateGcRepositoryRunner ...
	CreateGcRepositoryRunner(c *gin.Context, req *api.CreateGcRepositoryRunnerRequest)
	// ListGcRepositoryRunners ...
	ListGcRepositoryRunners(c *gin.Context, req *api.ListGcRepositoryRunnersRequest)
	// GetGcRepositoryRunner ...
	GetGcRepositoryRunner(c *gin.Context, req *api.GetGcRepositoryRunnerRequest)
	// ListGcRepositoryRecords ...
	ListGcRepositoryRecords(c *gin.Context, req *api.ListGcRepositoryRecordsRequest)
	// GetGcRepositoryRecord ...
	GetGcRepositoryRecord(c *gin.Context, req *api.GetGcRepositoryRecordRequest)

	// UpdateGcArtifactRule ...
	UpdateGcArtifactRule(c *gin.Context, req *api.UpdateGcArtifactRuleRequest)
	// GetGcArtifactRule ...
	GetGcArtifactRule(c *gin.Context, req *api.GetGcArtifactRuleRequest)
	// GetGcArtifactLatestRunner ...
	GetGcArtifactLatestRunner(c *gin.Context, req *api.GetGcArtifactLatestRunnerRequest)
	// CreateGcArtifactRunner ...
	CreateGcArtifactRunner(c *gin.Context, req *api.CreateGcArtifactRunnerRequest)
	// ListGcArtifactRunners ...
	ListGcArtifactRunners(c *gin.Context, req *api.ListGcArtifactRunnersRequest)
	// GetGcArtifactRunner ...
	GetGcArtifactRunner(c *gin.Context, req *api.GetGcArtifactRunnerRequest)
	// ListGcArtifactRecords ...
	ListGcArtifactRecords(c *gin.Context, req *api.ListGcArtifactRecordsRequest)
	// GetGcArtifactRecord ...
	GetGcArtifactRecord(c *gin.Context, req *api.GetGcArtifactRecordRequest)

	// UpdateGcBlobRule ...
	UpdateGcBlobRule(c *gin.Context, req *api.UpdateGcBlobRuleRequest)
	// GetGcBlobRule ...
	GetGcBlobRule(c *gin.Context, req *api.GetGcBlobRuleRequest)
	// GetGcBlobLatestRunner ...
	GetGcBlobLatestRunner(c *gin.Context, req *api.GetGcBlobLatestRunnerRequest)
	// CreateGcBlobRunner ...
	CreateGcBlobRunner(c *gin.Context, req *api.CreateGcBlobRunnerRequest)
	// ListGcBlobRunners ...
	ListGcBlobRunners(c *gin.Context, req *api.ListGcBlobRunnersRequest)
	// GetGcBlobRunner ...
	GetGcBlobRunner(c *gin.Context, req *api.GetGcBlobRunnerRequest)
	// ListGcBlobRecords ...
	ListGcBlobRecords(c *gin.Context, req *api.ListGcBlobRecordsRequest)
	// GetGcBlobRecord ...
	GetGcBlobRecord(c *gin.Context, req *api.GetGcBlobRecordRequest)
}

var _ Handler = &handler{}

type handler struct {
	dig.In

	DaemonSvc daemons.Service
}

type factory struct{}

// Initialize initializes the namespace handlers
func (f factory) Initialize(digCon *dig.Container) error {
	return digCon.Invoke(func(e *gin.Engine, h handler) error {
		daemonGroup := e.Group(consts.APIV1 + "/daemons")

		daemonGroup.PUT("/gc-repository/:namespace_id/", server.WrapRequest(h.UpdateGcRepositoryRule))
		daemonGroup.GET("/gc-repository/:namespace_id/", server.WrapRequest(h.GetGcRepositoryRule))
		daemonGroup.GET("/gc-repository/:namespace_id/runners/latest", server.WrapRequest(h.GetGcRepositoryLatestRunner))
		daemonGroup.POST("/gc-repository/:namespace_id/runners/", server.WrapRequest(h.CreateGcRepositoryRunner))
		daemonGroup.GET("/gc-repository/:namespace_id/runners/", server.WrapRequest(h.ListGcRepositoryRunners))
		daemonGroup.GET("/gc-repository/:namespace_id/runners/:runner_id", server.WrapRequest(h.GetGcRepositoryRunner))
		daemonGroup.GET("/gc-repository/:namespace_id/runners/:runner_id/records/", server.WrapRequest(h.ListGcRepositoryRecords))
		daemonGroup.GET("/gc-repository/:namespace_id/runners/:runner_id/records/:record_id", server.WrapRequest(h.GetGcRepositoryRecord))

		daemonGroup.PUT("/gc-tag/:namespace_id/", server.WrapRequest(h.UpdateGcTagRule))
		daemonGroup.GET("/gc-tag/:namespace_id/", server.WrapRequest(h.GetGcTagRule))
		daemonGroup.GET("/gc-tag/:namespace_id/runners/latest", server.WrapRequest(h.GetGcTagLatestRunner))
		daemonGroup.POST("/gc-tag/:namespace_id/runners/", server.WrapRequest(h.CreateGcTagRunner))
		daemonGroup.GET("/gc-tag/:namespace_id/runners/", server.WrapRequest(h.ListGcTagRunners))
		daemonGroup.GET("/gc-tag/:namespace_id/runners/:runner_id", server.WrapRequest(h.GetGcTagRunner))
		daemonGroup.GET("/gc-tag/:namespace_id/runners/:runner_id/records/", server.WrapRequest(h.ListGcTagRecords))
		daemonGroup.GET("/gc-tag/:namespace_id/runners/:runner_id/records/:record_id", server.WrapRequest(h.GetGcTagRecord))

		daemonGroup.PUT("/gc-artifact/:namespace_id/", server.WrapRequest(h.UpdateGcArtifactRule))
		daemonGroup.GET("/gc-artifact/:namespace_id/", server.WrapRequest(h.GetGcArtifactRule))
		daemonGroup.GET("/gc-artifact/:namespace_id/runners/latest", server.WrapRequest(h.GetGcArtifactLatestRunner))
		daemonGroup.POST("/gc-artifact/:namespace_id/runners/", server.WrapRequest(h.CreateGcArtifactRunner))
		daemonGroup.GET("/gc-artifact/:namespace_id/runners/", server.WrapRequest(h.ListGcArtifactRunners))
		daemonGroup.GET("/gc-artifact/:namespace_id/runners/:runner_id", server.WrapRequest(h.GetGcArtifactRunner))
		daemonGroup.GET("/gc-artifact/:namespace_id/runners/:runner_id/records/", server.WrapRequest(h.ListGcArtifactRecords))
		daemonGroup.GET("/gc-artifact/:namespace_id/runners/:runner_id/records/:record_id", server.WrapRequest(h.GetGcArtifactRecord))

		daemonGroup.PUT("/gc-blob/:namespace_id/", server.WrapRequest(h.UpdateGcBlobRule))
		daemonGroup.GET("/gc-blob/:namespace_id/", server.WrapRequest(h.GetGcBlobRule))
		daemonGroup.GET("/gc-blob/:namespace_id/runners/latest", server.WrapRequest(h.GetGcBlobLatestRunner))
		daemonGroup.POST("/gc-blob/:namespace_id/runners/", server.WrapRequest(h.CreateGcBlobRunner))
		daemonGroup.GET("/gc-blob/:namespace_id/runners/", server.WrapRequest(h.ListGcBlobRunners))
		daemonGroup.GET("/gc-blob/:namespace_id/runners/:runner_id", server.WrapRequest(h.GetGcBlobRunner))
		daemonGroup.GET("/gc-blob/:namespace_id/runners/:runner_id/records/", server.WrapRequest(h.ListGcBlobRecords))
		daemonGroup.GET("/gc-blob/:namespace_id/runners/:runner_id/records/:record_id", server.WrapRequest(h.GetGcBlobRecord))

		return nil
	})
}

func init() {
	utils.PanicIf(handlers.Routers.Register(path.Base(reflect.TypeFor[factory]().PkgPath()), &factory{}))
}
