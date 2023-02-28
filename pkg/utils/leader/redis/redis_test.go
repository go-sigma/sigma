package redis

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/ximager/ximager/pkg/utils"
	"github.com/ximager/ximager/pkg/utils/leader"
)

func TestNew(t *testing.T) {
	ctx, ctxCancel := context.WithCancel(context.Background())

	utils.SetLevel(0)

	miniRedis := miniredis.RunT(t)
	viper.SetDefault("redis.url", "redis://"+miniRedis.Addr())

	var f = factory{}
	_, err := f.New(ctx, leader.Options{
		Name:          "leader",
		LeaseDuration: time.Second * 15,
		RenewDeadline: time.Second * 3,
		RetryPeriod:   time.Second * 2,
	})
	assert.NoError(t, err)

	time.Sleep(time.Second * 3)

	ctxCancel()

	time.Sleep(time.Second * 5)
}

func TestLeaderChange(t *testing.T) {
	utils.SetLevel(0)

	miniRedis := miniredis.RunT(t)
	viper.SetDefault("redis.url", "redis://"+miniRedis.Addr())

	ctx1, ctxCancel1 := context.WithCancel(context.Background())
	var f = factory{}
	_, err := f.New(ctx1, leader.Options{
		Name:          "leader1",
		LeaseDuration: time.Second * 15,
		RenewDeadline: time.Second * 3,
		RetryPeriod:   time.Second * 2,
	})
	assert.NoError(t, err)

	time.Sleep(time.Second * 1)

	var f1 = factory{}
	leader1, err := f1.New(context.Background(), leader.Options{
		Name:          "leader2",
		LeaseDuration: time.Second * 15,
		RenewDeadline: time.Second * 3,
		RetryPeriod:   time.Second * 2,
	})
	assert.NoError(t, err)

	time.Sleep(time.Second * 3)

	ctxCancel1()

	time.Sleep(time.Second * 3)

	assert.True(t, leader1.IsLeader())

	time.Sleep(time.Second * 5)
}
