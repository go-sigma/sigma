// Copyright 2024 sigma
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

package redis

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/require"

	"github.com/go-sigma/sigma/pkg/config"
)

func TestRedis(t *testing.T) {
	// No connection info → disabled, returns nil client
	client, err := New(&config.Configuration{
		Redis: config.ConfigurationRedis{},
	})
	require.NoError(t, err)
	require.Nil(t, client)

	// Invalid URL (missing scheme) → error
	client, err = New(&config.Configuration{
		Redis: config.ConfigurationRedis{
			URL: miniredis.RunT(t).Addr(),
		},
	})
	require.Error(t, err)
	require.Nil(t, client)

	// Valid URL → client created
	client, err = New(&config.Configuration{
		Redis: config.ConfigurationRedis{
			URL: "redis://" + miniredis.RunT(t).Addr(),
		},
	})
	require.NoError(t, err)
	require.NotNil(t, client)
}
