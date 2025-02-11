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

package users

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/tests"
	"github.com/go-sigma/sigma/pkg/utils/password"
	"github.com/go-sigma/sigma/pkg/utils/token"
)

// const (
// 	privateKeyString = "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlDWFFJQkFBS0JnUUN2bmwyeU1hRmR0NTJFOFhIN2tFdkVIbnBtelpWbFBTOWFrZTJ5TmQrNm13VXBlaVQ5CnVqVkZwTmJ2RkFna002TUd3dll5N1hkV1FwNTBaOXVVS0d1UlJEZSt4QXQvbklObVZCcVJwU3VnYzhPOVdMNzQKU294UldJSjFVcWJ3NnYvaFU3K1dSMFlORU1ubVlodzJDNXZPQ3c3UlIrQnJET2h5aEtuKzJ3MWRDUUlEQVFBQgpBb0dBSGtjY2VsTnFNY0V0YkRWQVpKSE5Ma1BlOEloelFHQWJJTzlWM3NyQkJ1Z2hMTFI5V2kxWGIrbHFrUStRCkU4Vy9UclFnUkVtQ3NLR050aDROMG01aGxRR3dBS0tsYUhLOWxzYUtPVDBpV0lwYk1HSm1rMWJQZEV5RTRlL1QKcjN2bUMwU0NaZGJOZElkL1FuMzlkY2hZY2I3MGtBaW5kNFlHQXYvNU45UXdSZ0VDUVFEa2JlcnU4bTRRdXhOagpmTysyTUJmL1NoaUtUbHdYZlNXYURvcW9tTE14MG9BeHpwVkU2RzdZMStJd0xYSXd6VEswUXdIUTdDWEl4ZmkvCi9pRyt6T3BCQWtFQXhOQ3ZhSHJhZklpWjVmZVFESlR6T0kzS3B4WDNSWFlaTytDTHlLeHlic0tZQklTSm9Db0YKVkw4K0diRGZJMU9adm5lTXZEcEE3WFhEQkt3TXFHMXd5UUpCQU9BMGRzUWpWUjY4ejdIMW5iNmZnOTVCbHNhaApWTWlGUUJQdXMrLzVPT0RzOElCeWVKWlM0UUdiRzFvWU1SMXZPcFl0c3FtaUx3L2FLR1loaEhPbTQwRUNRRWhLCmZxTlp2TGJSVmZYcUlMYitYdmYrM05qU2NLaks0Q25tS0hIbEpZTVpaczBDQWFzYXhDcUV0RUtyZk1wMUFwdTcKUGE1RmwyT2hSYWlKcVh5VDlrRUNRUUNYdXlrdWR3eXdudEhHL3d2SmVoeWFSYkxGczd5UG1SbUVEL0FHcEY0QgpKcFZrZFJNQVJpa1g1OE84OWF6WXQyT3pkTGNlTWQ3WWlJRGd4UVhBSEcyagotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo="
// )

func TestFactory(t *testing.T) {
	digCon := dig.New()
	require.NoError(t, digCon.Provide(func() configs.Configuration { return configs.Configuration{} }))
	require.NoError(t, digCon.Provide(func() token.Service { return nil }))
	require.NoError(t, digCon.Provide(func() password.Service { return nil }))
	require.NoError(t, digCon.Provide(func() dao.UserServiceFactory { return nil }))
	require.NoError(t, digCon.Provide(tests.NewEcho))
	require.NoError(t, factory{}.Initialize(digCon))
}

// func TestFactory(t *testing.T) {
// 	logger.SetLevel("debug")
// 	e := echo.New()
// 	e.HideBanner = true
// 	e.HidePort = true
// 	validators.Initialize(e)
// 	assert.NoError(t, tests.Initialize(t))
// 	assert.NoError(t, tests.DB.Init())
// 	defer func() {
// 		conn, err := dal.DB.DB()
// 		assert.NoError(t, err)
// 		assert.NoError(t, conn.Close())
// 		assert.NoError(t, tests.DB.DeInit())
// 	}()

// 	config := &configs.Configuration{
// 		Auth: configs.ConfigurationAuth{
// 			Admin: configs.ConfigurationAuthAdmin{
// 				Username: "sigma",
// 				Password: "sigma",
// 				Email:    "sigma@gmail.com",
// 			},
// 			Jwt: configs.ConfigurationAuthJwt{
// 				PrivateKey: privateKeyString,
// 			},
// 		},
// 	}
// 	configs.SetConfiguration(config)

// 	assert.NoError(t, inits.Initialize(ptr.To(configs.GetConfiguration())))

// 	var f = factory{}
// 	err := f.Initialize(e)
// 	assert.NoError(t, err)

// 	go func() {
// 		err = e.Start(":8080")
// 		assert.ErrorIs(t, err, http.ErrServerClosed)
// 	}()

// 	time.Sleep(1 * time.Second)

// 	url := "http://127.0.0.1:8080/user/login"

// 	req, err := http.NewRequest("GET", url, nil)
// 	assert.NoError(t, err)
// 	req.SetBasicAuth("sigma", "sigma")

// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	assert.NoError(t, err)
// 	err = resp.Body.Close()
// 	assert.NoError(t, err)

// 	assert.NoError(t, e.Shutdown(context.Background()))
// }

// func TestFactoryFailed(t *testing.T) {
// 	config := &configs.Configuration{
// 		Auth: configs.ConfigurationAuth{
// 			Admin: configs.ConfigurationAuthAdmin{
// 				Username: "sigma",
// 				Password: "sigma",
// 				Email:    "sigma@gmail.com",
// 			},
// 			Jwt: configs.ConfigurationAuthJwt{
// 				PrivateKey: privateKeyString + "1",
// 			},
// 		},
// 	}
// 	configs.SetConfiguration(config)
// 	assert.Error(t, factory{}.Initialize(echo.New()))
// }
