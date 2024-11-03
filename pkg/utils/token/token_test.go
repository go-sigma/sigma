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
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/types/enums"
)

const (
	privateKeyString        = "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlDWFFJQkFBS0JnUUN2bmwyeU1hRmR0NTJFOFhIN2tFdkVIbnBtelpWbFBTOWFrZTJ5TmQrNm13VXBlaVQ5CnVqVkZwTmJ2RkFna002TUd3dll5N1hkV1FwNTBaOXVVS0d1UlJEZSt4QXQvbklObVZCcVJwU3VnYzhPOVdMNzQKU294UldJSjFVcWJ3NnYvaFU3K1dSMFlORU1ubVlodzJDNXZPQ3c3UlIrQnJET2h5aEtuKzJ3MWRDUUlEQVFBQgpBb0dBSGtjY2VsTnFNY0V0YkRWQVpKSE5Ma1BlOEloelFHQWJJTzlWM3NyQkJ1Z2hMTFI5V2kxWGIrbHFrUStRCkU4Vy9UclFnUkVtQ3NLR050aDROMG01aGxRR3dBS0tsYUhLOWxzYUtPVDBpV0lwYk1HSm1rMWJQZEV5RTRlL1QKcjN2bUMwU0NaZGJOZElkL1FuMzlkY2hZY2I3MGtBaW5kNFlHQXYvNU45UXdSZ0VDUVFEa2JlcnU4bTRRdXhOagpmTysyTUJmL1NoaUtUbHdYZlNXYURvcW9tTE14MG9BeHpwVkU2RzdZMStJd0xYSXd6VEswUXdIUTdDWEl4ZmkvCi9pRyt6T3BCQWtFQXhOQ3ZhSHJhZklpWjVmZVFESlR6T0kzS3B4WDNSWFlaTytDTHlLeHlic0tZQklTSm9Db0YKVkw4K0diRGZJMU9adm5lTXZEcEE3WFhEQkt3TXFHMXd5UUpCQU9BMGRzUWpWUjY4ejdIMW5iNmZnOTVCbHNhaApWTWlGUUJQdXMrLzVPT0RzOElCeWVKWlM0UUdiRzFvWU1SMXZPcFl0c3FtaUx3L2FLR1loaEhPbTQwRUNRRWhLCmZxTlp2TGJSVmZYcUlMYitYdmYrM05qU2NLaks0Q25tS0hIbEpZTVpaczBDQWFzYXhDcUV0RUtyZk1wMUFwdTcKUGE1RmwyT2hSYWlKcVh5VDlrRUNRUUNYdXlrdWR3eXdudEhHL3d2SmVoeWFSYkxGczd5UG1SbUVEL0FHcEY0QgpKcFZrZFJNQVJpa1g1OE84OWF6WXQyT3pkTGNlTWQ3WWlJRGd4UVhBSEcyagotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo="
	privateInvalidKeyString = "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlDWFFJQkFBS0JnUUN2bmwyeU1hRmR0NTJFOFhIN2tFdkVIbnBtelpWbFBTOWFrZTJ5TmQrNm13VXBlaVQ5CnVqVkZwTmJ2RkFna002TUd3dll5N1hkV1FwNTBaOXVVS0d1UlJEZSt4QXQvbklObVZCcVJwU3VnYzhPOVdMNzQKU294UldJSjFVcWJ3NnYvaFU3K1dSMFlORU1ubVlodzJDNXZPQ3c3UlIrQnJET2h5aEtuKzJ3MWRDUUlEQVFBQgpBb0dBSGtjY2VsTnFNY0V0YkRWQVpKSE5Ma1BlOEloelFHQWJJTzlWM3NyQkJ1Z2hMTFI5V2kxWGIrbHFrUStRCkU4Vy9UclFnUkVtQ3NLR050aDROMG01aGxRR3dBS0tsYUhLOWxzYUtPVDBpV0lwYk1HSm1rMWJQZEV5RTRlL1QKcjN2bUMwU0NaZGJOZElkL1FuMzlkY2hZY2I3MGtBaW5kNFlHQXYvNU45UXdSZ0VDUVFEa2JlcnU4bTRRdXhOagpmTysyTUJmL1NoaUtUbHdYZlNXYURvcW9tTE14MG9BeHpwVkU2RzdZMStJd0xYSXd6VEswUXdIUTdDWEl4ZmkvCi9pRyt6T3BCQWtFQXhOQ3ZhSHJhZklpWjVmZVFESlR6T0kzS3B4WDNSWFlaTytDTHlLeHlic0tZQklTSm9Db0YKVkw4K0diRGZJMU9adm5lTXZEcEE3WFhEQkt3TXFHMXd5UUpCQU9BMGRzUWpWUjY4ejdIMW5iNmZnOTVCbHNhaApWTWlGUUJQdXMrLzVPT0RzOElCeWVKWlM0UUdiRzFvWU1SMXZPcFl0c3FtaUx3L2FLR1loaEhPbTQwRUNRRWhLCmZxTlp2TGJSVmZYcUlMYitYdmYrM05qU2NLaks0Q25tS0hIbEpZTVpaczBDQWFzYXhDcUV0RUtyZk1wMUFwdTcKUGE1RmwyT2hSYWlKcVh5VDlrRUNRUUNYdXlrdWR3eXdudEhHL3d2SmVoeWFSYkxGczd5UG1SbUVEL0FHcEY0QgpKcFZrZFJNQVJpa1g1OE84OWF6WXQyT3pkTGNlTWQ3WWlJRGd4UVhBSEcyagotLS0tLUVORCBSU0EgUFJJVkFURSBLRVkxLS0tLS0="
)

