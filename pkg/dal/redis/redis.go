package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/types/enums"
)

// Client ...
var Client redis.UniversalClient

// Initialize init redis
func Initialize(ctx context.Context, config configs.Configuration) error {
	if config.Redis.Type == enums.RedisTypeNone {
		return nil
	}
	redisOpt, err := redis.ParseURL(config.Redis.Url)
	if err != nil {
		return fmt.Errorf("redis.ParseURL error: %v", err)
	}
	redisCli := redis.NewClient(redisOpt)
	res, err := redisCli.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("redis ping error: %v", err)
	}
	if res != "PONG" {
		return fmt.Errorf("redis ping should got PONG, real: %s", res)
	}
	Client = redisCli
	return nil
}
