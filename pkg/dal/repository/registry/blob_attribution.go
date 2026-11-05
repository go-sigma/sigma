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

	"gorm.io/gorm/clause"

	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

// Quota scope types recorded in blob_attributions.
const (
	// ScopeTypeNamespace records the blobs charged to a namespace.
	ScopeTypeNamespace = "namespace"
	// ScopeTypeRepository records the blobs charged to a repository.
	ScopeTypeRepository = "repository"
)

const (
	// blobAttributionTable is the table name used by raw GORM queries.
	blobAttributionTable = "blob_attributions"
	// attributionBatchSize bounds the number of rows per INSERT/DELETE.
	attributionBatchSize = 100
)

//go:generate go tool mockgen -destination=mocks/blob_attribution_mocks.go -package=mocks github.com/go-sigma/sigma/pkg/dal/repository/registry BlobAttributionRepository

// BlobAttributionRepository records and queries which blob is charged to
// which quota scope. The (scope_type, scope_id, blob_id) primary key makes
// writes idempotent, so concurrent manifest pushes cannot double count a blob.
type BlobAttributionRepository interface {
	// UpsertBatch inserts the blobs attributed to a scope. Existing rows are
	// left untouched (their batch_id is preserved), which makes the call
	// idempotent. The batch_id identifies the rows actually inserted by this
	// caller so that SumBatch can compute the delta.
	UpsertBatch(ctx context.Context, scopeType, scopeID string, blobs []*models.Blob, batchID string) error
	// SumBatch sums the sizes of the rows inserted with the given batch_id.
	SumBatch(ctx context.Context, scopeType, scopeID, batchID string) (int64, error)
	// SumByScope sums the sizes of all blobs attributed to the scope.
	SumByScope(ctx context.Context, scopeType, scopeID string) (int64, error)
	// SyncAndSum repairs the attribution rows for the scope against the
	// current reachable blob set, then returns the deduplicated size.
	SyncAndSum(ctx context.Context, scopeType, scopeID string) (int64, error)
	// DeleteByBlobIDs removes all attribution rows for the given blobs. It is
	// used when a blob is physically deleted.
	DeleteByBlobIDs(ctx context.Context, blobIDs []string) error
}

type blobAttributionRepository struct {
	tx *query.Query
}

// NewBlobAttributionRepository creates a new blob attribution repository with
// the optional query transaction.
func NewBlobAttributionRepository(txs ...*query.Query) BlobAttributionRepository {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &blobAttributionRepository{tx: tx}
}

// UpsertBatch inserts the blobs attributed to a scope.
func (s *blobAttributionRepository) UpsertBatch(ctx context.Context, scopeType, scopeID string, blobs []*models.Blob, batchID string) error {
	if len(blobs) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(blobs))
	rows := make([]*models.BlobAttribution, 0, len(blobs))
	for _, blob := range blobs {
		if blob == nil {
			continue
		}
		if _, ok := seen[blob.ID]; ok {
			continue
		}
		seen[blob.ID] = struct{}{}
		rows = append(rows, &models.BlobAttribution{
			ScopeType: scopeType,
			ScopeID:   scopeID,
			BlobID:    blob.ID,
			Size:      blob.Size,
			BatchID:   batchID,
		})
	}
	if len(rows) == 0 {
		return nil
	}

	return s.tx.UnderlyingDB().WithContext(ctx).
		Table(blobAttributionTable).
		Clauses(clause.OnConflict{DoNothing: true}).
		CreateInBatches(rows, attributionBatchSize).Error
}

// SumBatch sums the sizes of the rows inserted with the given batch_id.
func (s *blobAttributionRepository) SumBatch(ctx context.Context, scopeType, scopeID, batchID string) (int64, error) {
	var row struct {
		Total int64
	}
	err := s.tx.UnderlyingDB().WithContext(ctx).
		Table(blobAttributionTable).
		Select("COALESCE(SUM(size), 0) AS total").
		Where("scope_type = ? AND scope_id = ? AND batch_id = ?", scopeType, scopeID, batchID).
		Scan(&row).Error
	if err != nil {
		return 0, err
	}
	return row.Total, nil
}

