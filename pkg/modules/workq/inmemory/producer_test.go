// Copyright 2024 sigma
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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/modules/workq/definition"
	"github.com/go-sigma/sigma/pkg/types/enums"
)

func TestProducer(t *testing.T) {
	producer, err := NewWorkQueueProducer(configs.Configuration{}, nil)
	assert.NoError(t, err)
	assert.NotNil(t, producer)

	packs[enums.DaemonBuilder] = make(chan *models.WorkQueue, 10)
	err = producer.Produce(context.Background(), enums.DaemonBuilder, "test", definition.ProducerOption{})
	assert.NoError(t, err)
}
