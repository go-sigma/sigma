package user

import (
	"encoding/base64"
	"time"

	"github.com/golang-jwt/jwt/v4"
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
		},
		Name: "ximager",
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
