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

package coderepos

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/server"
	"github.com/go-sigma/sigma/pkg/server/validators"
	"github.com/go-sigma/sigma/pkg/testkit"
)

// TestBindUpdateCredentialRequest guards the request binding of the credential
// endpoints: a credential body only fills one of the secret fields, and validator
// panics with "Bad field type *string" when max is combined with required_if on a
// pointer field.
func TestBindUpdateCredentialRequest(t *testing.T) {
	require.NoError(t, validators.Initialize())

	tests := []struct {
		name string
		body string
	}{
		{name: "token", body: `{"type":"token","token":"ghp_sigma"}`},
		{name: "ssh", body: `{"type":"ssh","ssh_key":"-----BEGIN OPENSSH PRIVATE KEY-----"}`},
		{name: "username", body: `{"type":"username","username":"sigma","password":"sigma"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			engine := testkit.NewGin()
			engine.PUT("/:provider/credential", server.WrapRequest(func(_ *gin.Context, req *api.UpdateCodeRepositoryCredentialRequest) {
				require.Equal(t, tt.name, req.Type.String())
			}))

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPut, "/github/credential?provider=github", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			require.NotPanics(t, func() { engine.ServeHTTP(recorder, req) })
			require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
		})
	}
}
