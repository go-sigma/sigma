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

package gc

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunWorkersReturnsFirstError(t *testing.T) {
	expectedErr := errors.New("worker failed")
	var processed atomic.Int64

	err := runWorkers(t.Context(), 2, []int{1, 2, 3}, func(_ context.Context, item int) error {
		processed.Add(1)
		if item == 2 {
			return expectedErr
		}
		return nil
	})

	require.ErrorIs(t, err, expectedErr)
	require.Positive(t, processed.Load())
}
