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

package bootstrap

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/consts"
	reposetting "github.com/go-sigma/sigma/pkg/dal/repository/setting"
	svctoken "github.com/go-sigma/sigma/pkg/service/token"
	"github.com/go-sigma/sigma/pkg/testkit"
)

func newJwtTestConfig() *config.Configuration {
	return &config.Configuration{
		Locker: config.ConfigurationLocker{
			Type:   enums.LockerTypeRedis,
			Prefix: "sigma-locker",
		},
		Redis: config.ConfigurationRedis{
			URL: "redis://:sigma@localhost:6379/0",
		},
	}
}

func newJwtDigContainer(t *testing.T, cfg *config.Configuration) *dig.Container {
	t.Helper()

	digCon := dig.New()
	require.NoError(t, digCon.Provide(func() *config.Configuration { return cfg }))
	require.NoError(t, testkit.InitializeIntegration(t, digCon))
	return digCon
}

func TestJwtBootstrapGeneratesKey(t *testing.T) {
	cfg := newJwtTestConfig()
	digCon := newJwtDigContainer(t, cfg)

	require.NoError(t, jwtKey(digCon))
	require.NotEmpty(t, cfg.Auth.Jwt.PrivateKey)

	settingObj, err := reposetting.NewSettingRepository().Get(t.Context(), consts.SettingJwtPrivateKey)
	require.NoError(t, err)
	require.Equal(t, cfg.Auth.Jwt.PrivateKey, string(settingObj.Val))

	privateKey, err := svctoken.ParsePrivateKey(cfg.Auth.Jwt.PrivateKey)
	require.NoError(t, err)
	require.Len(t, privateKey, 64)

	// A restart must reuse the stored key, otherwise every token issued before
	// the restart would stop verifying.
	cfg.Auth.Jwt.PrivateKey = ""
	require.NoError(t, jwtKey(digCon))
	require.Equal(t, string(settingObj.Val), cfg.Auth.Jwt.PrivateKey)
}

func TestJwtBootstrapLoadsStoredKey(t *testing.T) {
	cfg := newJwtTestConfig()
	digCon := newJwtDigContainer(t, cfg)

	key, err := svctoken.GeneratePrivateKey()
	require.NoError(t, err)
	require.NoError(t, reposetting.NewSettingRepository().Create(t.Context(), consts.SettingJwtPrivateKey, []byte(key)))

	require.NoError(t, jwtKey(digCon))
	require.Equal(t, key, cfg.Auth.Jwt.PrivateKey)
}

func TestJwtBootstrapWithConfiguredKey(t *testing.T) {
	cfg := newJwtTestConfig()
	configured, err := svctoken.GeneratePrivateKey()
	require.NoError(t, err)
	cfg.Auth.Jwt.PrivateKey = configured

	digCon := newJwtDigContainer(t, cfg)

	// The settings table already holds a different key: the configured one is
	// authoritative.
	stored, err := svctoken.GeneratePrivateKey()
	require.NoError(t, err)
	require.NoError(t, reposetting.NewSettingRepository().Create(t.Context(), consts.SettingJwtPrivateKey, []byte(stored)))

	require.NoError(t, jwtKey(digCon))
	require.Equal(t, configured, cfg.Auth.Jwt.PrivateKey)
}

func TestJwtBootstrapRejectsInvalidConfiguredKey(t *testing.T) {
	cfg := newJwtTestConfig()
	cfg.Auth.Jwt.PrivateKey = newRSAKeyString(t)
	digCon := newJwtDigContainer(t, cfg)

	require.Error(t, jwtKey(digCon))
}

func newRSAKeyString(t *testing.T) string {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	privateKeyDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)
	return base64.StdEncoding.EncodeToString(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyDER}))
}