func TestJWTClaimsValid(t *testing.T) {
	claims := &JWTClaims{}
	assert.NoError(t, claims.Valid())
}

func TestNew(t *testing.T) {
	tests := []struct {
		name      string
		newDigCon func(*testing.T) *dig.Container
		want      Service
		wantErr   bool
	}{
		{
			name: "bad redis",
			newDigCon: func(t *testing.T) *dig.Container {
				miniRedis := miniredis.RunT(t)
				digCon := dig.New()
				err := digCon.Provide(func() configs.Configuration {
					return configs.Configuration{
						Redis: configs.ConfigurationRedis{
							Type: enums.RedisTypeExternal,
							URL:  "redis:////" + miniRedis.Addr(),
						},
						Cache: configs.ConfigurationCache{
							Type: enums.CacherTypeRedis,
						},
						Auth: configs.ConfigurationAuth{
							Jwt: configs.ConfigurationAuthJwt{
								PrivateKey: privateKeyString,
							},
						},
					}
				})
				assert.NoError(t, err)
				return digCon
			},
			wantErr: true,
		},
		{
			name: "invalid key",
			newDigCon: func(t *testing.T) *dig.Container {
				miniRedis := miniredis.RunT(t)
				digCon := dig.New()
				err := digCon.Provide(func() configs.Configuration {
					return configs.Configuration{
						Redis: configs.ConfigurationRedis{
							Type: enums.RedisTypeExternal,
							URL:  "redis://" + miniRedis.Addr(),
						},
						Cache: configs.ConfigurationCache{
							Type: enums.CacherTypeRedis,
						},
						Auth: configs.ConfigurationAuth{
							Jwt: configs.ConfigurationAuthJwt{
								PrivateKey: privateInvalidKeyString,
							},
						},
					}
				})
				assert.NoError(t, err)
				return digCon
			},
			wantErr: true,
		},
		{
			name: "bad key",
			newDigCon: func(t *testing.T) *dig.Container {
				miniRedis := miniredis.RunT(t)
				digCon := dig.New()
				err := digCon.Provide(func() configs.Configuration {
					return configs.Configuration{
						Redis: configs.ConfigurationRedis{
							Type: enums.RedisTypeExternal,
							URL:  "redis://" + miniRedis.Addr(),
						},
						Cache: configs.ConfigurationCache{
							Type: enums.CacherTypeRedis,
						},
						Auth: configs.ConfigurationAuth{
							Jwt: configs.ConfigurationAuthJwt{
								PrivateKey: privateKeyString + "-",
							},
						},
					}
				})
				assert.NoError(t, err)
				return digCon
			},
			wantErr: true,
		},
		{
			name: "normal",
			newDigCon: func(t *testing.T) *dig.Container {
				digCon := dig.New()
				miniRedis := miniredis.RunT(t)
				err := digCon.Provide(func() configs.Configuration {
					return configs.Configuration{
						Redis: configs.ConfigurationRedis{
							Type: enums.RedisTypeExternal,
							URL:  "redis://" + miniRedis.Addr(),
						},
						Cache: configs.ConfigurationCache{
							Type: enums.CacherTypeRedis,
						},
						Auth: configs.ConfigurationAuth{
							Jwt: configs.ConfigurationAuthJwt{
								PrivateKey: privateKeyString,
							},
						},
					}
				})
				assert.NoError(t, err)
				return digCon
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenService, err := New(tt.newDigCon(t))
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			token, err := tokenService.New(100, time.Second*30)
			assert.NoError(t, err)
			assert.NotEmpty(t, token)

			id, uid, err := tokenService.Validate(context.Background(), token)
			assert.NoError(t, err)
			assert.Equal(t, int64(100), uid)
			_, err = uuid.Parse(id)
			assert.NoError(t, err)

			err = tokenService.Revoke(context.Background(), id)
			assert.NoError(t, err)

			_, _, err = tokenService.Validate(context.Background(), token)
			assert.Error(t, err)
		})
	}
}
