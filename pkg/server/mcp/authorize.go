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

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/dal/models"
)

func (s *Server) authorizeNamespace(ctx context.Context, user *models.User, namespaceID string, auth enums.Auth) error {
	allowed, err := s.authorizer.Namespace(ctx, *user, namespaceID, auth)
	if err != nil {
		return err
	}
	if !allowed {
		return errForbidden
	}
	return nil
}

func (s *Server) authorizeRepository(ctx context.Context, user *models.User, repositoryID string, auth enums.Auth) error {
	allowed, err := s.authorizer.Repository(ctx, *user, repositoryID, auth)
	if err != nil {
		return err
	}
	if !allowed {
		return errForbidden
	}
	return nil
}

func (s *Server) authorizeTag(ctx context.Context, user *models.User, tagID string, auth enums.Auth) error {
	allowed, err := s.authorizer.Tag(ctx, *user, tagID, auth)
	if err != nil {
		return err
	}
	if !allowed {
		return errForbidden
	}
	return nil
}
