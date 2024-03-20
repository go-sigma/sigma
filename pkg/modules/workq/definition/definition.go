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

package definition

import (
	"context"
	"time"

	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/types/enums"
)

//go:generate mockgen -destination=mocks/workq.go -package=mocks github.com/go-sigma/sigma/pkg/modules/workq/definition WorkQueueProducer

// Consumer ...
type Consumer struct {
	Handler     func(ctx context.Context, payload []byte) error
	Concurrency int
	MaxRetry    int
	Timeout     time.Duration
}

// ProducerOption ...
type ProducerOption struct {
	Tx *query.Query
}

// WorkQueueProducer ...
type WorkQueueProducer interface {
	// Produce ...
	Produce(ctx context.Context, topic enums.Daemon, payload any, option ProducerOption) error
}
