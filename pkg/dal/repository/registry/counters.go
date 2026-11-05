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

	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/dal/query"
)

// incrementTagCount adds delta to both the repository and namespace tag
// counters. It must run in the same transaction as the tag row insert.
func incrementTagCount(ctx context.Context, tx *query.Query, repositoryID string, delta int64) error {
	if delta <= 0 {
		return nil
	}
	repositoryObj, err := tx.Repository.WithContext(ctx).
		Select(tx.Repository.NamespaceID).
		Where(tx.Repository.ID.Eq(repositoryID)).
		First()
	if err != nil {
		return err
	}
	// Lock the namespace before the repository so the order matches the push
	// path (namespace -> repository) and cannot deadlock against it.
	if err := incrementNamespaceCounts(ctx, tx, repositoryObj.NamespaceID, delta, 0); err != nil {
		return err
	}
	_, err = tx.Repository.WithContext(ctx).
		Where(tx.Repository.ID.Eq(repositoryID)).
		UpdateColumns(map[string]any{
			"tag_count": gorm.Expr("tag_count + ?", delta),
		})
	return err
}

// decrementTagCount subtracts delta from both the repository and namespace tag
// counters. It must run in the same transaction as the tag row delete.
func decrementTagCount(ctx context.Context, tx *query.Query, repositoryID string, delta int64) error {
	if delta <= 0 {
		return nil
	}
	repositoryObj, err := tx.Repository.WithContext(ctx).
		Select(tx.Repository.NamespaceID).
		Where(tx.Repository.ID.Eq(repositoryID)).
		First()
	if err != nil {
		return err
	}
	// Lock the namespace before the repository so the order matches the push
	// path (namespace -> repository) and cannot deadlock against it.
	if err := decrementNamespaceCounts(ctx, tx, repositoryObj.NamespaceID, delta, 0); err != nil {
		return err
	}
	return decrementRepositoryTagCount(ctx, tx, repositoryID, delta)
}

// incrementNamespaceCounts adds the given tag and repository deltas to a
// namespace. It must run in the same transaction as the write it accounts for.
func incrementNamespaceCounts(ctx context.Context, tx *query.Query, namespaceID string, tagDelta, repositoryDelta int64) error {
	updates := make(map[string]any, 2)
	if tagDelta > 0 {
		updates["tag_count"] = gorm.Expr("tag_count + ?", tagDelta)
	}
	if repositoryDelta > 0 {
		updates["repository_count"] = gorm.Expr("repository_count + ?", repositoryDelta)
	}
	if len(updates) == 0 {
		return nil
	}
	_, err := tx.Namespace.WithContext(ctx).
		Where(tx.Namespace.ID.Eq(namespaceID)).
		UpdateColumns(updates)
	return err
}

// decrementNamespaceCounts subtracts the given tag and repository deltas from a
// namespace, flooring each counter at zero.
func decrementNamespaceCounts(ctx context.Context, tx *query.Query, namespaceID string, tagDelta, repositoryDelta int64) error {
	if tagDelta <= 0 && repositoryDelta <= 0 {
		return nil
	}
	namespaceObj, err := tx.Namespace.WithContext(ctx).
		Select(tx.Namespace.TagCount, tx.Namespace.RepositoryCount).
		Where(tx.Namespace.ID.Eq(namespaceID)).
		First()
	if err != nil {
		return err
	}
	if tagDelta > 0 {
		result, err := tx.Namespace.WithContext(ctx).
			Where(tx.Namespace.ID.Eq(namespaceID), tx.Namespace.TagCount.Gte(tagDelta)).
			UpdateColumns(map[string]any{
				"tag_count": gorm.Expr("tag_count - ?", tagDelta),
			})
		if err != nil {
			return err
		}
		if result.RowsAffected == 0 && namespaceObj.TagCount > 0 {
			if _, err = tx.Namespace.WithContext(ctx).
				Where(tx.Namespace.ID.Eq(namespaceID)).
				UpdateColumns(map[string]any{"tag_count": 0}); err != nil {
				return err
			}
		}
	}
	if repositoryDelta > 0 {
		result, err := tx.Namespace.WithContext(ctx).
			Where(tx.Namespace.ID.Eq(namespaceID), tx.Namespace.RepositoryCount.Gte(repositoryDelta)).
			UpdateColumns(map[string]any{
				"repository_count": gorm.Expr("repository_count - ?", repositoryDelta),
			})
		if err != nil {
			return err
		}
		if result.RowsAffected == 0 && namespaceObj.RepositoryCount > 0 {
			if _, err = tx.Namespace.WithContext(ctx).
				Where(tx.Namespace.ID.Eq(namespaceID)).
				UpdateColumns(map[string]any{"repository_count": 0}); err != nil {
				return err
			}
		}
	}
	return nil
}

// decrementRepositoryTagCount subtracts delta from a repository tag counter,
// flooring it at zero.
func decrementRepositoryTagCount(ctx context.Context, tx *query.Query, repositoryID string, delta int64) error {
	repositoryObj, err := tx.Repository.WithContext(ctx).
		Select(tx.Repository.TagCount).
		Where(tx.Repository.ID.Eq(repositoryID)).
		First()
	if err != nil {
		return err
	}
	result, err := tx.Repository.WithContext(ctx).
		Where(tx.Repository.ID.Eq(repositoryID), tx.Repository.TagCount.Gte(delta)).
		UpdateColumns(map[string]any{
			"tag_count": gorm.Expr("tag_count - ?", delta),
		})
	if err != nil {
		return err
	}
	if result.RowsAffected == 0 && repositoryObj.TagCount > 0 {
		_, err = tx.Repository.WithContext(ctx).
			Where(tx.Repository.ID.Eq(repositoryID)).
			UpdateColumns(map[string]any{"tag_count": 0})
	}
	return err
}
