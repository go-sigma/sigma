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

package oauth2

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/config"
)

func TestGetClientID(t *testing.T) {
	service := &service{Config: &config.Configuration{
		Auth: config.ConfigurationAuth{
			Oauth2: config.ConfigurationAuthOauth2{
				Github: config.ConfigurationAuthOauth2Github{ClientID: "github-client"},
				Gitlab: config.ConfigurationAuthOauth2Gitlab{ClientID: "gitlab-client"},
				Gitea:  config.ConfigurationAuthOauth2Gitea{ClientID: "gitea-client"},
			},
		},
	}}

	tests := []struct {
		name     string
		provider enums.Provider
		want     string
		wantErr  bool
	}{
		{name: "github", provider: enums.ProviderGithub, want: "github-client"},
		{name: "gitlab", provider: enums.ProviderGitlab, want: "gitlab-client"},
		{name: "gitea", provider: enums.ProviderGitea, want: "gitea-client"},
		{name: "invalid", provider: enums.Provider("invalid"), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := service.GetClientID(t.Context(), tt.provider)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
