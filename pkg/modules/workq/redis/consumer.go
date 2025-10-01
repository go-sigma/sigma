// Copyright 2023 sigma
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

package redis

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/logger"
	"github.com/go-sigma/sigma/pkg/modules/workq"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
)

func init() {
	utils.PanicIf(workq.RegisterConsumer(enums.WorkQueueTypeRedis, &consumerFactory{}))
}

type consumerFactory struct{}

var _ workq.ConsumerFactory = consumerFactory{}

// New ...
func (consumerFactory) New(digCon *dig.Container, topicHandlers map[enums.Daemon]workq.Consumer) error {
	config := utils.MustGetObjFromDigCon[configs.Configuration](digCon)
	if config.Redis.Type != enums.RedisTypeExternal {
		return fmt.Errorf("work queue: please check redis configuration, it should be external")
	}
	redisCli := utils.MustGetObjFromDigCon[redis.UniversalClient](digCon)
	asyncSrv := asynq.NewServer(
		&makeClient{redisCli: redisCli},
		asynq.Config{
			Concurrency: config.WorkQueue.Redis.Concurrency,
			Logger:      &logger.Logger{},
		},
	)
	mux := asynq.NewServeMux()
	for topic, handler := range topicHandlers {
		mux.HandleFunc(topic.String(), func(consumer workq.Consumer) func(context.Context, *asynq.Task) error {
			return func(ctx context.Context, task *asynq.Task) error {
				return consumer.Handler(ctx, task.Payload())
			}
		}(handler))
	}

	go func() {
		err := asyncSrv.Run(mux)
		if err != nil {
			log.Fatal().Err(err).Msg("srv run failed")
		}
	}()

	return nil
}
