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
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/modules/cacher"
	"github.com/go-sigma/sigma/pkg/modules/cacher/definition"
	"github.com/go-sigma/sigma/pkg/utils"
)

//go:generate mockgen -destination=mocks/token.go -package=mocks github.com/go-sigma/sigma/pkg/utils/token Service

const (
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

// Service is the interface for token service.
type Service interface {
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
	cacheCli   definition.Cacher[string]
}

// New creates a new token service.
func New(digCon *dig.Container) (Service, error) {
	config := utils.MustGetObjFromDigCon[configs.Configuration](digCon)
	privateKeyBytes, err := base64.StdEncoding.DecodeString(config.Auth.Jwt.PrivateKey)
	if err != nil {
		return nil, err
	}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyBytes)
	if err != nil {
		return nil, err
	}
	publicKey := &privateKey.PublicKey
	cacheCli, err := cacher.New[string](digCon, consts.AppName+":expire:jwt", nil)
	if err != nil {
		return nil, fmt.Errorf("new cacher failed: %v", err)
	}
	return &tokenService{
		privateKey: privateKey,
		publicKey:  publicKey,
		cacheCli:   cacheCli,
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

	val, err := s.cacheCli.Get(ctx, id)
	if err != nil && !errors.Is(err, definition.ErrNotFound) {
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
	err := s.cacheCli.Set(ctx, id, expireVal, time.Second*3600)
	if err != nil {
		return err
	}
	return nil
}
