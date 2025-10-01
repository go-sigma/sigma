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
	"time"

	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
)

//go:generate mockgen -destination=workq_mocks.go -package=workq github.com/go-sigma/sigma/pkg/modules/workq Producer,ProducerFactory,ConsumerFactory

// Consumer ...
type Consumer struct {
	Handler     func(ctx context.Context, payload []byte) error
	Concurrency int
	MaxRetry    int
	Timeout     time.Duration
}

// ProducerOption ...
type ProducerOption struct {
	Tx *query.Query
}

// Producer ...
type Producer interface {
	// Produce ...
	Produce(ctx context.Context, topic enums.Daemon, payload any, option ProducerOption) error
}

// ProducerFactory is the interface for the producer factory
type ProducerFactory interface {
	New(digCon *dig.Container) (Producer, error)
}

// ConsumerFactory is the interface for the consumer factory
type ConsumerFactory interface {
	New(digCon *dig.Container, topicHandlers map[enums.Daemon]Consumer) error
}

// Message ...
type Message struct {
	Topic   string
	Payload []byte
}

// TopicHandlers ...
var TopicHandlers = make(map[enums.Daemon]Consumer)

var producerFactories = make(map[enums.WorkQueueType]ProducerFactory)

var consumerFactories = make(map[enums.WorkQueueType]ConsumerFactory)

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
func InitProducer(digCon *dig.Container) (Producer, error) {
	config := utils.MustGetObjFromDigCon[configs.Configuration](digCon)
	factory, ok := producerFactories[config.WorkQueue.Type]
	if !ok {
		return nil, fmt.Errorf("workq %q not support", config.WorkQueue.Type.String())
	}
	return factory.New(digCon)
}

// InitConsumer ...
func InitConsumer(digCon *dig.Container, topicHandlers map[enums.Daemon]Consumer) error {
	config := utils.MustGetObjFromDigCon[configs.Configuration](digCon)
	factory, ok := consumerFactories[config.WorkQueue.Type]
	if !ok {
		return fmt.Errorf("workq %q not support", config.WorkQueue.Type.String())
	}
	return factory.New(digCon, topicHandlers)
}
