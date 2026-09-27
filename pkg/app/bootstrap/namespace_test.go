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

package bootstrap

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/dig"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/dal/models"
	reponamespace "github.com/go-sigma/sigma/pkg/dal/repository/namespace"
	reponamespacemocks "github.com/go-sigma/sigma/pkg/dal/repository/namespace/mocks"
	"github.com/go-sigma/sigma/pkg/logger"
)

func TestInitNamespaces(t *testing.T) {
	logger.SetLevel("debug")

	tests := []struct {
		name      string
		namespace config.ConfigurationNamespace
		genDigCon func(*testing.T, *config.Configuration) *dig.Container
		wantErr   bool
	}{
		{
			name:      "empty",
			namespace: config.ConfigurationNamespace{},
			genDigCon: func(t *testing.T, _ *config.Configuration) *dig.Container {
				ctrl := gomock.NewController(t)
				mock := reponamespacemocks.NewMockNamespaceRepository(ctrl)
				digCon := dig.New()
				require.NoError(t, digCon.Provide(func() reponamespace.NamespaceRepository { return mock }))
				return digCon
			},
		},
		{
			name: "ensure",
			namespace: config.ConfigurationNamespace{
				Initialize: []config.ConfigurationNamespaceInit{
					{Name: "library", Visibility: enums.VisibilityPublic},
					{Name: "internal", Visibility: enums.VisibilityPrivate},
				},
			},
			genDigCon: func(t *testing.T, _ *config.Configuration) *dig.Container {
				ctrl := gomock.NewController(t)
				mock := reponamespacemocks.NewMockNamespaceRepository(ctrl)
				mock.EXPECT().GetByName(gomock.Any(), "library").
					Return(&models.Namespace{ID: "namespace-1", Name: "library"}, nil)
				mock.EXPECT().GetByName(gomock.Any(), "internal").
					Return(nil, gorm.ErrRecordNotFound)
				mock.EXPECT().Create(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, namespace *models.Namespace) error {
						require.Equal(t, "internal", namespace.Name)
						require.Equal(t, enums.VisibilityPrivate, namespace.Visibility)
						require.NotEmpty(t, namespace.ID)
						return nil
					})
				digCon := dig.New()
				require.NoError(t, digCon.Provide(func() reponamespace.NamespaceRepository { return mock }))
				return digCon
			},
		},
		{
			name: "get error",
			namespace: config.ConfigurationNamespace{
				Initialize: []config.ConfigurationNamespaceInit{
					{Name: "library", Visibility: enums.VisibilityPublic},
				},
			},
			genDigCon: func(t *testing.T, _ *config.Configuration) *dig.Container {
				ctrl := gomock.NewController(t)
				mock := reponamespacemocks.NewMockNamespaceRepository(ctrl)
				mock.EXPECT().GetByName(gomock.Any(), "library").
					Return(nil, gorm.ErrInvalidDB)
				digCon := dig.New()
				require.NoError(t, digCon.Provide(func() reponamespace.NamespaceRepository { return mock }))
				return digCon
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Configuration{Namespace: tt.namespace}
			digCon := tt.genDigCon(t, cfg)
			require.NoError(t, digCon.Provide(func() *config.Configuration { return cfg }))
			err := initNamespaces(digCon)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
