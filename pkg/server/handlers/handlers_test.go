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
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/app/bootstrap"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/infra/registry"
	"github.com/go-sigma/sigma/pkg/logger"
	"github.com/go-sigma/sigma/pkg/service/password"
	"github.com/go-sigma/sigma/pkg/testkit"
	"github.com/go-sigma/sigma/pkg/validators"

	_ "github.com/go-sigma/sigma/pkg/infra/lock/inmemory"
)

const (
	privateKeyString = "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlDWFFJQkFBS0JnUUN2bmwyeU1hRmR0NTJFOFhIN2tFdkVIbnBtelpWbFBTOWFrZTJ5TmQrNm13VXBlaVQ5CnVqVkZwTmJ2RkFna002TUd3dll5N1hkV1FwNTBaOXVVS0d1UlJEZSt4QXQvbklObVZCcVJwU3VnYzhPOVdMNzQKU294UldJSjFVcWJ3NnYvaFU3K1dSMFlORU1ubVlodzJDNXZPQ3c3UlIrQnJET2h5aEtuKzJ3MWRDUUlEQVFBQgpBb0dBSGtjY2VsTnFNY0V0YkRWQVpKSE5Ma1BlOEloelFHQWJJTzlWM3NyQkJ1Z2hMTFI5V2kxWGIrbHFrUStRCkU4Vy9UclFnUkVtQ3NLR050aDROMG01aGxRR3dBS0tsYUhLOWxzYUtPVDBpV0lwYk1HSm1rMWJQZEV5RTRlL1QKcjN2bUMwU0NaZGJOZElkL1FuMzlkY2hZY2I3MGtBaW5kNFlHQXYvNU45UXdSZ0VDUVFEa2JlcnU4bTRRdXhOagpmTysyTUJmL1NoaUtUbHdYZlNXYURvcW9tTE14MG9BeHpwVkU2RzdZMStJd0xYSXd6VEswUXdIUTdDWEl4ZmkvCi9pRyt6T3BCQWtFQXhOQ3ZhSHJhZklpWjVmZVFESlR6T0kzS3B4WDNSWFlaTytDTHlLeHlic0tZQklTSm9Db0YKVkw4K0diRGZJMU9adm5lTXZEcEE3WFhEQkt3TXFHMXd5UUpCQU9BMGRzUWpWUjY4ejdIMW5iNmZnOTVCbHNhaApWTWlGUUJQdXMrLzVPT0RzOElCeWVKWlM0UUdiRzFvWU1SMXZPcFl0c3FtaUx3L2FLR1loaEhPbTQwRUNRRWhLCmZxTlp2TGJSVmZYcUlMYitYdmYrM05qU2NLaks0Q25tS0hIbEpZTVpaczBDQWFzYXhDcUV0RUtyZk1wMUFwdTcKUGE1RmwyT2hSYWlKcVh5VDlrRUNRUUNYdXlrdWR3eXdudEhHL3d2SmVoeWFSYkxGczd5UG1SbUVEL0FHcEY0QgpKcFZrZFJNQVJpa1g1OE84OWF6WXQyT3pkTGNlTWQ3WWlJRGd4UVhBSEcyagotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo="
)

func TestInitializeSkipAuth(t *testing.T) {
	logger.SetLevel("debug")

	digCon := dig.New()
	require.NoError(t, digCon.Provide(testkit.NewGin))
	require.NoError(t, validators.Initialize())

	err := digCon.Provide(func() *config.Configuration {
		return &config.Configuration{
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
	})
	require.NoError(t, err)

	err = digCon.Provide(func() password.Service {
		return password.New()
	})
	require.NoError(t, err)

	require.NoError(t, testkit.InitializeIntegration(t, digCon))

	require.NoError(t, bootstrap.Initialize(digCon))

	require.NoError(t, Initialize(digCon))
}

type factoryOk struct{}

func (f *factoryOk) Initialize(*dig.Container) error {
	return nil
}

func TestInitializeOK(t *testing.T) {
	Routers = make(registry.Factories[string, Factory])
	require.NoError(t, Routers.Register("ok", &factoryOk{}))
	digCon := dig.New()
	require.NoError(t, digCon.Provide(testkit.NewGin))
	require.NoError(t, Initialize(digCon))
}

type factoryErr struct{}

func (f *factoryErr) Initialize(*dig.Container) error {
	return errors.New("error")
}

func TestInitializeErr(t *testing.T) {
	Routers = make(registry.Factories[string, Factory])
	require.NoError(t, Routers.Register("err", &factoryErr{}))
	digCon := dig.New()
	require.NoError(t, digCon.Provide(testkit.NewGin))
	require.Error(t, Initialize(digCon))
}

func TestInitializeDistributionMissingEngine(t *testing.T) {
	require.Error(t, InitializeDistribution(dig.New()))
}

func TestInitializeDup(t *testing.T) {
	Routers = make(registry.Factories[string, Factory])
	require.NoError(t, Routers.Register("err", &factoryErr{}))
	require.Error(t, Routers.Register("err", &factoryErr{}))
}
