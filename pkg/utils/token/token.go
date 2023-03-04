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
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"

	"github.com/ximager/ximager/pkg/consts"
)

const (
	expireKey = "ximager:expire:jwt:%s"
	expireVal = "1"
)

// JWTClaims is the claims for the JWT token
type JWTClaims struct {
	jwt.RegisteredClaims

	Username string `json:"dat"`
}

// Valid validates the claims
func (j JWTClaims) Valid() error {
	return nil
}

// TokenService is the interface for token service.
type TokenService interface {
	// New creates a new token.
	New(username string) (string, error)
	// Validate validates the token.
	Validate(ctx context.Context, token string) (string, string, error)
	// Revoke revokes the token.
	Revoke(ctx context.Context, id string) error
}

type tokenService struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	ttl        time.Duration
	redisCli   redis.UniversalClient
}

// NewTokenService creates a new token service.
func NewTokenService(privateKeyString, publicKeyString string) (TokenService, error) {
	privateKeyBytes, err := base64.StdEncoding.DecodeString(privateKeyString)
	if err != nil {
		return nil, err
	}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyBytes)
	if err != nil {
		return nil, err
	}
	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKeyString)
	if err != nil {
		return nil, err
	}
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyBytes)
	if err != nil {
		return nil, err
	}
	redisOpt, err := redis.ParseURL(viper.GetString("redis.url"))
	if err != nil {
		return nil, fmt.Errorf("redis.ParseURL error: %v", err)
	}
	redisCli := redis.NewClient(redisOpt)
	return &tokenService{
		privateKey: privateKey,
		publicKey:  publicKey,
		redisCli:   redisCli,
		ttl:        viper.GetDuration("admin.jwt.expire"),
	}, nil
}

// New creates a new token.
func (s *tokenService) New(username string) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    consts.AppName,
			ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
		Username: username,
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodRS512, claims).SignedString(s.privateKey)
	if err != nil {
		return "", err
	}
	return token, nil
}

// Validate validates the token.
func (s *tokenService) Validate(ctx context.Context, token string) (string, string, error) {
	jwtToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return s.publicKey, nil
	})
	if err != nil {
		return "", "", err
	}
	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok || !jwtToken.Valid {
		return "", "", fmt.Errorf("invalid token")
	}
	result, ok := claims["dat"].(string)
	if !ok {
		return "", "", fmt.Errorf("invalid token")
	}
	id, ok := claims["jti"].(string)
	if !ok {
		return "", "", fmt.Errorf("invalid token")
	}

	val, err := s.redisCli.Get(context.Background(), fmt.Sprintf(expireKey, id)).Result()
	if err != nil && err != redis.Nil {
		return "", "", err
	}
	if val == expireVal {
		return "", "", fmt.Errorf("token has been revoked")
	}

	return id, result, nil
}

// Revoke revokes the token.
func (s *tokenService) Revoke(ctx context.Context, id string) error {
	_, err := s.redisCli.Set(ctx, fmt.Sprintf(expireKey, id), expireVal, viper.GetDuration("admin.jwt.expire")).Result()
	if err != nil {
		return err
	}
	return nil
}
