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
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/authz"
	"github.com/go-sigma/sigma/pkg/dal/models"
)

func TestAuthorizeHelpers(t *testing.T) {
	expectedErr := errors.New("lookup failed")
	user := &models.User{ID: "user-1"}
	tests := []struct {
		name       string
		authorizer authz.Authorizer
		invoke     func(*Server) error
		wantErr    error
	}{
		{
			name:       "namespace allowed",
			authorizer: fakeAuthorizer{namespaceAllowed: true},
			invoke: func(server *Server) error {
				return server.authorizeNamespace(t.Context(), user, "namespace-1", enums.AuthRead)
			},
		},
		{
			name:       "namespace denied",
			authorizer: fakeAuthorizer{},
			invoke: func(server *Server) error {
				return server.authorizeNamespace(t.Context(), user, "namespace-1", enums.AuthRead)
			},
			wantErr: errForbidden,
		},
		{
			name:       "repository error",
			authorizer: fakeAuthorizer{repositoryErr: expectedErr},
			invoke: func(server *Server) error {
				return server.authorizeRepository(t.Context(), user, "repository-1", enums.AuthRead)
			},
			wantErr: expectedErr,
		},
		{
			name:       "repository allowed",
			authorizer: fakeAuthorizer{repositoryAllowed: true},
			invoke: func(server *Server) error {
				return server.authorizeRepository(t.Context(), user, "repository-1", enums.AuthManage)
			},
		},
		{
			name:       "tag denied",
			authorizer: fakeAuthorizer{},
			invoke:     func(server *Server) error { return server.authorizeTag(t.Context(), user, "tag-1", enums.AuthRead) },
			wantErr:    errForbidden,
		},
		{
			name:       "tag allowed",
			authorizer: fakeAuthorizer{tagAllowed: true},
			invoke:     func(server *Server) error { return server.authorizeTag(t.Context(), user, "tag-1", enums.AuthAdmin) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.invoke(&Server{authorizer: tt.authorizer})
			if tt.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tt.wantErr)
			}
		})
	}
}

type fakeAuthorizer struct {
	namespaceAllowed  bool
	repositoryAllowed bool
	tagAllowed        bool
	repositoryErr     error
}

var _ authz.Authorizer = fakeAuthorizer{}

func (fakeAuthorizer) Authorize(context.Context, string, bool, string, string) (bool, error) {
	return false, nil
}

func (f fakeAuthorizer) Namespace(context.Context, models.User, string, enums.Auth) (bool, error) {
	return f.namespaceAllowed, nil
}

func (fakeAuthorizer) NamespaceRole(context.Context, models.User, string) (*enums.NamespaceRole, error) {
	return nil, nil
}

func (fakeAuthorizer) NamespacesRole(context.Context, models.User, []string) (map[string]*enums.NamespaceRole, error) {
	return nil, nil
}

func (f fakeAuthorizer) Repository(context.Context, models.User, string, enums.Auth) (bool, error) {
	return f.repositoryAllowed, f.repositoryErr
}

func (f fakeAuthorizer) Tag(context.Context, models.User, string, enums.Auth) (bool, error) {
	return f.tagAllowed, nil
}

func (fakeAuthorizer) Artifact(context.Context, models.User, string, enums.Auth) (bool, error) {
	return false, nil
}
