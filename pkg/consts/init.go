package consts

import (
	"time"

	"github.com/spf13/viper"
)

// Initialize inject the default values for the consts
func Initialize() {
	viper.SetDefault("auth.jwt.type", "RS256")            // the jwt token type
	viper.SetDefault("auth.jwt.ttl", time.Hour)           // the jwt token ttl
	viper.SetDefault("auth.jwt.refreshTtl", time.Hour*24) // the refresh token ttl
}
