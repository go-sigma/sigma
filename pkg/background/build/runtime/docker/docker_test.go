// Copyright 2023 sigma
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

package docker

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDockerHost(t *testing.T) {
	require.Equal(t, "unix:///var/run/docker.sock", dockerHost("/var/run/docker.sock"))
	require.Equal(t, "unix:///var/run/docker.sock", dockerHost("unix:///var/run/docker.sock"))
	require.Equal(t, "unix:///Users/sigma/.orbstack/run/docker.sock", dockerHost("/Users/sigma/.orbstack/run/docker.sock"))
	require.Equal(t, "tcp://127.0.0.1:2375", dockerHost("tcp://127.0.0.1:2375"))
	require.Equal(t, "npipe:////./pipe/docker_engine", dockerHost("npipe:////./pipe/docker_engine"))
}
