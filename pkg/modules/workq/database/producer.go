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

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/modules/workq/definition"
	"github.com/go-sigma/sigma/pkg/utils"
)

type producer struct {
	workQueueServiceFactory dao.WorkQueueServiceFactory
}

// NewWorkQueueProducer ...
func NewWorkQueueProducer(_ configs.Configuration, _ map[string]definition.Consumer) (definition.WorkQueueProducer, error) {
	p := &producer{
		workQueueServiceFactory: dao.NewWorkQueueServiceFactory(),
	}
	return p, nil
}

// Produce ...
func (p *producer) Produce(ctx context.Context, topic string, payload any) error {
	wq := &models.WorkQueue{
		Topic:   topic,
		Payload: utils.MustMarshal(payload),
		Version: uuid.New().String(),
	}
	workQueueService := p.workQueueServiceFactory.New()
	return workQueueService.Create(ctx, wq)
}
