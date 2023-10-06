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

package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/modules/workq/definition"
	"github.com/go-sigma/sigma/pkg/utils"
)

// ConsumerGroupHandler ...
type ConsumerGroupHandler struct {
	processingSemaphore chan struct{}
	consumer            definition.Consumer
	producer            sarama.SyncProducer
}

// Setup ...
func (h *ConsumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup ...
func (h *ConsumerGroupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim ...
func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		h.processingSemaphore <- struct{}{}
		select {
		case message := <-claim.Messages():
			go func() {
				defer func() {
					<-h.processingSemaphore
				}()
				var msg MessageWrapper
				err := json.Unmarshal(message.Value, &msg)
				if err != nil {
					log.Error().Err(err).Str("message", string(message.Value)).Msg("Unmarshal message failed")
					return
				}
				ctx := session.Context()
				if h.consumer.Timeout != 0 {
					var ctxCancel context.CancelFunc
					ctx, ctxCancel = context.WithTimeout(session.Context(), h.consumer.Timeout)
					defer ctxCancel()
				}
				err = h.consumer.Handler(ctx, message.Value)
				if err != nil {
					msg.Times++
					if msg.Times < h.consumer.MaxRetry {
						_, _, err := h.producer.SendMessage(&sarama.ProducerMessage{
							Topic: message.Topic,
							Value: sarama.ByteEncoder(utils.MustMarshal(msg)),
						})
						if err != nil {
							log.Error().Err(err).Str("topic", message.Topic).Str("payload", string(utils.MustMarshal(msg))).Msg("Resend message failed")
							return
						}
					}
				}
			}()
		case <-session.Context().Done():
			return nil
		}
	}
}

// MessageWrapper ...
type MessageWrapper struct {
	Times   int
	Payload []byte
}

// NewWorkQueueConsumer ...
func NewWorkQueueConsumer(_ configs.Configuration, topicHandlers map[string]definition.Consumer) error {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Consumer.Group.Session.Timeout = 10 * time.Second
	config.Consumer.Group.Heartbeat.Interval = 3 * time.Second
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()

	client, err := sarama.NewClient([]string{}, config)
	if err != nil {
		return err
	}

	producer, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		log.Error().Err(err).Msg("Create producer failed")
	}

	for topic, c := range topicHandlers {
		consumerGroup, err := sarama.NewConsumerGroupFromClient(fmt.Sprintf("%s-%s", consts.AppName, topic), client)
		if err != nil {
			return err
		}
		go func(consumer definition.Consumer, topic string) {
			for {
				handler := &ConsumerGroupHandler{
					processingSemaphore: make(chan struct{}, consumer.Concurrency),
					consumer:            consumer,
					producer:            producer,
				}
				err := consumerGroup.Consume(context.Background(), []string{topic}, handler)
				if err != nil {
					log.Error().Err(err).Str("topic", topic).Msg("Consume topics failed")
					return
				}
			}
		}(c, topic)
	}

	return nil
}
