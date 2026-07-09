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
	"github.com/go-sigma/sigma/pkg/utils"
)

const defaultQueueName = "default"

func init() {
	utils.PanicIf(workq.RegisterProducer(enums.WorkQueueTypeRedis, &producerFactory{}))
}

type queueInspector interface {
	GetQueueInfo(queue string) (*asynq.QueueInfo, error)
}

type producer struct {
	client     *asynq.Client
	inspector  queueInspector
	maxBacklog int
}

type producerFactory struct{}

var _ workq.ProducerFactory = producerFactory{}

// New ...
func (producerFactory) New(params workq.ProducerParams) (workq.Producer, error) {
	config := params.Config
	if !config.Redis.Enabled() {
		return nil, fmt.Errorf("work queue: please check redis configuration, it should be configured")
	}
	if params.RedisClientFactory == nil {
		return nil, fmt.Errorf("work queue: redis client factory is required")
	}
	redisCli, err := params.RedisClientFactory()
	if err != nil {
		return nil, err
	}
	if redisCli == nil {
		return nil, fmt.Errorf("work queue: redis client is required")
	}
	redisClient := &makeClient{redisCli: redisCli}
	p := &producer{
		client:     asynq.NewClient(redisClient),
		inspector:  asynq.NewInspector(redisClient),
		maxBacklog: config.WorkQueue.Redis.MaxBacklog,
	}
	return p, nil
}

// Produce ...
func (p *producer) Produce(ctx context.Context, topic enums.Daemon, payload any) error {
	if err := p.ensureBacklogAvailable(); err != nil {
		return err
	}
	data, err := workq.MarshalPayload(ctx, payload)
	if err != nil {
		return fmt.Errorf("marshal work queue payload: %w", err)
	}
	_, err = p.client.EnqueueContext(ctx, asynq.NewTask(topic.String(), data))
	return err
}

func (p *producer) ensureBacklogAvailable() error {
	if p.maxBacklog <= 0 {
		return nil
	}
	queueInfo, err := p.inspector.GetQueueInfo(defaultQueueName)
	if err != nil {
		return fmt.Errorf("inspect work queue backlog: %w", err)
	}
	backlog := queueBacklog(queueInfo)
	if backlog >= p.maxBacklog {
		return fmt.Errorf("work queue backlog limit reached: queue=%s backlog=%d maxBacklog=%d", defaultQueueName, backlog, p.maxBacklog)
	}
	return nil
}

func queueBacklog(queueInfo *asynq.QueueInfo) int {
	if queueInfo == nil {
		return 0
	}
	return queueInfo.Pending + queueInfo.Scheduled + queueInfo.Retry + queueInfo.Aggregating
}
