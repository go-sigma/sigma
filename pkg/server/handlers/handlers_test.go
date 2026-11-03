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

package handlers

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/app/bootstrap"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/infra/counter"
	"github.com/go-sigma/sigma/pkg/infra/workq"
	"github.com/go-sigma/sigma/pkg/logger"
	"github.com/go-sigma/sigma/pkg/server/validators"
	"github.com/go-sigma/sigma/pkg/service/password"
	"github.com/go-sigma/sigma/pkg/service/token"
	"github.com/go-sigma/sigma/pkg/storage"
	"github.com/go-sigma/sigma/pkg/testkit"
)

// privateKeyString is generated at runtime so the test never carries key
// material; the JWT bootstrap validates it and expects Ed25519.
var privateKeyString = func() string {
	key, err := token.GeneratePrivateKey()
	if err != nil {
		panic(err)
	}
	return key
}()

func TestInitializeSkipAuth(t *testing.T) {
	logger.SetLevel("debug")

	digCon := dig.New()
	require.NoError(t, digCon.Provide(testkit.NewGin))
	require.NoError(t, validators.Initialize())

	cfg := &config.Configuration{
		Auth: config.ConfigurationAuth{
			Admin: config.ConfigurationAuthAdmin{
				Username: "sigma",
				Password: "sigma",
				Email:    "sigma@gmail.com",
			},
			Jwt: config.ConfigurationAuthJwt{
				PrivateKey: privateKeyString,
			},
		},
		Locker: config.ConfigurationLocker{
			Type:   enums.LockerTypeInmemory,
			Prefix: "sigma-locker",
		},
	}
	cfg.WithDefaults()
	err := digCon.Provide(func() *config.Configuration { return cfg })
	require.NoError(t, err)

	err = digCon.Provide(func() password.Service {
		return password.New()
	})
	require.NoError(t, err)
	require.NoError(t, digCon.Provide(func() counter.Counter { return nil }))
	require.NoError(t, digCon.Provide(func() workq.Producer { return nil }))
	require.NoError(t, digCon.Provide(func() token.Service { return nil }))
	require.NoError(t, digCon.Provide(func() storage.StorageDriver { return nil }))

	require.NoError(t, testkit.InitializeIntegration(t, digCon))

	require.NoError(t, bootstrap.Initialize(digCon))

	require.NoError(t, Initialize(digCon))
}

func TestInitializeDistributionMissingEngine(t *testing.T) {
	require.Error(t, InitializeDistribution(dig.New()))
}
