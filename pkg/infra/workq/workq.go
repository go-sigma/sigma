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
	"sync"
	"time"

	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/config"
	dalredis "github.com/go-sigma/sigma/pkg/dal/redis"
)

//go:generate go tool mockgen -destination=workq_mocks.go -package=workq github.com/go-sigma/sigma/pkg/infra/workq Producer,ProducerFactory,ConsumerFactory

// Consumer ...
type Consumer struct {
	Handler     func(ctx context.Context, payload []byte) error
	Concurrency int
	Timeout     time.Duration
}

// Producer ...
type Producer interface {
	// Produce ...
	Produce(ctx context.Context, topic enums.Daemon, payload any) error
}

// ProducerFactory is the interface for the producer factory
type ProducerFactory interface {
	New(params ProducerParams) (Producer, error)
}

// ConsumerFactory is the interface for the consumer factory
type ConsumerFactory interface {
	New(params ConsumerParams) error
}

// HandlerRegistry stores work queue topic handlers.
type HandlerRegistry interface {
	Register(topic enums.Daemon, consumer Consumer) error
	Handlers() map[enums.Daemon]Consumer
}

// ProducerParams declares dependencies needed to construct a producer.
type ProducerParams struct {
	dig.In

	Config             *config.Configuration
	RedisClientFactory dalredis.ClientFactory `optional:"true"`
}

// ConsumerParams declares dependencies needed to construct consumers.
type ConsumerParams struct {
	dig.In

	Config             *config.Configuration
	RedisClientFactory dalredis.ClientFactory `optional:"true"`
	HandlerRegistry    HandlerRegistry
}

// Message ...
type Message struct {
	Topic   string
	Payload []byte
}

var producerFactories = make(map[enums.WorkQueueType]ProducerFactory)

var consumerFactories = make(map[enums.WorkQueueType]ConsumerFactory)

type handlerRegistry struct {
	mu       sync.RWMutex
	handlers map[enums.Daemon]Consumer
}

// NewHandlerRegistry creates a work queue handler registry.
func NewHandlerRegistry() HandlerRegistry {
	return &handlerRegistry{
		handlers: make(map[enums.Daemon]Consumer),
	}
}

// Register registers a work queue consumer for the topic.
func (r *handlerRegistry) Register(topic enums.Daemon, consumer Consumer) error {
	if consumer.Handler == nil {
		return fmt.Errorf("work queue topic %q handler is nil", topic.String())
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.handlers[topic]; ok {
		return fmt.Errorf("work queue topic %q already registered", topic.String())
	}
	r.handlers[topic] = consumer
	return nil
}

// Handlers returns a snapshot of registered work queue consumers.
func (r *handlerRegistry) Handlers() map[enums.Daemon]Consumer {
	r.mu.RLock()
	defer r.mu.RUnlock()
	handlers := make(map[enums.Daemon]Consumer, len(r.handlers))
	for topic, consumer := range r.handlers {
		handlers[topic] = consumer
	}
	return handlers
}

// RegisterProducer registers a storage factory driver by name.
// If Register is called twice with the same name or if driver is nil, it panics.
func RegisterProducer(name enums.WorkQueueType, factory ProducerFactory) error {
	if _, ok := producerFactories[name]; ok {
		return fmt.Errorf("driver %q already registered", name)
	}
	producerFactories[name] = factory
	return nil
}

// RegisterConsumer registers a storage factory driver by name.
// If Register is called twice with the same name or if driver is nil, it panics.
func RegisterConsumer(name enums.WorkQueueType, factory ConsumerFactory) error {
	if _, ok := consumerFactories[name]; ok {
		return fmt.Errorf("driver %q already registered", name)
	}
	consumerFactories[name] = factory
	return nil
}

// InitProducer ...
func InitProducer(params ProducerParams) (Producer, error) {
	factory, ok := producerFactories[params.Config.WorkQueue.Type]
	if !ok {
		return nil, fmt.Errorf("workq %q not support", params.Config.WorkQueue.Type.String())
	}
	return factory.New(params)
}

// InitConsumer ...
func InitConsumer(params ConsumerParams) error {
	factory, ok := consumerFactories[params.Config.WorkQueue.Type]
	if !ok {
		return fmt.Errorf("workq %q not support", params.Config.WorkQueue.Type.String())
	}
	return factory.New(params)
}
