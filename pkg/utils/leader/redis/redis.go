package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/ximager/ximager/pkg/utils/leader"
)

const (
	name             = "redis"
	leaderElectorKey = "ximager:leader"
)

var (
	luaRelease = redis.NewScript(`if redis.call("get", KEYS[1]) == ARGV[1] then return redis.call("del", KEYS[1]) else return 0 end`)
	luaRenew   = redis.NewScript(`if redis.call('get', KEYS[1]) == ARGV[1] then return redis.call('expire', KEYS[1], ARGV[2]) else return 0 end`)
)

func init() {
	err := leader.RegisterLeaderFactory(name, &factory{})
	if err != nil {
		panic(fmt.Sprintf("fail to register leader factory: %v", err))
	}
}

type factory struct{}

var _ leader.Factory = &factory{}

type redisLeaderElector struct {
	isLeader bool
}

// New ...
func (f factory) New(ctx context.Context, opts leader.Options) (leader.LeaderElector, error) {
	redisOpt, err := redis.ParseURL(viper.GetString("redis.url"))
	if err != nil {
		return nil, fmt.Errorf("redis.ParseURL error: %v", err)
	}
	redisCli := redis.NewClient(redisOpt)

	// isLeader := false
	l := &redisLeaderElector{isLeader: false}

	go func() {
		renewTicker := time.NewTicker(opts.RenewDeadline)
		defer renewTicker.Stop()

		periodTicker := time.NewTicker(opts.RetryPeriod)
		defer periodTicker.Stop()

	begin:

		l.isLeader = false

		lockContent := fmt.Sprintf("%d", time.Now().Unix())

		for {
			select {
			case <-ctx.Done():
				return
			case <-periodTicker.C:
				ok, err := redisCli.SetNX(ctx, leaderElectorKey, lockContent, opts.LeaseDuration).Result()
				if err != nil {
					log.Error().Err(err).Str("name", opts.Name).Msg("leader election failed to set redis key")
					continue
				}
				if ok {
					l.isLeader = true
					log.Info().Str("name", opts.Name).Msg("leader election succeeded")
				}
				log.Debug().Str("name", opts.Name).Msg("leader election failed")
			}
			if l.isLeader {
				break
			}
		}
		for {
			select {
			case <-ctx.Done():
				if !l.isLeader {
					return
				}
				res, err := luaRelease.Run(context.Background(), redisCli, []string{leaderElectorKey}, lockContent).Result()
				if err == redis.Nil {
					log.Error().Str("name", opts.Name).Msg("leader election failed to release: lock not held")
					return
				} else if err != nil {
					log.Error().Err(err).Str("name", opts.Name).Msg("leader election failed to release")
					return
				}
				if i, ok := res.(int64); !ok || i != 1 {
					log.Error().Str("name", opts.Name).Msg("leader election failed to release: lock not held")
					return
				}
				log.Debug().Str("name", opts.Name).Msg("leader election released")
				return
			case <-renewTicker.C:
				res, err := luaRenew.Run(ctx, redisCli, []string{leaderElectorKey},
					lockContent, int64(opts.LeaseDuration/time.Second)).Result()
				if err == redis.Nil {
					log.Error().Str("name", opts.Name).Msg("leader election failed to renew redis key, lock not held")
					goto begin
				}
				if err != nil {
					log.Error().Err(err).Msg("leader election failed to renew redis key")
					continue
				}
				if i, ok := res.(int64); !ok || i != 1 {
					log.Error().Str("name", opts.Name).Msg("leader election failed to renew redis key, lock not held")
					goto begin
				}
				log.Debug().Str("name", opts.Name).Msg("leader election renewed")
			}
		}
	}()

	return l, nil
}

// IsLeader returns whether the current pod is the leader
func (l redisLeaderElector) IsLeader() bool {
	return l.isLeader
}
