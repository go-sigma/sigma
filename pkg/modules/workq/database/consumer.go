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
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"go.uber.org/dig"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/modules/workq"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
)

func init() {
	utils.PanicIf(workq.RegisterConsumer(enums.WorkQueueTypeDatabase, &consumerFactory{}))
}

type consumerFactory struct{}

var _ workq.ConsumerFactory = consumerFactory{}

// New ...
func (consumerFactory) New(_ *dig.Container, topicHandlers map[enums.Daemon]workq.Consumer) error {
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
				log.Error().Err(err).Msg("consume topic failed")
			}
		}()
		<-time.After(time.Second * 5)
	}
}

func (h *consumerHandler) consume(topic enums.Daemon) error {
	defer func() {
		<-h.processingSemaphore
	}()
	workQueueService := dao.NewWorkQueueServiceFactory().New()
	// daoCtx := log.Logger.WithContext(context.Background())
	daoCtx := context.Background()
	wq, err := workQueueService.Get(daoCtx, topic)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Trace().Err(err).Msgf("none task in topic(%s)", topic)
			return nil
		}
		return err
	}
	newVersion := uuid.New().String()
	err = workQueueService.UpdateStatus(daoCtx, wq.ID, wq.Version, newVersion, wq.Times, enums.TaskCommonStatusDoing)
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
	wq.Times++
	if err != nil {
		log.Error().Err(err).Str("topic", topic.String()).Int64("workQueueID", wq.ID).Msg("daemon task run failed")
		if wq.Times < h.consumer.MaxRetry {
			return workQueueService.UpdateStatus(daoCtx, wq.ID, newVersion, uuid.New().String(), wq.Times, enums.TaskCommonStatusPending)
		}
		return workQueueService.UpdateStatus(daoCtx, wq.ID, newVersion, uuid.New().String(), wq.Times, enums.TaskCommonStatusFailed)
	}
	return workQueueService.UpdateStatus(daoCtx, wq.ID, newVersion, uuid.New().String(), wq.Times, enums.TaskCommonStatusSuccess)
}
