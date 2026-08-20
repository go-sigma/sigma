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
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewAPIClientTLSVerify(t *testing.T) {
	client := newAPIClient("token", "https://sigma.example", true)
	transport, err := client.cli.HTTPTransport()
	require.NoError(t, err)
	require.False(t, tlsInsecureSkipVerify(transport))
}

func TestNewAPIClientTLSVerifyDisabled(t *testing.T) {
	client := newAPIClient("token", "https://sigma.example", false)
	transport, err := client.cli.HTTPTransport()
	require.NoError(t, err)
	require.True(t, tlsInsecureSkipVerify(transport))
}

func TestNewAPIClientTLSVerifyDisabledForHTTP(t *testing.T) {
	client := newAPIClient("token", "http://sigma.example", false)
	transport, err := client.cli.HTTPTransport()
	require.NoError(t, err)
	require.False(t, tlsInsecureSkipVerify(transport))
}

func tlsInsecureSkipVerify(transport *http.Transport) bool {
	return transport != nil && transport.TLSClientConfig != nil && transport.TLSClientConfig.InsecureSkipVerify
}
