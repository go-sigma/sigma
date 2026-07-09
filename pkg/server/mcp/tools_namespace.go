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

func (s *Server) registerNamespaceTools(mcpServer *mcpsdk.MCPServer) {
	s.addTool(mcpServer, "namespace-list", "List namespaces visible to the authenticated user", false, s.namespaceList)
	s.addTool(mcpServer, "namespace-get", "Get a namespace by id", false, s.namespaceGet)
	s.addTool(mcpServer, "namespace-create", "Create a namespace", true, s.namespaceCreate)
	s.addTool(mcpServer, "namespace-update", "Update a namespace", true, s.namespaceUpdate)
	s.addTool(mcpServer, "namespace-delete", "Delete a namespace", true, s.namespaceDelete)
	s.addTool(mcpServer, "namespace-member-list", "List namespace members", false, s.namespaceMemberList)
	s.addTool(mcpServer, "namespace-member-add", "Add a namespace member", true, s.namespaceMemberAdd)
	s.addTool(mcpServer, "namespace-member-update", "Update a namespace member role", true, s.namespaceMemberUpdate)
	s.addTool(mcpServer, "namespace-member-delete", "Delete a namespace member", true, s.namespaceMemberDelete)
	s.addTool(mcpServer, "namespace-member-self-get", "Get authenticated user's namespace membership", false, s.namespaceMemberSelfGet)
}

func (s *Server) namespaceList(ctx context.Context, req mcp.CallToolRequest) (any, error) {
	user, _ := userFromContext(ctx)
	args, err := bindArguments[api.ListNamespaceRequest](req)
	if err != nil {
		return nil, err
	}
	items, total, err := s.namespaceSvc.ListNamespaces(ctx, user.ID, args.Name, args.Pagination, args.Sortable)
	if err != nil {
		return nil, err
	}
	return listResponse(items, total, args.Pagination), nil
}

func (s *Server) namespaceGet(ctx context.Context, req mcp.CallToolRequest) (any, error) {
	user, _ := userFromContext(ctx)
	args, err := bindArguments[api.GetNamespaceRequest](req)
	if err != nil {
		return nil, err
	}
	if e := s.authorizeNamespace(ctx, user, args.ID, enums.AuthRead); e != nil {
		return nil, e
	}
	namespaceObj, repositoryCount, tagCount, err := s.namespaceSvc.GetNamespace(ctx, args.ID)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"namespace":        namespaceObj,
		"repository_count": repositoryCount,
		"tag_count":        tagCount,
	}, nil
}

func (s *Server) namespaceCreate(ctx context.Context, req mcp.CallToolRequest) (any, error) {
	user, _ := userFromContext(ctx)
	args, err := bindArguments[api.PostNamespaceRequest](req)
	if err != nil {
		return nil, err
	}
	namespaceObj, err := s.namespaceSvc.CreateNamespace(ctx, user.ID, args)
	if err != nil {
		return nil, err
	}
	return map[string]string{"id": namespaceObj.ID}, nil
}

func (s *Server) namespaceUpdate(ctx context.Context, req mcp.CallToolRequest) (any, error) {
	user, _ := userFromContext(ctx)
	args, err := bindArguments[api.UpdateNamespaceRequest](req)
	if err != nil {
		return nil, err
	}
	if e := s.authorizeNamespace(ctx, user, args.ID, enums.AuthAdmin); e != nil {
		return nil, e
	}
	if e := s.namespaceSvc.UpdateNamespace(ctx, user.ID, args.ID, args); e != nil {
		return nil, e
	}
	return okResponse(), nil
}

func (s *Server) namespaceDelete(ctx context.Context, req mcp.CallToolRequest) (any, error) {
	user, _ := userFromContext(ctx)
	args, err := bindArguments[api.DeleteNamespaceRequest](req)
	if err != nil {
		return nil, err
	}
	if e := s.authorizeNamespace(ctx, user, args.ID, enums.AuthAdmin); e != nil {
		return nil, e
	}
	if e := s.namespaceSvc.DeleteNamespace(ctx, user.ID, args.ID); e != nil {
		return nil, e
	}
	return okResponse(), nil
}

func (s *Server) namespaceMemberList(ctx context.Context, req mcp.CallToolRequest) (any, error) {
	user, _ := userFromContext(ctx)
	args, err := bindArguments[api.ListNamespaceMemberRequest](req)
	if err != nil {
		return nil, err
	}
	if e := s.authorizeNamespace(ctx, user, args.NamespaceID, enums.AuthRead); e != nil {
		return nil, e
	}
	members, total, err := s.namespaceSvc.ListNamespaceMembers(ctx, args.NamespaceID, args.Name, args.Pagination, args.Sortable)
	if err != nil {
		return nil, err
	}
	return listResponse(members, total, args.Pagination), nil
}

func (s *Server) namespaceMemberAdd(ctx context.Context, req mcp.CallToolRequest) (any, error) {
	user, _ := userFromContext(ctx)
	args, err := bindArguments[api.AddNamespaceMemberRequest](req)
	if err != nil {
		return nil, err
	}
	if e := s.authorizeNamespace(ctx, user, args.NamespaceID, enums.AuthAdmin); e != nil {
		return nil, e
	}
	member, err := s.namespaceSvc.AddNamespaceMember(ctx, user.ID, args.NamespaceID, args.UserID, args.Role)
	if err != nil {
		return nil, err
	}
	return map[string]string{"id": member.ID}, nil
}

func (s *Server) namespaceMemberUpdate(ctx context.Context, req mcp.CallToolRequest) (any, error) {
	user, _ := userFromContext(ctx)
	args, err := bindArguments[api.UpdateNamespaceMemberRequest](req)
	if err != nil {
		return nil, err
	}
	if e := s.authorizeNamespace(ctx, user, args.NamespaceID, enums.AuthAdmin); e != nil {
		return nil, e
	}
	if e := s.namespaceSvc.UpdateNamespaceMember(ctx, user.ID, args.NamespaceID, args.UserID, args.Role); e != nil {
		return nil, e
	}
	return okResponse(), nil
}

func (s *Server) namespaceMemberDelete(ctx context.Context, req mcp.CallToolRequest) (any, error) {
	user, _ := userFromContext(ctx)
	args, err := bindArguments[api.DeleteNamespaceMemberRequest](req)
	if err != nil {
		return nil, err
	}
	if e := s.authorizeNamespace(ctx, user, args.NamespaceID, enums.AuthAdmin); e != nil {
		return nil, e
	}
	if e := s.namespaceSvc.DeleteNamespaceMember(ctx, user.ID, args.NamespaceID, args.UserID); e != nil {
		return nil, e
	}
	return okResponse(), nil
}

func (s *Server) namespaceMemberSelfGet(ctx context.Context, req mcp.CallToolRequest) (any, error) {
	user, _ := userFromContext(ctx)
	args, err := bindArguments[api.GetNamespaceMemberSelfRequest](req)
	if err != nil {
		return nil, err
	}
	member, err := s.namespaceSvc.GetNamespaceMember(ctx, args.NamespaceID, user.ID)
	if err != nil {
		return nil, err
	}
	return member, nil
}
