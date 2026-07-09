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

type tagManifestRawGetRequest struct {
	RepositoryID string `json:"repository_id"`
	Digest       string `json:"digest"`
}

func (s *Server) registerTagTools(mcpServer *mcpsdk.MCPServer) {
	s.addTool(mcpServer, "tag-list", "List repository tags", false, s.tagList)
	s.addTool(mcpServer, "tag-get", "Get a tag by id", false, s.tagGet)
	s.addTool(mcpServer, "tag-delete", "Delete a tag by id", true, s.tagDelete)
	s.addTool(mcpServer, "tag-manifest-raw-get", "Get raw manifest content by artifact digest", false, s.tagManifestRawGet)
}

func (s *Server) tagList(ctx context.Context, req mcp.CallToolRequest) (any, error) {
	user, _ := userFromContext(ctx)
	args, err := bindArguments[api.ListTagRequest](req)
	if err != nil {
		return nil, err
	}
	if e := s.authorizeRepository(ctx, user, args.RepositoryID, enums.AuthRead); e != nil {
		return nil, e
	}
	tags, total, err := s.tagSvc.ListTags(ctx, args.NamespaceID, args.RepositoryID, args.Name, args.Type, args.Pagination, args.Sortable)
	if err != nil {
		return nil, err
	}
	return listResponse(tags, total, args.Pagination), nil
}

func (s *Server) tagGet(ctx context.Context, req mcp.CallToolRequest) (any, error) {
	user, _ := userFromContext(ctx)
	args, err := bindArguments[api.GetTagRequest](req)
	if err != nil {
		return nil, err
	}
	if e := s.authorizeTag(ctx, user, args.ID, enums.AuthRead); e != nil {
		return nil, e
	}
	return s.tagSvc.GetTag(ctx, args.ID)
}

func (s *Server) tagDelete(ctx context.Context, req mcp.CallToolRequest) (any, error) {
	user, _ := userFromContext(ctx)
	args, err := bindArguments[api.DeleteTagRequest](req)
	if err != nil {
		return nil, err
	}
	if e := s.authorizeTag(ctx, user, args.ID, enums.AuthManage); e != nil {
		return nil, e
	}
	if e := s.tagSvc.DeleteTag(ctx, args.NamespaceID, args.RepositoryID, args.ID); e != nil {
		return nil, e
	}
	return okResponse(), nil
}

func (s *Server) tagManifestRawGet(ctx context.Context, req mcp.CallToolRequest) (any, error) {
	user, _ := userFromContext(ctx)
	args, err := bindArguments[tagManifestRawGetRequest](req)
	if err != nil {
		return nil, err
	}
	if e := s.authorizeRepository(ctx, user, args.RepositoryID, enums.AuthRead); e != nil {
		return nil, e
	}
	raw, err := s.tagSvc.GetArtifactRaw(ctx, args.Digest)
	if err != nil {
		return nil, err
	}
	return map[string]string{"raw": string(raw)}, nil
}
