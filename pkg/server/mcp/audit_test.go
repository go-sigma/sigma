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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRedactSensitive(t *testing.T) {
	input := map[string]any{
		"username":    "sigma",
		"password":    "secret",
		"scm_token":   "token",
		"private_key": "key",
		"nested": map[string]any{
			"authorization": "Basic value",
			"value":         "kept",
		},
		"items": []any{
			map[string]any{"sk": "secret-key"},
		},
	}

	redacted, ok := redactSensitive(input).(map[string]any)
	require.True(t, ok)
	require.Equal(t, "sigma", redacted["username"])
	require.Equal(t, "[REDACTED]", redacted["password"])
	require.Equal(t, "[REDACTED]", redacted["scm_token"])
	require.Equal(t, "[REDACTED]", redacted["private_key"])

	nested, ok := redacted["nested"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "[REDACTED]", nested["authorization"])
	require.Equal(t, "kept", nested["value"])

	items, ok := redacted["items"].([]any)
	require.True(t, ok)
	item, ok := items[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "[REDACTED]", item["sk"])
}
