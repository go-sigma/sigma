// Copyright 2026 sigma
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

package redis

import (
	"fmt"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
)

type fakeQueueInspector struct {
	queueInfo *asynq.QueueInfo
	err       error
}

func (f fakeQueueInspector) GetQueueInfo(string) (*asynq.QueueInfo, error) {
	return f.queueInfo, f.err
}

func TestEnsureBacklogAvailable(t *testing.T) {
	tests := []struct {
		name       string
		producer   producer
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "disabled",
			producer: producer{
				maxBacklog: 0,
			},
		},
		{
			name: "below limit",
			producer: producer{
				inspector: fakeQueueInspector{
					queueInfo: &asynq.QueueInfo{Pending: 9},
				},
				maxBacklog: 10,
			},
		},
		{
			name: "reaches limit",
			producer: producer{
				inspector: fakeQueueInspector{
					queueInfo: &asynq.QueueInfo{Pending: 10},
				},
				maxBacklog: 10,
			},
			wantErr:    true,
			wantErrMsg: "work queue backlog limit reached",
		},
		{
			name: "inspect failed",
			producer: producer{
				inspector: fakeQueueInspector{
					err: fmt.Errorf("redis down"),
				},
				maxBacklog: 10,
			},
			wantErr:    true,
			wantErrMsg: "inspect work queue backlog",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.producer.ensureBacklogAvailable()
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrMsg)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestQueueBacklog(t *testing.T) {
	require.Equal(t, 0, queueBacklog(nil))
	require.Equal(t, 10, queueBacklog(&asynq.QueueInfo{
		Pending:     1,
		Active:      100,
		Scheduled:   2,
		Retry:       3,
		Aggregating: 4,
		Archived:    100,
		Completed:   100,
	}))
}
