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

package mcpserver

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	mcpsdk "github.com/mark3labs/mcp-go/server"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/authz"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	repoaudit "github.com/go-sigma/sigma/pkg/dal/repository/audit"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	svcartifacts "github.com/go-sigma/sigma/pkg/service/artifacts"
	svcnamespaces "github.com/go-sigma/sigma/pkg/service/namespaces"
	svcrepositories "github.com/go-sigma/sigma/pkg/service/repositories"
	svcsystems "github.com/go-sigma/sigma/pkg/service/systems"
	svctags "github.com/go-sigma/sigma/pkg/service/tags"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/version"
)

// Params declares dependencies needed to register the MCP server.
type Params struct {
	dig.In

	Config          *config.Configuration
	Engine          *gin.Engine
	AuditRepository repoaudit.AuditRepository
	Authorizer      authz.Authorizer
	SystemsSvc      svcsystems.SystemsService
	NamespaceSvc    svcnamespaces.NamespaceService
	RepositorySvc   svcrepositories.RepositoryService
	TagSvc          svctags.TagService
	ArtifactSvc     svcartifacts.ArtifactService
}

type Server struct {
	config          *config.Configuration
	auditRepository repoaudit.AuditRepository
	authorizer      authz.Authorizer
	systemsSvc      svcsystems.SystemsService
	namespaceSvc    svcnamespaces.NamespaceService
	repositorySvc   svcrepositories.RepositoryService
	tagSvc          svctags.TagService
	artifactSvc     svcartifacts.ArtifactService
}

// Register registers the embedded MCP HTTP endpoint when it is enabled.
func Register(digCon *dig.Container) error {
	return digCon.Invoke(func(params Params) error {
		if !params.Config.MCP.Enabled {
			return nil
		}
		serverObj := &Server{
			config:          params.Config,
			auditRepository: params.AuditRepository,
			authorizer:      params.Authorizer,
			systemsSvc:      params.SystemsSvc,
			namespaceSvc:    params.NamespaceSvc,
			repositorySvc:   params.RepositorySvc,
			tagSvc:          params.TagSvc,
			artifactSvc:     params.ArtifactSvc,
		}
		handler := serverObj.newHTTPHandler()
		params.Engine.Any(params.Config.MCP.Path, func(c *gin.Context) {
			user, ok := utils.GetFromCtx[*models.User](c, consts.ContextUser)
			if !ok || user == nil || user.Role == enums.UserRoleAnonymous {
				recordAuthFailure("missing_user")
				errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized)
				return
			}
			if !strings.HasPrefix(c.GetHeader(consts.HeaderAuthorization), "Basic ") {
				recordAuthFailure("non_basic_auth")
				c.Header(consts.HeaderWWWAuthenticate, `Basic realm="sigma-mcp"`)
				errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized, "MCP requires Basic Auth")
				return
			}
			if params.Config.MCP.MaxRequestBody > 0 && c.Request.Body != nil {
				c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, params.Config.MCP.MaxRequestBody)
			}
			ctx := withUser(c.Request.Context(), user)
			ctx = withClientInfo(ctx, c.ClientIP(), c.Request.UserAgent())
			handler.ServeHTTP(c.Writer, c.Request.WithContext(ctx))
		})
		slog.Info("mcp server registered", "path", params.Config.MCP.Path)
		return nil
	})
}

func (s *Server) newHTTPHandler() http.Handler {
	mcpServer := mcpsdk.NewMCPServer(
		consts.AppName,
		version.Version,
		mcpsdk.WithToolCapabilities(true),
		mcpsdk.WithResourceCapabilities(true, false),
		mcpsdk.WithInstructions("Sigma MCP exposes registry management tools. Every request must use Sigma Basic Auth."),
	)
	s.registerSystemTools(mcpServer)
	s.registerNamespaceTools(mcpServer)
	s.registerRepositoryTools(mcpServer)
	s.registerTagTools(mcpServer)
	s.registerArtifactTools(mcpServer)
	s.registerResources(mcpServer)
	return mcpsdk.NewStreamableHTTPServer(
		mcpServer,
		mcpsdk.WithStateLess(true),
		mcpsdk.WithHTTPContextFunc(func(ctx context.Context, r *http.Request) context.Context {
			if user, ok := userFromContext(r.Context()); ok {
				ctx = withUser(ctx, user)
			}
			clientIP, userAgent := clientInfoFromContext(r.Context())
			return withClientInfo(ctx, clientIP, userAgent)
		}),
	)
}
