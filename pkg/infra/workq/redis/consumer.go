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

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/infra/workq"
	"github.com/go-sigma/sigma/pkg/logger"
	"github.com/go-sigma/sigma/pkg/utils"
)

func init() {
	utils.PanicIf(workq.RegisterConsumer(enums.WorkQueueTypeRedis, &consumerFactory{}))
}

type consumerFactory struct{}

var _ workq.ConsumerFactory = consumerFactory{}

// New ...
func (consumerFactory) New(params workq.ConsumerParams) error {
	config := params.Config
	if !config.Redis.Enabled() {
		return fmt.Errorf("work queue: please check redis configuration, it should be configured")
	}
	if params.RedisClientFactory == nil {
		return fmt.Errorf("work queue: redis client factory is required")
	}
	redisCli, err := params.RedisClientFactory()
	if err != nil {
		return err
	}
	if redisCli == nil {
		return fmt.Errorf("work queue: redis client is required")
	}
	asyncSrv := asynq.NewServer(
		&makeClient{redisCli: redisCli},
		asynq.Config{
			Concurrency: config.WorkQueue.Redis.Concurrency,
			Logger:      &logger.Logger{},
		},
	)
	mux := asynq.NewServeMux()
	for topic, handler := range params.HandlerRegistry.Handlers() {
		mux.HandleFunc(topic.String(), func(consumer workq.Consumer) func(context.Context, *asynq.Task) error {
			return func(ctx context.Context, task *asynq.Task) error {
				ctx, payload := workq.UnmarshalPayload(ctx, task.Payload())
				ctx, span := workq.StartConsumerSpan(ctx, topic)
				defer span.End()
				err := consumer.Handler(ctx, payload)
				if err != nil {
					span.RecordError(err)
				}
				return err
			}
		}(handler))
	}

	go func() {
		err := asyncSrv.Run(mux)
		if err != nil {
			logger.Fatal("srv run failed", "err", err)
		}
	}()

	return nil
}
