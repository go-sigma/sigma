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

func (s *Server) registerRepositoryTools(mcpServer *mcpsdk.MCPServer) {
	s.addTool(mcpServer, "repository-list", "List repositories in a namespace", false, s.repositoryList)
	s.addTool(mcpServer, "repository-get", "Get a repository by id", false, s.repositoryGet)
	s.addTool(mcpServer, "repository-create", "Create a repository", true, s.repositoryCreate)
	s.addTool(mcpServer, "repository-update", "Update a repository", true, s.repositoryUpdate)
	s.addTool(mcpServer, "repository-delete", "Delete a repository", true, s.repositoryDelete)
}

func (s *Server) repositoryList(ctx context.Context, req mcp.CallToolRequest) (any, error) {
	user, _ := userFromContext(ctx)
	args, err := bindArguments[api.ListRepositoryRequest](req)
	if err != nil {
		return nil, err
	}
	if e := s.authorizeNamespace(ctx, user, args.NamespaceID, enums.AuthRead); e != nil {
		return nil, e
	}
	repositories, builders, total, err := s.repositorySvc.ListRepositories(
		ctx, user.ID, args.NamespaceID, args.Name, args.Pagination, args.Sortable,
	)
	if err != nil {
		return nil, err
	}
	return listResponse(map[string]any{
		"repositories": repositories,
		"builders":     builders,
	}, total, args.Pagination), nil
}

func (s *Server) repositoryGet(ctx context.Context, req mcp.CallToolRequest) (any, error) {
	user, _ := userFromContext(ctx)
	args, err := bindArguments[api.GetRepositoryRequest](req)
	if err != nil {
		return nil, err
	}
	if e := s.authorizeRepository(ctx, user, args.ID, enums.AuthRead); e != nil {
		return nil, e
	}
	return s.repositorySvc.GetRepository(ctx, args.ID)
}

func (s *Server) repositoryCreate(ctx context.Context, req mcp.CallToolRequest) (any, error) {
	user, _ := userFromContext(ctx)
	args, err := bindArguments[api.CreateRepositoryRequest](req)
	if err != nil {
		return nil, err
	}
	if e := s.authorizeNamespace(ctx, user, args.NamespaceID, enums.AuthManage); e != nil {
		return nil, e
	}
	repository, err := s.repositorySvc.CreateRepository(ctx, user.ID, args)
	if err != nil {
		return nil, err
	}
	return map[string]string{"id": repository.ID}, nil
}

func (s *Server) repositoryUpdate(ctx context.Context, req mcp.CallToolRequest) (any, error) {
	user, _ := userFromContext(ctx)
	args, err := bindArguments[api.UpdateRepositoryRequest](req)
	if err != nil {
		return nil, err
	}
	if e := s.authorizeRepository(ctx, user, args.ID, enums.AuthManage); e != nil {
		return nil, e
	}
	if e := s.repositorySvc.UpdateRepository(ctx, user.ID, args); e != nil {
		return nil, e
	}
	return okResponse(), nil
}

func (s *Server) repositoryDelete(ctx context.Context, req mcp.CallToolRequest) (any, error) {
	user, _ := userFromContext(ctx)
	args, err := bindArguments[api.DeleteRepositoryRequest](req)
	if err != nil {
		return nil, err
	}
	if e := s.authorizeRepository(ctx, user, args.ID, enums.AuthManage); e != nil {
		return nil, e
	}
	if e := s.repositorySvc.DeleteRepository(ctx, user.ID, args.NamespaceID, args.ID); e != nil {
		return nil, e
	}
	return okResponse(), nil
}
