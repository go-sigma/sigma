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

	"github.com/mark3labs/mcp-go/mcp"
	mcpsdk "github.com/mark3labs/mcp-go/server"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
)

func (s *Server) registerArtifactTools(mcpServer *mcpsdk.MCPServer) {
	s.addTool(mcpServer, "artifact-list", "List repository artifacts", false, s.artifactList)
	s.addTool(mcpServer, "artifact-get", "Get an artifact by repository name and digest", false, s.artifactGet)
	s.addTool(mcpServer, "artifact-delete", "Delete an artifact by repository name and digest", true, s.artifactDelete)
}

func (s *Server) artifactList(ctx context.Context, req mcp.CallToolRequest) (any, error) {
	user, _ := userFromContext(ctx)
	args, err := bindArguments[api.ListArtifactRequest](req)
	if err != nil {
		return nil, err
	}
	repositoryObj, err := s.repositorySvc.GetRepositoryByName(ctx, args.Repository)
	if err != nil {
		return nil, err
	}
	if e := s.authorizeRepository(ctx, user, repositoryObj.ID, enums.AuthRead); e != nil {
		return nil, e
	}
	artifacts, total, err := s.artifactSvc.ListArtifacts(ctx, args)
	if err != nil {
		return nil, err
	}
	return listResponse(artifacts, total, args.Pagination), nil
}

func (s *Server) artifactGet(ctx context.Context, req mcp.CallToolRequest) (any, error) {
	user, _ := userFromContext(ctx)
	args, err := bindArguments[api.GetArtifactRequest](req)
	if err != nil {
		return nil, err
	}
	repositoryObj, err := s.repositorySvc.GetRepositoryByName(ctx, args.Repository)
	if err != nil {
		return nil, err
	}
	if e := s.authorizeRepository(ctx, user, repositoryObj.ID, enums.AuthRead); e != nil {
		return nil, e
	}
	return s.artifactSvc.GetArtifact(ctx, args.Repository, args.Digest)
}

func (s *Server) artifactDelete(ctx context.Context, req mcp.CallToolRequest) (any, error) {
	user, _ := userFromContext(ctx)
	args, err := bindArguments[api.DeleteArtifactRequest](req)
	if err != nil {
		return nil, err
	}
	repositoryObj, err := s.repositorySvc.GetRepositoryByName(ctx, args.Repository)
	if err != nil {
		return nil, err
	}
	if e := s.authorizeRepository(ctx, user, repositoryObj.ID, enums.AuthManage); e != nil {
		return nil, e
	}
	if e := s.artifactSvc.DeleteArtifact(ctx, args.Repository, args.Digest); e != nil {
		return nil, e
	}
	return okResponse(), nil
}
