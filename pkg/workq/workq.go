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

package workq

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-sigma/sigma/pkg/configs"
)

// Message ...
type Message struct {
	Topic   string
	Payload []byte
}

type Consumer struct {
	Handler     func(ctx context.Context, payload []byte) error
	Concurrency int
	MaxRetry    int
	Timeout     time.Duration
}

var TopicConsumers = make(map[string]Consumer)

// WorkQueueProducer ...
type WorkQueueProducer interface {
	// Produce ...
	Produce(ctx context.Context, topic string, payload any) error
}

// WorkQueueConsumer ...
type WorkQueueConsumer interface {
	// Consume ...
	Run(ctx context.Context)
}

// ConsumerClientFactory ...
type ConsumerClientFactory interface {
	New(config configs.Configuration) error
}

// ProducerClientFactory ...
type ProducerClientFactory interface {
	New(config configs.Configuration) (WorkQueueProducer, error)
}

// ConsumerClientFactories ...
var ConsumerClientFactories = make(map[string]ConsumerClientFactory, 5)

// ProducerClientFactories ...
var ProducerClientFactories = make(map[string]ProducerClientFactory, 5)

// Initialize ...
func Initialize(config configs.Configuration) error {
	consumerClientFactory, ok := ConsumerClientFactories[strings.ToLower(config.WorkQueue.Type.String())]
	if !ok {
		return fmt.Errorf("Work queue consumer(%s) not support", config.WorkQueue.Type.String())
	}
	err := consumerClientFactory.New(config)
	if err != nil {
		return err
	}
	producerClientFactory, ok := ProducerClientFactories[strings.ToLower(config.WorkQueue.Type.String())]
	if !ok {
		return fmt.Errorf("Work queue producer(%s) not support", config.WorkQueue.Type.String())
	}
	producer, err := producerClientFactory.New(config)
	if err != nil {
		return err
	}
	ProducerClient = producer
	return nil
}

// ProducerClient ...
var ProducerClient WorkQueueProducer