// SumByScope sums the sizes of all blobs attributed to the scope.
func (s *blobAttributionRepository) SumByScope(ctx context.Context, scopeType, scopeID string) (int64, error) {
	var row struct {
		Total int64
	}
	err := s.tx.UnderlyingDB().WithContext(ctx).
		Table(blobAttributionTable).
		Select("COALESCE(SUM(size), 0) AS total").
		Where("scope_type = ? AND scope_id = ?", scopeType, scopeID).
		Scan(&row).Error
	if err != nil {
		return 0, err
	}
	return row.Total, nil
}

// SyncAndSum repairs the attribution rows for the scope against the current
// reachable blob set, then returns the deduplicated size.
func (s *blobAttributionRepository) SyncAndSum(ctx context.Context, scopeType, scopeID string) (int64, error) {
	reachable, err := s.reachableBlobs(ctx, scopeType, scopeID)
	if err != nil {
		return 0, err
	}

	if err = s.UpsertBatch(ctx, scopeType, scopeID, reachable, uuid.NewV7String()); err != nil {
		return 0, err
	}

	attributed, err := s.attributedBlobIDs(ctx, scopeType, scopeID)
	if err != nil {
		return 0, err
	}
	reachableSet := make(map[string]struct{}, len(reachable))
	for _, blob := range reachable {
		reachableSet[blob.ID] = struct{}{}
	}
	extra := make([]string, 0)
	for _, blobID := range attributed {
		if _, ok := reachableSet[blobID]; !ok {
			extra = append(extra, blobID)
		}
	}
	if err = s.deleteByBlobIDs(ctx, scopeType, scopeID, extra); err != nil {
		return 0, err
	}

	return s.SumByScope(ctx, scopeType, scopeID)
}

// DeleteByBlobIDs removes all attribution rows for the given blobs.
func (s *blobAttributionRepository) DeleteByBlobIDs(ctx context.Context, blobIDs []string) error {
	if len(blobIDs) == 0 {
		return nil
	}
	for start := 0; start < len(blobIDs); start += attributionBatchSize {
		end := min(start+attributionBatchSize, len(blobIDs))
		err := s.tx.UnderlyingDB().WithContext(ctx).
			Table(blobAttributionTable).
			Where("blob_id IN ?", blobIDs[start:end]).
			Delete(&models.BlobAttribution{}).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *blobAttributionRepository) deleteByBlobIDs(ctx context.Context, scopeType, scopeID string, blobIDs []string) error {
	if len(blobIDs) == 0 {
		return nil
	}
	for start := 0; start < len(blobIDs); start += attributionBatchSize {
		end := min(start+attributionBatchSize, len(blobIDs))
		err := s.tx.UnderlyingDB().WithContext(ctx).
			Table(blobAttributionTable).
			Where("scope_type = ? AND scope_id = ? AND blob_id IN ?", scopeType, scopeID, blobIDs[start:end]).
			Delete(&models.BlobAttribution{}).Error
		if err != nil {
			return err
		}
	}
	return nil
}

// attributedBlobIDs lists the blob IDs currently attributed to the scope.
func (s *blobAttributionRepository) attributedBlobIDs(ctx context.Context, scopeType, scopeID string) ([]string, error) {
	var blobIDs []string
	err := s.tx.UnderlyingDB().WithContext(ctx).
		Table(blobAttributionTable).
		Where("scope_type = ? AND scope_id = ?", scopeType, scopeID).
		Pluck("blob_id", &blobIDs).Error
	return blobIDs, err
}

// reachableBlobs returns the distinct blobs reachable from the scope.
func (s *blobAttributionRepository) reachableBlobs(ctx context.Context, scopeType, scopeID string) ([]*models.Blob, error) {
	db := s.tx.UnderlyingDB().WithContext(ctx).
		Table("blobs AS b").
		Select("DISTINCT b.id AS id, b.size AS size").
		Joins("JOIN artifact_blobs AS ab ON ab.blob_id = b.id").
		Joins("JOIN artifacts AS a ON a.id = ab.artifact_id AND a.deleted_at = 0").
		Where("b.deleted_at = 0")

	switch scopeType {
	case ScopeTypeNamespace:
		db = db.
			Joins("JOIN repositories AS r ON r.id = a.repository_id AND r.deleted_at = 0").
			Where("r.namespace_id = ?", scopeID)
	case ScopeTypeRepository:
		db = db.Where("a.repository_id = ?", scopeID)
	default:
		return nil, fmt.Errorf("invalid quota scope type: %s", scopeType)
	}

	var blobs []*models.Blob
	if err := db.Scan(&blobs).Error; err != nil {
		return nil, err
	}
	return blobs, nil
}
