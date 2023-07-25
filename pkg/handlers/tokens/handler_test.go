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

package token

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/go-sigma/sigma/pkg/dal"
	"github.com/go-sigma/sigma/pkg/inits"
	"github.com/go-sigma/sigma/pkg/logger"
	"github.com/go-sigma/sigma/pkg/tests"
	"github.com/go-sigma/sigma/pkg/validators"
)

const (
	privateKeyString = "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlDWFFJQkFBS0JnUUN2bmwyeU1hRmR0NTJFOFhIN2tFdkVIbnBtelpWbFBTOWFrZTJ5TmQrNm13VXBlaVQ5CnVqVkZwTmJ2RkFna002TUd3dll5N1hkV1FwNTBaOXVVS0d1UlJEZSt4QXQvbklObVZCcVJwU3VnYzhPOVdMNzQKU294UldJSjFVcWJ3NnYvaFU3K1dSMFlORU1ubVlodzJDNXZPQ3c3UlIrQnJET2h5aEtuKzJ3MWRDUUlEQVFBQgpBb0dBSGtjY2VsTnFNY0V0YkRWQVpKSE5Ma1BlOEloelFHQWJJTzlWM3NyQkJ1Z2hMTFI5V2kxWGIrbHFrUStRCkU4Vy9UclFnUkVtQ3NLR050aDROMG01aGxRR3dBS0tsYUhLOWxzYUtPVDBpV0lwYk1HSm1rMWJQZEV5RTRlL1QKcjN2bUMwU0NaZGJOZElkL1FuMzlkY2hZY2I3MGtBaW5kNFlHQXYvNU45UXdSZ0VDUVFEa2JlcnU4bTRRdXhOagpmTysyTUJmL1NoaUtUbHdYZlNXYURvcW9tTE14MG9BeHpwVkU2RzdZMStJd0xYSXd6VEswUXdIUTdDWEl4ZmkvCi9pRyt6T3BCQWtFQXhOQ3ZhSHJhZklpWjVmZVFESlR6T0kzS3B4WDNSWFlaTytDTHlLeHlic0tZQklTSm9Db0YKVkw4K0diRGZJMU9adm5lTXZEcEE3WFhEQkt3TXFHMXd5UUpCQU9BMGRzUWpWUjY4ejdIMW5iNmZnOTVCbHNhaApWTWlGUUJQdXMrLzVPT0RzOElCeWVKWlM0UUdiRzFvWU1SMXZPcFl0c3FtaUx3L2FLR1loaEhPbTQwRUNRRWhLCmZxTlp2TGJSVmZYcUlMYitYdmYrM05qU2NLaks0Q25tS0hIbEpZTVpaczBDQWFzYXhDcUV0RUtyZk1wMUFwdTcKUGE1RmwyT2hSYWlKcVh5VDlrRUNRUUNYdXlrdWR3eXdudEhHL3d2SmVoeWFSYkxGczd5UG1SbUVEL0FHcEY0QgpKcFZrZFJNQVJpa1g1OE84OWF6WXQyT3pkTGNlTWQ3WWlJRGd4UVhBSEcyagotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo="
)

func TestFactory(t *testing.T) {
	viper.Reset()
	logger.SetLevel("debug")
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	validators.Initialize(e)
	err := tests.Initialize(t)
	assert.NoError(t, err)
	err = tests.DB.Init()
	assert.NoError(t, err)
	defer func() {
		conn, err := dal.DB.DB()
		assert.NoError(t, err)
		err = conn.Close()
		assert.NoError(t, err)
		err = tests.DB.DeInit()
		assert.NoError(t, err)
	}()

	viper.SetDefault("auth.internalUser.password", "internal-ximager")
	viper.SetDefault("auth.internalUser.username", "internal-ximager")
	viper.SetDefault("auth.admin.password", "ximager")
	viper.SetDefault("auth.admin.username", "ximager")
	viper.SetDefault("auth.jwt.privateKey", privateKeyString)

	miniRedis := miniredis.RunT(t)
	viper.SetDefault("redis.url", "redis://"+miniRedis.Addr())

	err = inits.Initialize()
	assert.NoError(t, err)

	var f = factory{}
	err = f.Initialize(e)
	assert.NoError(t, err)
}

func TestFactoryFailed(t *testing.T) {
	viper.Reset()
	viper.SetDefault("auth.jwt.privateKey", privateKeyString+"1")
	var f = factory{}
	err := f.Initialize(echo.New())
	assert.Error(t, err)
}
