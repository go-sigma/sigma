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

package manifest

// import (
// 	"fmt"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/labstack/echo/v4"
// 	"github.com/opencontainers/go-digest"
// 	"github.com/stretchr/testify/assert"
// )

// func TestGetManifestFallbackProxyAuthError(t *testing.T) {
// 	mux := http.NewServeMux()

// 	mux.HandleFunc("/v2/", func(w http.ResponseWriter, _ *http.Request) {
// 		w.WriteHeader(http.StatusInternalServerError)
// 	})

// 	s := httptest.NewServer(mux)
// 	defer s.Close()

// 	h := &handler{}

// 	// test about proxy server auth internal server error
// 	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v2/library/busybox/manifests/%s", "sha256:f7d81d5be30e617068bf53a9b136400b13d91c0f54d097a72bf91127f43d0151"), nil)
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec := httptest.NewRecorder()
// 	c := echo.New().NewContext(req, rec)
// 	err := h.getManifestFallbackProxy(c, Refs{Digest: digest.Digest("sha256:f7d81d5be30e617068bf53a9b136400b13d91c0f54d097a72bf91127f43d0151")})
// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusInternalServerError, rec.Code)
// }

// func TestGetManifest(t *testing.T) {

// }
