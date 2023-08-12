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
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"

	"github.com/go-sigma/sigma/pkg/consts"
)

//go:generate mockgen -destination=mocks/token.go -package=mocks github.com/go-sigma/sigma/pkg/utils/token TokenService

const (
	expireKey = consts.AppName + ":expire:jwt:%s"
	expireVal = "1"
)

var (
	// ErrRevoked token has been revoked
	ErrRevoked = fmt.Errorf("token has been revoked")
)

// JWTClaims is the claims for the JWT token
type JWTClaims struct {
	jwt.RegisteredClaims

	UID string `json:"uid"`
}

// Valid validates the claims
func (j JWTClaims) Valid() error {
	return nil
}

// TokenService is the interface for token service.
type TokenService interface {
	// New creates a new token.
	New(id int64, expire time.Duration) (string, error)
	// Validate validates the token.
	Validate(ctx context.Context, token string) (string, int64, error)
	// Revoke revokes the token.
	Revoke(ctx context.Context, id string) error
}

type tokenService struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	redisCli   redis.UniversalClient
}

// NewTokenService creates a new token service.
func NewTokenService(privateKeyString string) (TokenService, error) {
	privateKeyBytes, err := base64.StdEncoding.DecodeString(privateKeyString)
	if err != nil {
		return nil, err
	}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyBytes)
	if err != nil {
		return nil, err
	}
	publicKey := &privateKey.PublicKey
	redisOpt, err := redis.ParseURL(viper.GetString("redis.url"))
	if err != nil {
		return nil, fmt.Errorf("redis.ParseURL error: %v", err)
	}
	redisCli := redis.NewClient(redisOpt)
	return &tokenService{
		privateKey: privateKey,
		publicKey:  publicKey,
		redisCli:   redisCli,
	}, nil
}

// New creates a new token.
func (s *tokenService) New(id int64, expire time.Duration) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(id, 10),
			Issuer:    consts.AppName,
			ExpiresAt: jwt.NewNumericDate(now.Add(expire)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
		UID: strconv.FormatInt(id, 10),
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodRS512, claims).SignedString(s.privateKey)
	if err != nil {
		return "", err
	}
	return token, nil
}

// Validate validates the token.
func (s *tokenService) Validate(ctx context.Context, token string) (string, int64, error) {
	jwtToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return s.publicKey, nil
	})
	if err != nil {
		return "", 0, err
	}
	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok || !jwtToken.Valid {
		return "", 0, fmt.Errorf("invalid token")
	}
	result, ok := claims["uid"].(string)
	if !ok {
		return "", 0, fmt.Errorf("invalid token")
	}
	id, ok := claims["jti"].(string)
	if !ok {
		return "", 0, fmt.Errorf("invalid token")
	}

	val, err := s.redisCli.Get(ctx, fmt.Sprintf(expireKey, id)).Result()
	if err != nil && err != redis.Nil {
		return "", 0, err
	}
	if val == expireVal {
		return "", 0, ErrRevoked
	}
	ret, err := strconv.ParseInt(result, 10, 0)
	if err != nil {
		return "", 0, fmt.Errorf("invalid token, parse uid(%s) failed: %v", result, err)
	}

	return id, ret, nil
}

// Revoke revokes the token.
func (s *tokenService) Revoke(ctx context.Context, id string) error {
	_, err := s.redisCli.Set(ctx, fmt.Sprintf(expireKey, id), expireVal, viper.GetDuration("auth.jwt.refreshTtl")).Result()
	if err != nil {
		return err
	}
	return nil
}
