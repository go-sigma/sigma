// Copyright 2023 XImager
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
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/ximager/ximager/pkg/dal/models"
	"github.com/ximager/ximager/pkg/utils"
)

const (
	privateKeyString        = "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlDWFFJQkFBS0JnUUN2bmwyeU1hRmR0NTJFOFhIN2tFdkVIbnBtelpWbFBTOWFrZTJ5TmQrNm13VXBlaVQ5CnVqVkZwTmJ2RkFna002TUd3dll5N1hkV1FwNTBaOXVVS0d1UlJEZSt4QXQvbklObVZCcVJwU3VnYzhPOVdMNzQKU294UldJSjFVcWJ3NnYvaFU3K1dSMFlORU1ubVlodzJDNXZPQ3c3UlIrQnJET2h5aEtuKzJ3MWRDUUlEQVFBQgpBb0dBSGtjY2VsTnFNY0V0YkRWQVpKSE5Ma1BlOEloelFHQWJJTzlWM3NyQkJ1Z2hMTFI5V2kxWGIrbHFrUStRCkU4Vy9UclFnUkVtQ3NLR050aDROMG01aGxRR3dBS0tsYUhLOWxzYUtPVDBpV0lwYk1HSm1rMWJQZEV5RTRlL1QKcjN2bUMwU0NaZGJOZElkL1FuMzlkY2hZY2I3MGtBaW5kNFlHQXYvNU45UXdSZ0VDUVFEa2JlcnU4bTRRdXhOagpmTysyTUJmL1NoaUtUbHdYZlNXYURvcW9tTE14MG9BeHpwVkU2RzdZMStJd0xYSXd6VEswUXdIUTdDWEl4ZmkvCi9pRyt6T3BCQWtFQXhOQ3ZhSHJhZklpWjVmZVFESlR6T0kzS3B4WDNSWFlaTytDTHlLeHlic0tZQklTSm9Db0YKVkw4K0diRGZJMU9adm5lTXZEcEE3WFhEQkt3TXFHMXd5UUpCQU9BMGRzUWpWUjY4ejdIMW5iNmZnOTVCbHNhaApWTWlGUUJQdXMrLzVPT0RzOElCeWVKWlM0UUdiRzFvWU1SMXZPcFl0c3FtaUx3L2FLR1loaEhPbTQwRUNRRWhLCmZxTlp2TGJSVmZYcUlMYitYdmYrM05qU2NLaks0Q25tS0hIbEpZTVpaczBDQWFzYXhDcUV0RUtyZk1wMUFwdTcKUGE1RmwyT2hSYWlKcVh5VDlrRUNRUUNYdXlrdWR3eXdudEhHL3d2SmVoeWFSYkxGczd5UG1SbUVEL0FHcEY0QgpKcFZrZFJNQVJpa1g1OE84OWF6WXQyT3pkTGNlTWQ3WWlJRGd4UVhBSEcyagotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo="
	publicKeyString         = "LS0tLS1CRUdJTiBQVUJMSUMgS0VZLS0tLS0KTUlHZk1BMEdDU3FHU0liM0RRRUJBUVVBQTRHTkFEQ0JpUUtCZ1FDdm5sMnlNYUZkdDUyRThYSDdrRXZFSG5wbQp6WlZsUFM5YWtlMnlOZCs2bXdVcGVpVDl1alZGcE5idkZBZ2tNNk1Hd3ZZeTdYZFdRcDUwWjl1VUtHdVJSRGUrCnhBdC9uSU5tVkJxUnBTdWdjOE85V0w3NFNveFJXSUoxVXFidzZ2L2hVNytXUjBZTkVNbm1ZaHcyQzV2T0N3N1IKUitCckRPaHloS24rMncxZENRSURBUUFCCi0tLS0tRU5EIFBVQkxJQyBLRVktLS0tLQo="
	privateInvalidKeyString = "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlDWFFJQkFBS0JnUUN2bmwyeU1hRmR0NTJFOFhIN2tFdkVIbnBtelpWbFBTOWFrZTJ5TmQrNm13VXBlaVQ5CnVqVkZwTmJ2RkFna002TUd3dll5N1hkV1FwNTBaOXVVS0d1UlJEZSt4QXQvbklObVZCcVJwU3VnYzhPOVdMNzQKU294UldJSjFVcWJ3NnYvaFU3K1dSMFlORU1ubVlodzJDNXZPQ3c3UlIrQnJET2h5aEtuKzJ3MWRDUUlEQVFBQgpBb0dBSGtjY2VsTnFNY0V0YkRWQVpKSE5Ma1BlOEloelFHQWJJTzlWM3NyQkJ1Z2hMTFI5V2kxWGIrbHFrUStRCkU4Vy9UclFnUkVtQ3NLR050aDROMG01aGxRR3dBS0tsYUhLOWxzYUtPVDBpV0lwYk1HSm1rMWJQZEV5RTRlL1QKcjN2bUMwU0NaZGJOZElkL1FuMzlkY2hZY2I3MGtBaW5kNFlHQXYvNU45UXdSZ0VDUVFEa2JlcnU4bTRRdXhOagpmTysyTUJmL1NoaUtUbHdYZlNXYURvcW9tTE14MG9BeHpwVkU2RzdZMStJd0xYSXd6VEswUXdIUTdDWEl4ZmkvCi9pRyt6T3BCQWtFQXhOQ3ZhSHJhZklpWjVmZVFESlR6T0kzS3B4WDNSWFlaTytDTHlLeHlic0tZQklTSm9Db0YKVkw4K0diRGZJMU9adm5lTXZEcEE3WFhEQkt3TXFHMXd5UUpCQU9BMGRzUWpWUjY4ejdIMW5iNmZnOTVCbHNhaApWTWlGUUJQdXMrLzVPT0RzOElCeWVKWlM0UUdiRzFvWU1SMXZPcFl0c3FtaUx3L2FLR1loaEhPbTQwRUNRRWhLCmZxTlp2TGJSVmZYcUlMYitYdmYrM05qU2NLaks0Q25tS0hIbEpZTVpaczBDQWFzYXhDcUV0RUtyZk1wMUFwdTcKUGE1RmwyT2hSYWlKcVh5VDlrRUNRUUNYdXlrdWR3eXdudEhHL3d2SmVoeWFSYkxGczd5UG1SbUVEL0FHcEY0QgpKcFZrZFJNQVJpa1g1OE84OWF6WXQyT3pkTGNlTWQ3WWlJRGd4UVhBSEcyagotLS0tLUVORCBSU0EgUFJJVkFURSBLRVkxLS0tLS0="
	publicInvalidKeyString  = "LS0tLS1CRUdJTiBQVUJMSUMgS0VZLS0tLS0KTUlHZk1BMEdDU3FHU0liM0RRRUJBUVVBQTRHTkFEQ0JpUUtCZ1FDdm5sMnlNYUZkdDUyRThYSDdrRXZFSG5wbQp6WlZsUFM5YWtlMnlOZCs2bXdVcGVpVDl1alZGcE5idkZBZ2tNNk1Hd3ZZeTdYZFdRcDUwWjl1VUtHdVJSRGUrCnhBdC9uSU5tVkJxUnBTdWdjOE85V0w3NFNveFJXSUoxVXFidzZ2L2hVNytXUjBZTkVNbm1ZaHcyQzV2T0N3N1IKUitCckRPaHloS24rMncxZENRSURBUUFCCi0tLS0tRU5EIFBVQkxJQyBLRVkxLS0tLS0="
)

