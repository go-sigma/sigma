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
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/dal/models"
)

func TestProducerProduce(t *testing.T) {
	originalPacks := packs
	t.Cleanup(func() {
		packs = originalPacks
	})

	tests := []struct {
		name       string
		before     func()
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "success",
			before: func() {
				packs = map[enums.Daemon]chan *models.WorkQueue{
					enums.DaemonBuilder: make(chan *models.WorkQueue, 1),
				}
			},
		},
		{
			name: "queue full",
			before: func() {
				queue := make(chan *models.WorkQueue, 1)
				queue <- &models.WorkQueue{Topic: enums.DaemonBuilder}
				packs = map[enums.Daemon]chan *models.WorkQueue{
					enums.DaemonBuilder: queue,
				}
			},
			wantErr:    true,
			wantErrMsg: "work queue backlog limit reached",
		},
		{
			name: "topic not initialized",
			before: func() {
				packs = map[enums.Daemon]chan *models.WorkQueue{}
			},
			wantErr:    true,
			wantErrMsg: "is not initialized",
		},
	}

	producer := &producer{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.before()
			err := producer.Produce(t.Context(), enums.DaemonBuilder, "test")
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrMsg)
				return
			}
			require.NoError(t, err)
			require.Len(t, packs[enums.DaemonBuilder], 1)
		})
	}
}

func TestProducerProduceMarshalFailed(t *testing.T) {
	producer := &producer{}

	err := producer.Produce(t.Context(), enums.DaemonBuilder, make(chan int))

	require.Error(t, err)
	require.Contains(t, err.Error(), "marshal work queue payload")
}
