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

package builder

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCacheRef(t *testing.T) {
	flow := &BuildFlow{
		Config{
			Endpoint:   "http://host.docker.internal:3000",
			Repository: base64.StdEncoding.EncodeToString([]byte("library/rrr")),
		},
	}

	ref, err := flow.cacheRef()
	require.NoError(t, err)
	require.Equal(t, "host.docker.internal:3000/library/rrr:buildcache", ref)
}

func TestCacheRefInvalidRepository(t *testing.T) {
	flow := &BuildFlow{
		Config{
			Endpoint:   "http://127.0.0.1:3000",
			Repository: "not-base64",
		},
	}

	_, err := flow.cacheRef()
	require.Error(t, err)
}
