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

package inmemory

import (
	"context"

	"github.com/rs/zerolog/log"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/modules/workq"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
)

func init() {
	utils.PanicIf(workq.RegisterConsumer(enums.WorkQueueTypeInmemory, &consumerFactory{}))
}

type consumerFactory struct{}

var _ workq.ConsumerFactory = consumerFactory{}

// This is only for small-scale deployment, a message queue with 1024 messages should suffice, and it can be adjusted appropriately if necessary.
var packs = make(map[enums.Daemon]chan *models.WorkQueue, 10)

// New ...
func (consumerFactory) New(digCon *dig.Container, topicHandlers map[enums.Daemon]workq.Consumer) error {
	config := utils.MustGetObjFromDigCon[configs.Configuration](digCon)
	for topic := range topicHandlers {
		packs[topic] = make(chan *models.WorkQueue, config.WorkQueue.Inmemory.Concurrency)
	}
	for topic, c := range topicHandlers {
		go func(consumer workq.Consumer, topic enums.Daemon) {
			handler := &consumerHandler{
				processingSemaphore: make(chan struct{}, consumer.Concurrency),
				consumer:            consumer,
			}
			handler.Consume(topic)
		}(c, topic)
	}
	return nil
}

type consumerHandler struct {
	processingSemaphore chan struct{}
	consumer            workq.Consumer
}

func (h *consumerHandler) Consume(topic enums.Daemon) {
	for {
		h.processingSemaphore <- struct{}{}
		go func() {
			err := h.consume(topic)
			if err != nil {
				log.Error().Err(err).Msg("Consume topic failed")
			}
		}()
	}
}

func (h *consumerHandler) consume(topic enums.Daemon) error { // nolint: unparam
	defer func() {
		<-h.processingSemaphore
	}()
	wq := <-packs[topic]
	ctx := context.Background()
	if h.consumer.Timeout != 0 {
		var ctxCancel context.CancelFunc
		ctx, ctxCancel = context.WithTimeout(ctx, h.consumer.Timeout)
		defer ctxCancel()
	}
	err := h.consumer.Handler(ctx, wq.Payload)
	wq.Times++
	if err != nil {
		log.Error().Err(err).Str("Topic", topic.String()).Msg("Daemon task run failed")
		if wq.Times < h.consumer.MaxRetry {
			packs[topic] <- wq
			return nil
		}
		log.Error().Err(err).Str("Topic", topic.String()).Msg("Daemon task run failed and reach max retry")
		return nil
	}
	return nil
}
