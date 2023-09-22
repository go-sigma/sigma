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
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/modules/workq/definition"
	"github.com/go-sigma/sigma/pkg/types/enums"
)

// NewWorkQueueConsumer ...
func NewWorkQueueConsumer(_ configs.Configuration, topicHandlers map[string]definition.Consumer) error {
	for topic, c := range topicHandlers {
		go func(consumer definition.Consumer, topic string) {
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
	consumer            definition.Consumer
}

func (h *consumerHandler) Consume(topic string) {
	for {
		h.processingSemaphore <- struct{}{}
		go func() {
			err := h.consume(topic)
			if err != nil {
				log.Error().Err(err).Msg("Consume topic failed")
			}
		}()
		<-time.After(time.Second * 5)
	}
}

func (h *consumerHandler) consume(topic string) error {
	defer func() {
		<-h.processingSemaphore
	}()
	workQueueService := dao.NewWorkQueueServiceFactory().New()
	// daoCtx := log.Logger.WithContext(context.Background())
	daoCtx := context.Background()
	wq, err := workQueueService.Get(daoCtx, strings.ToLower(topic))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Trace().Err(err).Msgf("None task in topic(%s)", topic)
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
		if wq.Times < h.consumer.MaxRetry {
			return workQueueService.UpdateStatus(daoCtx, wq.ID, newVersion, uuid.New().String(), wq.Times, enums.TaskCommonStatusPending)
		}
		return workQueueService.UpdateStatus(daoCtx, wq.ID, newVersion, uuid.New().String(), wq.Times, enums.TaskCommonStatusFailed)
	}
	return workQueueService.UpdateStatus(daoCtx, wq.ID, newVersion, uuid.New().String(), wq.Times, enums.TaskCommonStatusSuccess)
}
