package types

import "github.com/golang-jwt/jwt/v4"

// JWTClaims is the claims for the JWT token
type JWTClaims struct {
	jwt.RegisteredClaims

	Name string `json:"name"`
}

func (j JWTClaims) Valid() error {
	return nil
}
