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

package systems

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/version"
)

func TestService(t *testing.T) {
	configuration := &config.Configuration{
		HTTP: config.ConfigurationHTTP{Endpoint: "https://sigma.example.test"},
	}
	service := &service{Config: configuration}

	gotConfig, err := service.GetConfig(t.Context())
	require.NoError(t, err)
	require.Equal(t, *configuration, gotConfig)

	endpoint, err := service.GetEndpoint(t.Context())
	require.NoError(t, err)
	require.Equal(t, configuration.HTTP.Endpoint, endpoint)

	gotVersion, err := service.GetVersion(t.Context())
	require.NoError(t, err)
	require.Equal(t, version.Version, gotVersion.Version)
	require.Equal(t, version.GitHash, gotVersion.GitHash)
	require.Equal(t, version.BuildDate, gotVersion.BuildDate)
}
