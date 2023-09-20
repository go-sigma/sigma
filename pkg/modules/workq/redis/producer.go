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
	"time"

	"github.com/hibiken/asynq"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/modules/workq/definition"
	"github.com/go-sigma/sigma/pkg/utils"
)

type producer struct {
	client        *asynq.Client
	topicHandlers map[string]definition.Consumer
}

// NewWorkQueueProducer ...
func NewWorkQueueProducer(config configs.Configuration, topicHandlers map[string]definition.Consumer) (definition.WorkQueueProducer, error) {
	redisOpt, err := asynq.ParseRedisURI(config.Redis.Url)
	if err != nil {
		return nil, fmt.Errorf("asynq.ParseRedisURI error: %v", err)
	}
	p := &producer{
		client:        asynq.NewClient(redisOpt),
		topicHandlers: topicHandlers,
	}
	return p, nil
}

// Produce ...
func (p *producer) Produce(ctx context.Context, topic string, payload any) error {
	consumer, ok := p.topicHandlers[topic]
	if !ok {
		return fmt.Errorf("Topic %s not registered", topic)
	}
	var opts []asynq.Option
	if consumer.MaxRetry > 0 {
		opts = append(opts, asynq.MaxRetry(consumer.MaxRetry))
	} else {
		opts = append(opts, asynq.MaxRetry(1))
	}
	if consumer.Timeout > 0 {
		opts = append(opts, asynq.Timeout(consumer.Timeout))
	} else {
		opts = append(opts, asynq.Timeout(time.Hour))
	}
	_, err := p.client.Enqueue(asynq.NewTask(topic, utils.MustMarshal(payload)), opts...)
	return err
}
