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

package crypt

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSeal(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		plaintext string
	}{
		{name: "common", key: "sigma", plaintext: "sigma"},
		{name: "empty", key: "sigma", plaintext: ""},
		{name: "unicode", key: "sigma", plaintext: "口令-🔐"},
		{name: "long", key: "sigma", plaintext: strings.Repeat("ghp_", 512)},
		{name: "long key", key: strings.Repeat("k", 32), plaintext: "ghp_sigma"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ciphertext, err := Seal(tt.key, tt.plaintext)
			require.NoError(t, err)
			require.NotEqual(t, tt.plaintext, ciphertext)

			plaintext, err := Open(tt.key, ciphertext)
			require.NoError(t, err)
			require.Equal(t, tt.plaintext, plaintext)
		})
	}
}

func TestSealRandomized(t *testing.T) {
	first, err := Seal("sigma", "ghp_sigma")
	require.NoError(t, err)
	second, err := Seal("sigma", "ghp_sigma")
	require.NoError(t, err)
	require.NotEqual(t, first, second, "seal must use a random nonce")
}

func TestMustSeal(t *testing.T) {
	require.NotPanics(t, func() { MustSeal("sigma", "sigma") })
}

func TestOpenErrors(t *testing.T) {
	ciphertext, err := Seal("sigma", "ghp_sigma")
	require.NoError(t, err)

	t.Run("wrong key", func(t *testing.T) {
		_, err := Open("another", ciphertext)
		require.Error(t, err)
	})

	t.Run("tampered", func(t *testing.T) {
		raw, err := base64.StdEncoding.DecodeString(ciphertext)
		require.NoError(t, err)
		raw[len(raw)-1] ^= 0xff
		_, err = Open("sigma", base64.StdEncoding.EncodeToString(raw))
		require.Error(t, err)
	})

	t.Run("not base64", func(t *testing.T) {
		_, err := Open("sigma", "not-base64!!")
		require.Error(t, err)
	})

	t.Run("too short", func(t *testing.T) {
		_, err := Open("sigma", base64.StdEncoding.EncodeToString([]byte("short")))
		require.Error(t, err)
	})
}
