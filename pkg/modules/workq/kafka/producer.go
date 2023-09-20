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

	"github.com/IBM/sarama"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/modules/workq/definition"
	"github.com/go-sigma/sigma/pkg/utils"
)

// NewWorkQueueProducer ...
func NewWorkQueueProducer(_ configs.Configuration, _ map[string]definition.Consumer) (definition.WorkQueueProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	client, err := sarama.NewClient([]string{}, config)
	if err != nil {
		return nil, err
	}

	p, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		log.Error().Err(err).Msg("Create producer failed")
		return nil, err
	}
	return &producer{
		producer: p,
	}, nil
}

type producer struct {
	producer sarama.SyncProducer
}

func (p *producer) Produce(_ context.Context, topic string, payload any) error {
	message := MessageWrapper{
		Times:   0,
		Payload: utils.MustMarshal(payload),
	}
	_, _, err := p.producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(utils.MustMarshal(message)),
	})
	return err
}
