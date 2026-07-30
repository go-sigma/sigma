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
	"crypto"
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/consts"
	dalredis "github.com/go-sigma/sigma/pkg/dal/redis"
	cacher "github.com/go-sigma/sigma/pkg/infra/cache"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

//go:generate mockgen -destination=token_mocks.go -package=token github.com/go-sigma/sigma/pkg/service/token Service

const (
	expireVal = "1"
)

var (
	// ErrRevoked token has been revoked.
	ErrRevoked = fmt.Errorf("token has been revoked")
)

// JWTClaims is the claims for the JWT token.
type JWTClaims struct {
	jwt.RegisteredClaims

	UID string `json:"uid"`
}

// Valid validates the claims.
func (j JWTClaims) Valid() error {
	return nil
}

// Service is the interface for token service.
type Service interface {
	// New creates a new token.
	New(id string, expire time.Duration) (string, error)
	// Validate validates the token.
	Validate(ctx context.Context, token string) (string, string, error)
	// Revoke revokes the token.
	Revoke(ctx context.Context, id string) error
}

type service struct {
	dig.In `ignore-unexported:"true"`

	Config             *config.Configuration
	RedisClientFactory dalredis.ClientFactory `optional:"true"`

	privateKey crypto.PrivateKey
	publicKey  crypto.PublicKey
	cacheCli   cacher.Cacher[string]
}

// NewService creates the token service.
func NewService(params service) (Service, error) {
	privateKeyBytes, err := base64.StdEncoding.DecodeString(params.Config.Auth.Jwt.PrivateKey)
	if err != nil {
		return nil, err
	}
	privateKey, err := jwt.ParseEdPrivateKeyFromPEM(privateKeyBytes)
	if err != nil {
		return nil, err
	}
	publicKey, ok := privateKey.(ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("invalid private key")
	}

	cacheCli, err := cacher.New[string](cacher.Params{
		Config:             params.Config,
		RedisClientFactory: params.RedisClientFactory,
	}, consts.AppName+":expire:jwt", nil)
	if err != nil {
		return nil, fmt.Errorf("new cacher failed: %v", err)
	}
	params.privateKey = privateKey
	params.publicKey = publicKey.Public()
	params.cacheCli = cacheCli
	return &params, nil
}

// NewWithConfig creates a token service outside dig.
func NewWithConfig(configuration *config.Configuration, redisClientFactory dalredis.ClientFactory) (Service, error) {
	return NewService(service{
		Config:             configuration,
		RedisClientFactory: redisClientFactory,
	})
}

// New creates a new token.
func (s *service) New(id string, expire time.Duration) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   id,
			Issuer:    consts.AppName,
			ExpiresAt: jwt.NewNumericDate(now.Add(expire)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.NewV7String(),
		},
		UID: id,
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims).SignedString(s.privateKey)
	if err != nil {
		return "", err
	}
	return token, nil
}

// Validate validates the token.
func (s *service) Validate(ctx context.Context, token string) (string, string, error) {
	jwtToken, err := jwt.Parse(token, func(token *jwt.Token) (any, error) {
		return s.publicKey, nil
	})
	if err != nil {
		return "", "", err
	}
	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok || !jwtToken.Valid {
		return "", "", fmt.Errorf("invalid token")
	}
	result, ok := claims["uid"].(string)
	if !ok {
		return "", "", fmt.Errorf("invalid token")
	}
	id, ok := claims["jti"].(string)
	if !ok {
		return "", "", fmt.Errorf("invalid token")
	}

	val, err := s.cacheCli.Get(ctx, id)
	if err != nil && !errors.Is(err, cacher.ErrNotFound) {
		return "", "", err
	}
	if val == expireVal {
		return "", "", ErrRevoked
	}

	return id, result, nil
}

// Revoke revokes the token.
func (s *service) Revoke(ctx context.Context, id string) error {
	err := s.cacheCli.Set(ctx, id, expireVal, time.Second*3600)
	if err != nil {
		return err
	}
	return nil
}
