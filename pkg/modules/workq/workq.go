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
	"fmt"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/modules/workq/database"
	"github.com/go-sigma/sigma/pkg/modules/workq/definition"
	"github.com/go-sigma/sigma/pkg/modules/workq/inmemory"
	"github.com/go-sigma/sigma/pkg/modules/workq/kafka"
	"github.com/go-sigma/sigma/pkg/modules/workq/redis"
	"github.com/go-sigma/sigma/pkg/types/enums"
)

// Message ...
type Message struct {
	Topic   string
	Payload []byte
}

var TopicHandlers = make(map[enums.Daemon]definition.Consumer)

// ProducerClient ...
var ProducerClient definition.WorkQueueProducer

// InitProducer ...
func InitProducer(config configs.Configuration) error {
	var err error
	switch config.WorkQueue.Type {
	case enums.WorkQueueTypeDatabase:
		ProducerClient, err = database.NewWorkQueueProducer(config, TopicHandlers)
	case enums.WorkQueueTypeKafka:
		ProducerClient, err = kafka.NewWorkQueueProducer(config, TopicHandlers)
	case enums.WorkQueueTypeRedis:
		ProducerClient, err = redis.NewWorkQueueProducer(config, TopicHandlers)
	case enums.WorkQueueTypeInmemory:
		ProducerClient, err = inmemory.NewWorkQueueProducer(config, TopicHandlers)
	default:
		return fmt.Errorf("Workq %s not support", config.WorkQueue.Type.String())
	}
	if err != nil {
		return err
	}
	return nil
}

// Initialize ...
func Initialize(config configs.Configuration) error {
	var err error
	switch config.WorkQueue.Type {
	case enums.WorkQueueTypeDatabase:
		err = database.NewWorkQueueConsumer(config, TopicHandlers)
	case enums.WorkQueueTypeKafka:
		err = kafka.NewWorkQueueConsumer(config, TopicHandlers)
	case enums.WorkQueueTypeRedis:
		err = redis.NewWorkQueueConsumer(config, TopicHandlers)
	case enums.WorkQueueTypeInmemory:
		err = inmemory.NewWorkQueueConsumer(config, TopicHandlers)
	default:
		return fmt.Errorf("Workq %s not support", config.WorkQueue.Type.String())
	}
	if err != nil {
		return err
	}
	switch config.WorkQueue.Type {
	case enums.WorkQueueTypeDatabase:
		ProducerClient, err = database.NewWorkQueueProducer(config, TopicHandlers)
	case enums.WorkQueueTypeKafka:
		ProducerClient, err = kafka.NewWorkQueueProducer(config, TopicHandlers)
	case enums.WorkQueueTypeRedis:
		ProducerClient, err = redis.NewWorkQueueProducer(config, TopicHandlers)
	case enums.WorkQueueTypeInmemory:
		ProducerClient, err = inmemory.NewWorkQueueProducer(config, TopicHandlers)
	default:
		return fmt.Errorf("Workq %s not support", config.WorkQueue.Type.String())
	}
	if err != nil {
		return err
	}
	return nil
}
