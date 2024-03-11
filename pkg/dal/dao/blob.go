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
	"errors"
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
	FindWithLastPull(ctx context.Context, before int64, last, limit int64) ([]*models.Blob, error)
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
func (s *blobServiceFactory) New(txs ...*query.Query) BlobService {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &blobService{
		tx: tx,
	}
}

// Create creates a new blob.
func (s *blobService) Create(ctx context.Context, blob *models.Blob) error {
	return s.tx.Blob.WithContext(ctx).Create(blob)
}

// FindWithLastPull ...
func (s *blobService) FindWithLastPull(ctx context.Context, before int64, last, limit int64) ([]*models.Blob, error) {
	return s.tx.Blob.WithContext(ctx).
		Where(s.tx.Blob.ID.Gt(last)).
		Where(s.tx.Blob.LastPull.Lt(before)).
		Or(s.tx.Blob.LastPull.IsNull(), s.tx.Blob.UpdatedAt.Lt(before)).
		Order(s.tx.Blob.ID).Find()
}

// FindAssociateWithArtifact ...
func (s *blobService) FindAssociateWithArtifact(ctx context.Context, ids []int64) ([]int64, error) {
	var result []int64
	err := s.tx.Blob.WithContext(ctx).UnderlyingDB().Raw("SELECT blob_id FROM artifact_blobs LEFT JOIN artifacts ON artifacts.id = artifact_blobs.artifact_id WHERE artifacts.deleted_at = 0 AND blob_id in (?)", ids).Scan(&result).Error
	return result, err
}

// FindByDigest finds the blob with the specified digest.
func (s *blobService) FindByDigest(ctx context.Context, digest string) (*models.Blob, error) {
	return s.tx.Blob.WithContext(ctx).Where(s.tx.Blob.Digest.Eq(digest)).First()
}

// FindByDigests finds the blobs with the specified digests.
func (s *blobService) FindByDigests(ctx context.Context, digests []string) ([]*models.Blob, error) {
	return s.tx.Blob.WithContext(ctx).Where(s.tx.Blob.Digest.In(digests...)).Find()
}

// Exists checks if the blob with the specified digest exists.
func (s *blobService) Exists(ctx context.Context, digest string) (bool, error) {
	blob, err := s.tx.Blob.WithContext(ctx).Where(s.tx.Blob.Digest.Eq(digest)).First()
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return blob != nil, err
}

// Incr increases the pull times of the artifact.
func (s *blobService) Incr(ctx context.Context, id int64) error {
	_, err := s.tx.Blob.WithContext(ctx).Where(s.tx.Blob.ID.Eq(id)).
		UpdateColumns(map[string]interface{}{
			"pull_times": gorm.Expr("pull_times + ?", 1),
			"last_pull":  time.Now().UnixMilli(),
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
