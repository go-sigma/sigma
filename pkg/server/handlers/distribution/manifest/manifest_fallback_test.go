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

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/consts"
)

func TestFallbackProxy(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v2/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/v2/_catalog", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set(consts.HeaderContentType, "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"repositories":["library/alpine"]}`))
		require.NoError(t, err)
	})

	upstream := httptest.NewServer(mux)
	defer upstream.Close()

	h := &handler{Config: &config.Configuration{Proxy: config.ConfigurationProxy{Endpoint: upstream.URL}}}
	_, c := newManifestContext(t, http.MethodGet, "/v2/_catalog", nil)

	statusCode, _, body, err := h.fallbackProxy(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, statusCode)
	require.JSONEq(t, `{"repositories":["library/alpine"]}`, string(body))
}

func TestFallbackProxyUpstreamError(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer upstream.Close()

	h := &handler{Config: &config.Configuration{Proxy: config.ConfigurationProxy{Endpoint: upstream.URL}}}
	_, c := newManifestContext(t, http.MethodGet, "/v2/_catalog", nil)

	_, _, _, err := h.fallbackProxy(c)
	require.Error(t, err)
}
