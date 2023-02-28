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

package user

import (
	"encoding/base64"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/ximager/ximager/pkg/types"
	"github.com/ximager/ximager/pkg/xerrors"
)

// Login handles the login request
func (h *handlers) Login(c echo.Context) error {
	privateKeyBytes, err := base64.StdEncoding.DecodeString(viper.GetString("admin.jwt.privateKey"))
	if err != nil {
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyBytes)
	if err != nil {
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	now := time.Now()
	claims := types.JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "ximager",
			ExpiresAt: jwt.NewNumericDate(now.Add(viper.GetDuration("admin.jwt.expire"))),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
		Username: "ximager",
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(privateKey)
	if err != nil {
		log.Error().Err(err).Msg("failed to sign token")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	publicKeyBytes, err := base64.StdEncoding.DecodeString(viper.GetString("admin.jwt.publicKey"))
	if err != nil {
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyBytes)
	if err != nil {
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}
	tok, _ := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return publicKey, nil
	})
	if tok.Valid {
		return nil
	}
	return nil
}
