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

package daemon

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/hibiken/asynq"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/go-sigma/sigma/pkg/logger"
	"github.com/go-sigma/sigma/pkg/types/enums"
)

func TestRegisterTask(t *testing.T) {
	logger.SetLevel("debug")

	tasks = map[enums.Daemon]func(context.Context, *asynq.Task) error{}

	err := RegisterTask(enums.DaemonSbom, nil)
	assert.NoError(t, err)

	err = RegisterTask(enums.DaemonSbom, nil)
	assert.Error(t, err)
}

func TestInitializeServer(t *testing.T) {
	logger.SetLevel("debug")

	tasks = map[enums.Daemon]func(context.Context, *asynq.Task) error{}

	miniRedis := miniredis.RunT(t)
	viper.SetDefault("redis.url", "redis://"+miniRedis.Addr())
	viper.SetDefault("daemon.gc.cron", "0 2 * * 6")

	err := InitializeServer()
	assert.NoError(t, err)

	time.Sleep(1 * time.Second)

	DeinitServer()
}

func TestInitializeClient(t *testing.T) {
	logger.SetLevel("debug")

	tasks = map[enums.Daemon]func(context.Context, *asynq.Task) error{}

	miniRedis := miniredis.RunT(t)
	viper.SetDefault("redis.url", "redis://"+miniRedis.Addr())

	err := InitializeClient()
	assert.NoError(t, err)

	time.Sleep(1 * time.Second)

	err = DeinitClient()
	assert.NoError(t, err)
}
