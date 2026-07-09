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

	"github.com/stretchr/testify/require"

	"github.com/go-sigma/sigma/pkg/infra/counter"
)

func TestHIncrByAndHGetAll(t *testing.T) {
	ctr, err := factory{}.New(counter.Params{})
	require.NoError(t, err)

	err = ctr.HIncrBy(t.Context(), "hash1", map[string]int64{"a": 1, "b": 2})
	require.NoError(t, err)

	err = ctr.HIncrBy(t.Context(), "hash1", map[string]int64{"a": 3, "c": 4})
	require.NoError(t, err)

	data, err := ctr.HGetAll(t.Context(), "hash1")
	require.NoError(t, err)
	require.Equal(t, map[string]int64{"a": 4, "b": 2, "c": 4}, data)
}

func TestHGetAllMissingKey(t *testing.T) {
	ctr, err := factory{}.New(counter.Params{})
	require.NoError(t, err)

	data, err := ctr.HGetAll(t.Context(), "nonexistent")
	require.NoError(t, err)
	require.Nil(t, data)
}

func TestSetOperations(t *testing.T) {
	ctr, err := factory{}.New(counter.Params{})
	require.NoError(t, err)

	err = ctr.SAdd(t.Context(), "set1", "m1", "m2")
	require.NoError(t, err)

	members, err := ctr.SMembers(t.Context(), "set1")
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"m1", "m2"}, members)

	err = ctr.SRem(t.Context(), "set1", "m1")
	require.NoError(t, err)

	members, err = ctr.SMembers(t.Context(), "set1")
	require.NoError(t, err)
	require.Equal(t, []string{"m2"}, members)
}

func TestSMembersMissingKey(t *testing.T) {
	ctr, err := factory{}.New(counter.Params{})
	require.NoError(t, err)

	members, err := ctr.SMembers(t.Context(), "nonexistent")
	require.NoError(t, err)
	require.Nil(t, members)
}

func TestRename(t *testing.T) {
	ctr, err := factory{}.New(counter.Params{})
	require.NoError(t, err)

	err = ctr.HIncrBy(t.Context(), "old", map[string]int64{"x": 10})
	require.NoError(t, err)

	err = ctr.Rename(t.Context(), "old", "new")
	require.NoError(t, err)

	// old key should be gone
	data, err := ctr.HGetAll(t.Context(), "old")
	require.NoError(t, err)
	require.Nil(t, data)

	// new key should have the data
	data, err = ctr.HGetAll(t.Context(), "new")
	require.NoError(t, err)
	require.Equal(t, map[string]int64{"x": 10}, data)
}

func TestRenameTargetExists(t *testing.T) {
	ctr, err := factory{}.New(counter.Params{})
	require.NoError(t, err)

	err = ctr.HIncrBy(t.Context(), "old", map[string]int64{"x": 10})
	require.NoError(t, err)
	err = ctr.HIncrBy(t.Context(), "new", map[string]int64{"y": 20})
	require.NoError(t, err)

	// rename should be a no-op when target exists
	err = ctr.Rename(t.Context(), "old", "new")
	require.NoError(t, err)

	data, err := ctr.HGetAll(t.Context(), "old")
	require.NoError(t, err)
	require.Equal(t, map[string]int64{"x": 10}, data)

	data, err = ctr.HGetAll(t.Context(), "new")
	require.NoError(t, err)
	require.Equal(t, map[string]int64{"y": 20}, data)
}

func TestDel(t *testing.T) {
	ctr, err := factory{}.New(counter.Params{})
	require.NoError(t, err)

	err = ctr.HIncrBy(t.Context(), "hash1", map[string]int64{"a": 1})
	require.NoError(t, err)
	err = ctr.SAdd(t.Context(), "set1", "m1")
	require.NoError(t, err)

	err = ctr.Del(t.Context(), "hash1", "set1")
	require.NoError(t, err)

	data, err := ctr.HGetAll(t.Context(), "hash1")
	require.NoError(t, err)
	require.Nil(t, data)

	members, err := ctr.SMembers(t.Context(), "set1")
	require.NoError(t, err)
	require.Nil(t, members)
}

func TestExists(t *testing.T) {
	ctr, err := factory{}.New(counter.Params{})
	require.NoError(t, err)

	err = ctr.HIncrBy(t.Context(), "hash1", map[string]int64{"a": 1})
	require.NoError(t, err)
	err = ctr.SAdd(t.Context(), "set1", "m1")
	require.NoError(t, err)

	count, err := ctr.Exists(t.Context(), "hash1", "set1", "missing")
	require.NoError(t, err)
	require.Equal(t, int64(2), count)
}

func TestShutdown(t *testing.T) {
	ctr, err := factory{}.New(counter.Params{})
	require.NoError(t, err)

	err = ctr.Shutdown(t.Context())
	require.NoError(t, err)
}

func TestConcurrentAccess(t *testing.T) {
	ctr, err := factory{}.New(counter.Params{})
	require.NoError(t, err)

	const goroutines = 100
	const incrementsPerGoroutine = 100

	done := make(chan struct{})
	for range goroutines {
		go func() {
			for range incrementsPerGoroutine {
				_ = ctr.HIncrBy(t.Context(), "counter", map[string]int64{"val": 1})
				_ = ctr.SAdd(t.Context(), "set", "m")
			}
			done <- struct{}{}
		}()
	}
	for range goroutines {
		<-done
	}

	data, err := ctr.HGetAll(t.Context(), "counter")
	require.NoError(t, err)
	require.Equal(t, int64(goroutines*incrementsPerGoroutine), data["val"])

	members, err := ctr.SMembers(t.Context(), "set")
	require.NoError(t, err)
	require.Len(t, members, 1)
}
