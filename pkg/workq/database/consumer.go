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

package database

import (
	"context"
	"path"
	"reflect"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/workq"
)

func init() {
	workq.ConsumerClientFactories[path.Base(reflect.TypeOf(consumerFactory{}).PkgPath())] = &consumerFactory{}
}

type consumerFactory struct{}

// NewWorkQueueConsumer ...
func (f consumerFactory) New(_ configs.Configuration) error {
	for topic, c := range workq.TopicConsumers {
		go func(consumer workq.Consumer, topic string) {
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

func (h *consumerHandler) Consume(topic string) {
	for {
		err := h.consume()
		if err != nil {
			log.Error().Err(err).Msg("Consume topic failed")
		}
	}
}

func (h *consumerHandler) consume() error {
	h.processingSemaphore <- struct{}{}
	defer func() {
		<-h.processingSemaphore
	}()
	workQueueService := dao.NewWorkQueueServiceFactory().New()
	daoCtx := log.Logger.WithContext(context.Background())
	wq, err := workQueueService.Get(daoCtx)
	if err != nil {
		return err
	}
	newVersion := uuid.New().String()
	err = workQueueService.UpdateStatus(daoCtx, wq.ID, wq.Version, newVersion, enums.TaskCommonStatusDoing)
	if err != nil {
		return err
	}
	ctx := context.Background()
	if h.consumer.Timeout != 0 {
		var ctxCancel context.CancelFunc
		ctx, ctxCancel = context.WithTimeout(ctx, h.consumer.Timeout)
		defer ctxCancel()
	}
	err = h.consumer.Handler(ctx, wq.Payload)
	if err != nil {
		wq.Times++
		if wq.Times < h.consumer.MaxRetry {
			return workQueueService.UpdateStatus(daoCtx, wq.ID, newVersion, uuid.New().String(), enums.TaskCommonStatusPending)
		}
	}
	return nil
}