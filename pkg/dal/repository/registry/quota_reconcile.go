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

package registry

import (
	"context"
	"fmt"

	"github.com/go-sigma/sigma/pkg/dal/query"
)

// ReconcileScopeSize repairs the blob attributions for a scope and writes the
// resulting size back in one transaction.
//
// The first statement is a write to the scope row. That serializes the
// reconcile with IncrementSize/DecrementSize on the same row (and acquires the
// SQLite write lock), so a concurrent manifest push cannot commit between the
// size computation and the write-back. Keeping the attribution repair and the
// counter update in the same transaction also avoids clearing size_dirty on a
// value that a concurrent push has already changed.
func ReconcileScopeSize(ctx context.Context, scopeType, scopeID string) (int64, error) {
	var size int64
	err := query.Q.Transaction(func(tx *query.Query) error {
		switch scopeType {
		case ScopeTypeNamespace:
			if _, err := tx.Namespace.WithContext(ctx).
				Where(tx.Namespace.ID.Eq(scopeID)).
				UpdateColumns(map[string]any{"size_dirty": true}); err != nil {
				return err
			}
			actual, err := NewBlobAttributionRepository(tx).SyncAndSum(ctx, scopeType, scopeID)
			if err != nil {
				return err
			}
			size = actual
			_, err = tx.Namespace.WithContext(ctx).
				Where(tx.Namespace.ID.Eq(scopeID)).
				UpdateColumns(map[string]any{
					"size":       actual,
					"size_dirty": false,
				})
			return err
		case ScopeTypeRepository:
			if _, err := tx.Repository.WithContext(ctx).
				Where(tx.Repository.ID.Eq(scopeID)).
				UpdateColumns(map[string]any{"size_dirty": true}); err != nil {
				return err
			}
			actual, err := NewBlobAttributionRepository(tx).SyncAndSum(ctx, scopeType, scopeID)
			if err != nil {
				return err
			}
			size = actual
			_, err = tx.Repository.WithContext(ctx).
				Where(tx.Repository.ID.Eq(scopeID)).
				UpdateColumns(map[string]any{
					"size":       actual,
					"size_dirty": false,
				})
			return err
		default:
			return fmt.Errorf("invalid quota scope type: %s", scopeType)
		}
	})
	return size, err
}
