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
	"log/slog"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/infra/workq"
	"github.com/go-sigma/sigma/pkg/utils"
)

func init() {
	utils.PanicIf(workq.RegisterConsumer(enums.WorkQueueTypeInmemory, &consumerFactory{}))
}

type consumerFactory struct{}

var _ workq.ConsumerFactory = consumerFactory{}

const defaultMaxBacklog = 1000

// This is only for small-scale deployment; the default queue backlog limit is 1000 messages.
var packs = make(map[enums.Daemon]chan *models.WorkQueue, 10)

// New ...
func (consumerFactory) New(params workq.ConsumerParams) error {
	maxBacklog := params.Config.WorkQueue.Inmemory.MaxBacklog
	if maxBacklog <= 0 {
		maxBacklog = defaultMaxBacklog
	}
	topicHandlers := params.HandlerRegistry.Handlers()
	for topic := range topicHandlers {
		packs[topic] = make(chan *models.WorkQueue, maxBacklog)
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
				slog.Error("consume topic failed", "err", err)
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
	ctx, payload := workq.UnmarshalPayload(ctx, wq.Payload)
	ctx, span := workq.StartConsumerSpan(ctx, topic)
	defer span.End()
	err := h.consumer.Handler(ctx, payload)
	if err != nil {
		span.RecordError(err)
		slog.Error("daemon task run failed", "err", err, "topic", topic.String())
		return nil
	}
	return nil
}