func TestNew(t *testing.T) {
	utils.SetLevel(0)

	miniRedis := miniredis.RunT(t)
	viper.SetDefault("redis.url", "redis:////"+miniRedis.Addr())

	_, err := NewTokenService(privateKeyString, publicKeyString)
	assert.Error(t, err)

	viper.SetDefault("redis.url", "redis://"+miniRedis.Addr())

	viper.SetDefault("auth.jwt.expire", time.Second)

	_, err = NewTokenService(privateKeyString+"-", publicKeyString)
	assert.Error(t, err)

	_, err = NewTokenService(privateKeyString, publicKeyString+"-")
	assert.Error(t, err)

	_, err = NewTokenService(privateInvalidKeyString, publicKeyString)
	assert.Error(t, err)

	_, err = NewTokenService(privateKeyString, publicInvalidKeyString)
	assert.Error(t, err)

	tokenService, err := NewTokenService(privateKeyString, publicKeyString)
	assert.NoError(t, err)
	assert.NotNil(t, tokenService)

	token, err := tokenService.New(&models.User{Username: "test", Role: "root"}, time.Second*30)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	id, username, err := tokenService.Validate(context.Background(), token)
	assert.NoError(t, err)
	assert.Equal(t, "test", username)
	_, err = uuid.Parse(id)
	assert.NoError(t, err)

	err = tokenService.Revoke(context.Background(), id)
	assert.NoError(t, err)

	_, _, err = tokenService.Validate(context.Background(), token)
	assert.Error(t, err)

	miniRedis.Close()
	err = tokenService.Revoke(context.Background(), id)
	assert.Error(t, err)

	_, _, err = tokenService.Validate(context.Background(), token)
	assert.Error(t, err)
}

func TestJWTClaimsValid(t *testing.T) {
	claims := &JWTClaims{}
	assert.NoError(t, claims.Valid())
}
