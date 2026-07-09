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

package authz

import (
	"sync"
	"time"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/dal/models"
)

// cacheTTL is the time-to-live for cached entries. A short TTL keeps multi-
// instance state eventually consistent without requiring a Redis watcher.
const cacheTTL = 30 * time.Second

// roleCacheKey identifies a user's role within a namespace.
type roleCacheKey struct {
	userID      string
	namespaceID string
}

type roleCacheEntry struct {
	role      enums.NamespaceRole
	expiresAt time.Time
}

// roleCache is a thread-safe short-TTL cache for namespace membership lookups.
type roleCache struct {
	mu      sync.RWMutex
	entries map[roleCacheKey]roleCacheEntry
}

func newRoleCache() *roleCache {
	return &roleCache{entries: make(map[roleCacheKey]roleCacheEntry)}
}

// get returns the cached role and true if a fresh entry exists.
func (c *roleCache) get(userID, namespaceID string) (enums.NamespaceRole, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.entries[roleCacheKey{userID, namespaceID}]
	if !ok || time.Now().After(e.expiresAt) {
		return "", false
	}
	return e.role, true
}

// set stores a role with the standard TTL. An empty role is also cached to
// represent "not a member" and avoid repeated DB lookups.
func (c *roleCache) set(userID, namespaceID string, role enums.NamespaceRole) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[roleCacheKey{userID, namespaceID}] = roleCacheEntry{
		role:      role,
		expiresAt: time.Now().Add(cacheTTL),
	}
}

// nsCacheKey identifies a namespace by name (for /v2/ requests) or by id (for
// /api/ requests). Only one of name/id is populated per key.
type nsCacheKey struct {
	name string
	id   string
}

type nsCacheEntry struct {
	namespace *models.Namespace
	expiresAt time.Time
}

// namespaceCache is a thread-safe short-TTL cache for namespace lookups.
type namespaceCache struct {
	mu      sync.RWMutex
	entries map[nsCacheKey]nsCacheEntry
}

func newNamespaceCache() *namespaceCache {
	return &namespaceCache{entries: make(map[nsCacheKey]nsCacheEntry)}
}

func (c *namespaceCache) getByName(name string) (*models.Namespace, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.entries[nsCacheKey{name: name}]
	if !ok || time.Now().After(e.expiresAt) {
		return nil, false
	}
	return e.namespace, true
}

func (c *namespaceCache) getByID(id string) (*models.Namespace, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.entries[nsCacheKey{id: id}]
	if !ok || time.Now().After(e.expiresAt) {
		return nil, false
	}
	return e.namespace, true
}

func (c *namespaceCache) set(ns *models.Namespace) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry := nsCacheEntry{namespace: ns, expiresAt: time.Now().Add(cacheTTL)}
	c.entries[nsCacheKey{name: ns.Name}] = entry
	c.entries[nsCacheKey{id: ns.ID}] = entry
}
