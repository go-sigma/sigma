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

	"github.com/google/uuid"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/modules/workq"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
)

func init() {
	utils.PanicIf(workq.RegisterProducer(enums.WorkQueueTypeDatabase, &producerFactory{}))
}

type producer struct {
	workQueueServiceFactory dao.WorkQueueServiceFactory
}

type producerFactory struct{}

var _ workq.ProducerFactory = producerFactory{}

// New ...
func (producerFactory) New(_ *dig.Container) (workq.Producer, error) {
	p := &producer{
		workQueueServiceFactory: dao.NewWorkQueueServiceFactory(),
	}
	return p, nil
}

// Produce ...
func (p *producer) Produce(ctx context.Context, topic enums.Daemon, payload any, option workq.ProducerOption) error {
	tx := query.Q
	if option.Tx != nil {
		tx = option.Tx
	}
	wq := &models.WorkQueue{
		Topic:   topic,
		Payload: utils.MustMarshal(payload),
		Version: uuid.New().String(),
	}
	workQueueService := p.workQueueServiceFactory.New(tx)
	return workQueueService.Create(ctx, wq)
}
