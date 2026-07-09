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

package inmemory

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/go-sigma/sigma/pkg/infra/lock"
)

func TestAcquireTriesImmediately(t *testing.T) {
	locker := &lockerMemory{locks: make(map[string]*memoryLock)}

	acquiredLock, err := locker.Acquire(t.Context(), "key", lock.MinLockExpire, 10*time.Millisecond)

	require.NoError(t, err)
	require.NotNil(t, acquiredLock)
	require.NoError(t, acquiredLock.Unlock(t.Context()))
}
