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
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

func newEd25519PrivateKeyString(t *testing.T) string {
	t.Helper()

	_, privateKey, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	privateKeyDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)

	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyDER})
	return base64.StdEncoding.EncodeToString(pemBytes)
}

func newRSAKeyString(t *testing.T) string {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	privateKeyDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)

	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyDER})
	return base64.StdEncoding.EncodeToString(pemBytes)
}

func TestJWTClaimsValid(t *testing.T) {
	claims := &JWTClaims{}
	assert.NoError(t, claims.Valid())
}

func TestNew(t *testing.T) {
	privateKeyString := newEd25519PrivateKeyString(t)
	privateInvalidKeyString := newRSAKeyString(t)

	tests := []struct {
		name      string
		newParams func(*testing.T) service
		want      Service
		wantErr   bool
	}{
		{
			name: "bad redis",
			newParams: func(t *testing.T) service {
				return service{
					Config: &config.Configuration{
						Redis: config.ConfigurationRedis{},
						Cache: config.ConfigurationCache{
							Type: enums.CacherTypeRedis,
						},
						Auth: config.ConfigurationAuth{
							Jwt: config.ConfigurationAuthJwt{
								PrivateKey: privateKeyString,
							},
						},
					},
				}
			},
			wantErr: true,
		},
		{
			name: "invalid key",
			newParams: func(t *testing.T) service {
				miniRedis := miniredis.RunT(t)
				return service{
					Config: &config.Configuration{
						Redis: config.ConfigurationRedis{
							URL: "redis://" + miniRedis.Addr(),
						},
						Cache: config.ConfigurationCache{
							Type: enums.CacherTypeRedis,
						},
						Auth: config.ConfigurationAuth{
							Jwt: config.ConfigurationAuthJwt{
								PrivateKey: privateInvalidKeyString,
							},
						},
					},
					RedisClientFactory: func() (redis.UniversalClient, error) {
						return redis.NewClient(&redis.Options{Addr: miniRedis.Addr()}), nil
					},
				}
			},
			wantErr: true,
		},
		{
			name: "bad key",
			newParams: func(t *testing.T) service {
				miniRedis := miniredis.RunT(t)
				return service{
					Config: &config.Configuration{
						Redis: config.ConfigurationRedis{

							URL: "redis://" + miniRedis.Addr(),
						},
						Cache: config.ConfigurationCache{
							Type: enums.CacherTypeRedis,
						},
						Auth: config.ConfigurationAuth{
							Jwt: config.ConfigurationAuthJwt{
								PrivateKey: privateKeyString + "-",
							},
						},
					},
					RedisClientFactory: func() (redis.UniversalClient, error) {
						return redis.NewClient(&redis.Options{Addr: miniRedis.Addr()}), nil
					},
				}
			},
			wantErr: true,
		},
		{
			name: "normal",
			newParams: func(t *testing.T) service {
				miniRedis := miniredis.RunT(t)
				return service{
					Config: &config.Configuration{
						Redis: config.ConfigurationRedis{

							URL: "redis://" + miniRedis.Addr(),
						},
						Cache: config.ConfigurationCache{
							Type: enums.CacherTypeRedis,
						},
						Auth: config.ConfigurationAuth{
							Jwt: config.ConfigurationAuthJwt{
								PrivateKey: privateKeyString,
							},
						},
					},
					RedisClientFactory: func() (redis.UniversalClient, error) {
						return redis.NewClient(&redis.Options{Addr: miniRedis.Addr()}), nil
					},
				}
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenSvc, err := newService(tt.newParams(t))
			if (err != nil) != tt.wantErr {
				t.Errorf("newService() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			token, err := tokenSvc.New("100", time.Second*30)
			assert.NoError(t, err)
			assert.NotEmpty(t, token)

			id, uid, err := tokenSvc.Validate(context.Background(), token)
			assert.NoError(t, err)
			assert.Equal(t, "100", uid)
			_, err = uuid.Parse(id)
			assert.NoError(t, err)

			err = tokenSvc.Revoke(context.Background(), id)
			assert.NoError(t, err)

			_, _, err = tokenSvc.Validate(context.Background(), token)
			assert.Error(t, err)
		})
	}
}
