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
	"github.com/gin-gonic/gin"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/server"
	"github.com/go-sigma/sigma/pkg/service/daemons"
)

// Handler is the interface for the gc handlers
type Handler interface {
	// UpdateGcTagRule replaces the namespace's gc tag retention rule (cron schedule and retention policy).
	UpdateGcTagRule(c *gin.Context, req *api.UpdateGcTagRuleRequest)
	// GetGcTagRule returns the namespace's gc tag retention rule.
	GetGcTagRule(c *gin.Context, req *api.GetGcTagRuleRequest)
	// GetGcTagLatestRunner returns the most recent gc tag run of the namespace.
	GetGcTagLatestRunner(c *gin.Context, req *api.GetGcTagLatestRunnerRequest)
	// CreateGcTagRunner triggers an immediate gc tag run for the namespace.
	CreateGcTagRunner(c *gin.Context, req *api.CreateGcTagRunnerRequest)
	// ListGcTagRunners returns a paginated list of the namespace's gc tag runs.
	ListGcTagRunners(c *gin.Context, req *api.ListGcTagRunnersRequest)
	// GetGcTagRunner returns a single gc tag run by its runner id.
	GetGcTagRunner(c *gin.Context, req *api.GetGcTagRunnerRequest)
	// ListGcTagRecords returns a paginated list of the per-tag outcomes recorded by one gc tag run.
	ListGcTagRecords(c *gin.Context, req *api.ListGcTagRecordsRequest)
	// GetGcTagRecord returns a single per-tag outcome recorded by a gc tag run.
	GetGcTagRecord(c *gin.Context, req *api.GetGcTagRecordRequest)

	// UpdateGcRepositoryRule replaces the namespace's gc repository retention rule (cron schedule and retention policy).
	UpdateGcRepositoryRule(c *gin.Context, req *api.UpdateGcRepositoryRuleRequest)
	// GetGcRepositoryRule returns the namespace's gc repository retention rule.
	GetGcRepositoryRule(c *gin.Context, req *api.GetGcRepositoryRuleRequest)
	// GetGcRepositoryLatestRunner returns the most recent gc repository run of the namespace.
	GetGcRepositoryLatestRunner(c *gin.Context, req *api.GetGcRepositoryLatestRunnerRequest)
	// CreateGcRepositoryRunner triggers an immediate gc repository run for the namespace.
	CreateGcRepositoryRunner(c *gin.Context, req *api.CreateGcRepositoryRunnerRequest)
	// ListGcRepositoryRunners returns a paginated list of the namespace's gc repository runs.
	ListGcRepositoryRunners(c *gin.Context, req *api.ListGcRepositoryRunnersRequest)
	// GetGcRepositoryRunner returns a single gc repository run by its runner id.
	GetGcRepositoryRunner(c *gin.Context, req *api.GetGcRepositoryRunnerRequest)
	// ListGcRepositoryRecords returns a paginated list of the per-repository outcomes recorded by one gc repository run.
	ListGcRepositoryRecords(c *gin.Context, req *api.ListGcRepositoryRecordsRequest)
	// GetGcRepositoryRecord returns a single per-repository outcome recorded by a gc repository run.
	GetGcRepositoryRecord(c *gin.Context, req *api.GetGcRepositoryRecordRequest)

	// UpdateGcArtifactRule replaces the namespace's gc artifact retention rule (cron schedule and retention policy).
	UpdateGcArtifactRule(c *gin.Context, req *api.UpdateGcArtifactRuleRequest)
	// GetGcArtifactRule returns the namespace's gc artifact retention rule.
	GetGcArtifactRule(c *gin.Context, req *api.GetGcArtifactRuleRequest)
	// GetGcArtifactLatestRunner returns the most recent gc artifact run of the namespace.
	GetGcArtifactLatestRunner(c *gin.Context, req *api.GetGcArtifactLatestRunnerRequest)
	// CreateGcArtifactRunner triggers an immediate gc artifact run for the namespace.
	CreateGcArtifactRunner(c *gin.Context, req *api.CreateGcArtifactRunnerRequest)
	// ListGcArtifactRunners returns a paginated list of the namespace's gc artifact runs.
	ListGcArtifactRunners(c *gin.Context, req *api.ListGcArtifactRunnersRequest)
	// GetGcArtifactRunner returns a single gc artifact run by its runner id.
	GetGcArtifactRunner(c *gin.Context, req *api.GetGcArtifactRunnerRequest)
	// ListGcArtifactRecords returns a paginated list of the per-artifact outcomes recorded by one gc artifact run.
	ListGcArtifactRecords(c *gin.Context, req *api.ListGcArtifactRecordsRequest)
	// GetGcArtifactRecord returns a single per-artifact outcome recorded by a gc artifact run.
	GetGcArtifactRecord(c *gin.Context, req *api.GetGcArtifactRecordRequest)

	// UpdateGcBlobRule replaces the namespace's gc blob retention rule (cron schedule and retention policy).
	UpdateGcBlobRule(c *gin.Context, req *api.UpdateGcBlobRuleRequest)
	// GetGcBlobRule returns the namespace's gc blob retention rule.
	GetGcBlobRule(c *gin.Context, req *api.GetGcBlobRuleRequest)
	// GetGcBlobLatestRunner returns the most recent gc blob run of the namespace.
	GetGcBlobLatestRunner(c *gin.Context, req *api.GetGcBlobLatestRunnerRequest)
	// CreateGcBlobRunner triggers an immediate gc blob run for the namespace.
	CreateGcBlobRunner(c *gin.Context, req *api.CreateGcBlobRunnerRequest)
	// ListGcBlobRunners returns a paginated list of the namespace's gc blob runs.
	ListGcBlobRunners(c *gin.Context, req *api.ListGcBlobRunnersRequest)
	// GetGcBlobRunner returns a single gc blob run by its runner id.
	GetGcBlobRunner(c *gin.Context, req *api.GetGcBlobRunnerRequest)
	// ListGcBlobRecords returns a paginated list of the per-blob outcomes recorded by one gc blob run.
	ListGcBlobRecords(c *gin.Context, req *api.ListGcBlobRecordsRequest)
	// GetGcBlobRecord returns a single per-blob outcome recorded by a gc blob run.
	GetGcBlobRecord(c *gin.Context, req *api.GetGcBlobRecordRequest)
}

var _ Handler = &handler{}

type handler struct {
	dig.In

	DaemonSvc daemons.Service
}

// Initialize registers the handler routes.
func Initialize(e *gin.Engine, h handler) error {
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
}
