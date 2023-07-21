package configs

import (
	"fmt"

	"github.com/spf13/viper"
)

func init() {
	checkers = append(checkers, checkDeploy)
}

func checkDeploy() error {
	if viper.GetString("deploy") == "replica" {
		if viper.GetString("redis.type") == "internal" {
			return fmt.Errorf("Deploy replica should use external redis")
		}
	}
	return nil
}
