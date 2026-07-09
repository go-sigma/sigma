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
	"context"
	"maps"
	"sync"

	counter "github.com/go-sigma/sigma/pkg/infra/counter"
	"github.com/go-sigma/sigma/pkg/utils"
)

func init() {
	utils.PanicIf(counter.Factories.Register("inmemory", &factory{}))
}

type hashEntry struct {
	mu   sync.Mutex
	data map[string]int64
}

type setEntry struct {
	mu   sync.Mutex
	data map[string]struct{}
}

type inmemoryCounter struct {
	mu     sync.Mutex
	hashes map[string]*hashEntry
	sets   map[string]*setEntry
}

type factory struct{}

var _ counter.Factory = factory{}

// New creates a new in-memory counter.
func (factory) New(_ counter.Params) (counter.Counter, error) {
	return &inmemoryCounter{
		hashes: make(map[string]*hashEntry),
		sets:   make(map[string]*setEntry),
	}, nil
}

func (c *inmemoryCounter) getHash(key string) *hashEntry {
	c.mu.Lock()
	defer c.mu.Unlock()
	h, ok := c.hashes[key]
	if !ok {
		h = &hashEntry{data: make(map[string]int64)}
		c.hashes[key] = h
	}
	return h
}

func (c *inmemoryCounter) getSet(key string) *setEntry {
	c.mu.Lock()
	defer c.mu.Unlock()
	s, ok := c.sets[key]
	if !ok {
		s = &setEntry{data: make(map[string]struct{})}
		c.sets[key] = s
	}
	return s
}

// HIncrBy increments hash fields by the given deltas.
func (c *inmemoryCounter) HIncrBy(_ context.Context, key string, fields map[string]int64) error {
	h := c.getHash(key)
	h.mu.Lock()
	defer h.mu.Unlock()
	for field, delta := range fields {
		h.data[field] += delta
	}
	return nil
}

// HGetAll returns all fields and values of a hash.
func (c *inmemoryCounter) HGetAll(_ context.Context, key string) (map[string]int64, error) {
	c.mu.Lock()
	h, ok := c.hashes[key]
	c.mu.Unlock()
	if !ok {
		return nil, nil
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	return maps.Clone(h.data), nil
}

// Del removes one or more keys.
func (c *inmemoryCounter) Del(_ context.Context, keys ...string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, key := range keys {
		delete(c.hashes, key)
		delete(c.sets, key)
	}
	return nil
}

// SAdd adds members to a set.
func (c *inmemoryCounter) SAdd(_ context.Context, key string, members ...string) error {
	s := c.getSet(key)
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, m := range members {
		s.data[m] = struct{}{}
	}
	return nil
}

// SMembers returns all members of a set.
func (c *inmemoryCounter) SMembers(_ context.Context, key string) ([]string, error) {
	c.mu.Lock()
	s, ok := c.sets[key]
	c.mu.Unlock()
	if !ok {
		return nil, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]string, 0, len(s.data))
	for m := range s.data {
		result = append(result, m)
	}
	return result, nil
}

// SRem removes members from a set.
func (c *inmemoryCounter) SRem(_ context.Context, key string, members ...string) error {
	c.mu.Lock()
	s, ok := c.sets[key]
	c.mu.Unlock()
	if !ok {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, m := range members {
		delete(s.data, m)
	}
	return nil
}

// Rename atomically renames a key. Returns nil if the target already exists.
func (c *inmemoryCounter) Rename(_ context.Context, oldKey, newKey string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.hashes[newKey]; ok {
		return nil
	}
	if _, ok := c.sets[newKey]; ok {
		return nil
	}

	if h, ok := c.hashes[oldKey]; ok {
		c.hashes[newKey] = h
		delete(c.hashes, oldKey)
	}
	if s, ok := c.sets[oldKey]; ok {
		c.sets[newKey] = s
		delete(c.sets, oldKey)
	}
	return nil
}

// Exists returns the number of keys that exist among the given keys.
func (c *inmemoryCounter) Exists(_ context.Context, keys ...string) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	var count int64
	for _, key := range keys {
		if _, ok := c.hashes[key]; ok {
			count++
			continue
		}
		if _, ok := c.sets[key]; ok {
			count++
		}
	}
	return count, nil
}

// Shutdown gracefully shuts down the counter.
func (c *inmemoryCounter) Shutdown(_ context.Context) error {
	return nil
}
