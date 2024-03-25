package redis

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/types/enums"
)

func TestRedis(t *testing.T) {
	err := Initialize(context.Background(), configs.Configuration{
		Redis: configs.ConfigurationRedis{
			Type: enums.RedisTypeNone,
			Url:  "",
		},
	})
	assert.NoError(t, err)
	assert.Nil(t, Client)

	err = Initialize(context.Background(), configs.Configuration{
		Redis: configs.ConfigurationRedis{
			Type: enums.RedisTypeExternal,
			Url:  miniredis.RunT(t).Addr(),
		},
	})
	assert.Error(t, err)

	err = Initialize(context.Background(), configs.Configuration{
		Redis: configs.ConfigurationRedis{
			Type: enums.RedisTypeExternal,
			Url:  "redis://" + miniredis.RunT(t).Addr(),
		},
	})
	assert.NoError(t, err)
}
