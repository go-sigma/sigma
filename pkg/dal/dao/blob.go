// Copyright 2023 sigma
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

package dao

import (
	"context"
	"database/sql"
	"time"

	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
)

//go:generate mockgen -destination=mocks/blob.go -package=mocks github.com/go-sigma/sigma/pkg/dal/dao BlobService
//go:generate mockgen -destination=mocks/blob_factory.go -package=mocks github.com/go-sigma/sigma/pkg/dal/dao BlobServiceFactory

// BlobService defines the operations related to blobs
type BlobService interface {
	// Create creates a new blob.
	Create(ctx context.Context, blob *models.Blob) error
	// FindWithLastPull find with last pull
	FindWithLastPull(ctx context.Context, before time.Time, last, limit int64) ([]*models.Blob, error)
	// FindAssociateWithArtifact ...
	FindAssociateWithArtifact(ctx context.Context, ids []int64) ([]int64, error)
	// FindByDigest finds the blob with the specified digest.
	FindByDigest(ctx context.Context, digest string) (*models.Blob, error)
	// FindByDigests finds the blobs with the specified digests.
	FindByDigests(ctx context.Context, digests []string) ([]*models.Blob, error)
	// Exists checks if the blob with the specified digest exists.
	Exists(ctx context.Context, digest string) (bool, error)
	// Incr increases the pull times of the artifact.
	Incr(ctx context.Context, id int64) error
	// DeleteByID deletes the blob with the specified blob ID.
	DeleteByID(ctx context.Context, id int64) error
}

var _ BlobService = &blobService{}

type blobService struct {
	tx *query.Query
}

// BlobServiceFactory is the interface that provides the blob service factory methods.
type BlobServiceFactory interface {
	New(txs ...*query.Query) BlobService
}

type blobServiceFactory struct{}

// NewBlobServiceFactory creates a new blob service factory.
func NewBlobServiceFactory() BlobServiceFactory {
	return &blobServiceFactory{}
}

// New creates a new blob service.
func (f *blobServiceFactory) New(txs ...*query.Query) BlobService {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &blobService{
		tx: tx,
	}
}

// Create creates a new blob.
func (b *blobService) Create(ctx context.Context, blob *models.Blob) error {
	return b.tx.Blob.WithContext(ctx).Create(blob)
}

// FindWithLastPull ...
func (b *blobService) FindWithLastPull(ctx context.Context, before time.Time, last, limit int64) ([]*models.Blob, error) {
	return b.tx.Blob.WithContext(ctx).
		Where(b.tx.Blob.LastPull.Lt(sql.NullTime{Valid: true, Time: before})).
		Or(b.tx.Blob.LastPull.IsNull(), b.tx.Blob.UpdatedAt.Lt(before)).
		Where(b.tx.Blob.ID.Gt(last)).Find()
}

// FindAssociateWithArtifact ...
func (b *blobService) FindAssociateWithArtifact(ctx context.Context, ids []int64) ([]int64, error) {
	var result []int64
	err := b.tx.Blob.WithContext(ctx).UnderlyingDB().Raw("SELECT blob_id FROM artifact_blobs WHERE blob_id in (?)", ids).Scan(&result).Error
	return result, err
}

// FindByDigest finds the blob with the specified digest.
func (b *blobService) FindByDigest(ctx context.Context, digest string) (*models.Blob, error) {
	return b.tx.Blob.WithContext(ctx).Where(b.tx.Blob.Digest.Eq(digest)).First()
}

// FindByDigests finds the blobs with the specified digests.
func (b *blobService) FindByDigests(ctx context.Context, digests []string) ([]*models.Blob, error) {
	return b.tx.Blob.WithContext(ctx).Where(b.tx.Blob.Digest.In(digests...)).Find()
}

// Exists checks if the blob with the specified digest exists.
func (b *blobService) Exists(ctx context.Context, digest string) (bool, error) {
	blob, err := b.tx.Blob.WithContext(ctx).Where(b.tx.Blob.Digest.Eq(digest)).First()
	if err != nil && err != gorm.ErrRecordNotFound {
		return false, err
	}
	return blob != nil, nil
}

// Incr increases the pull times of the artifact.
func (s *blobService) Incr(ctx context.Context, id int64) error {
	_, err := s.tx.Blob.WithContext(ctx).Where(s.tx.Blob.ID.Eq(id)).
		UpdateColumns(map[string]interface{}{
			"pull_times": gorm.Expr("pull_times + ?", 1),
			"last_pull":  time.Now(),
		})
	return err
}

// DeleteByID deletes the blob with the specified blob ID.
func (s *blobService) DeleteByID(ctx context.Context, id int64) error {
	matched, err := s.tx.Blob.WithContext(ctx).Where(s.tx.Blob.ID.Eq(id)).Delete()
	if err != nil {
		return err
	}
	if matched.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
