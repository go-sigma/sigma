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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/modules/workq/definition"
	"github.com/go-sigma/sigma/pkg/types/enums"
)

func TestConsumer(t *testing.T) {
	var times int
	var topicHandlers = map[enums.Daemon]definition.Consumer{
		enums.DaemonBuilder: {
			Handler: func(ctx context.Context, payload []byte) error {
				times++
				return nil
			},
			Concurrency: 3,
			MaxRetry:    3,
			Timeout:     time.Second * 3,
		},
	}
	err := NewWorkQueueConsumer(configs.Configuration{}, topicHandlers)
	assert.NoError(t, err)

	packs[enums.DaemonBuilder] <- &models.WorkQueue{Topic: enums.DaemonBuilder, Payload: []byte{}}
	<-time.After(time.Second)
	assert.Equal(t, 1, times)
}

func TestConsumerWithTimeout(t *testing.T) {
	var times int
	var topicHandlers = map[enums.Daemon]definition.Consumer{
		enums.DaemonBuilder: {
			Handler: func(ctx context.Context, payload []byte) error {
				<-time.After(time.Second * 2)
				select {
				case <-ctx.Done():
					times++
					return fmt.Errorf("operation timeout")
				default:
				}
				return nil
			},
			Concurrency: 3,
			MaxRetry:    3,
			Timeout:     time.Second,
		},
	}
	err := NewWorkQueueConsumer(configs.Configuration{}, topicHandlers)
	assert.NoError(t, err)

	packs[enums.DaemonBuilder] <- &models.WorkQueue{Topic: enums.DaemonBuilder, Payload: []byte{}}
	<-time.After(time.Second * 10)
	assert.Equal(t, 3, times)
}
