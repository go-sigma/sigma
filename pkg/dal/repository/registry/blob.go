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

package registry

import (
	"context"
	"errors"
	"time"

	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
)

//go:generate go tool mockgen -destination=blob_mocks.go -package=registry github.com/go-sigma/sigma/pkg/dal/repository/registry BlobRepository

// BlobRepository defines blob repository operations
type BlobRepository interface {
	// Create creates a new blob
	Create(ctx context.Context, blob *models.Blob) error
	// UploadCreate creates a new blob upload
	UploadCreate(ctx context.Context, blobUpload *models.BlobUpload) error
	// UploadGetLastPart gets the last uploaded part with the specified upload ID
	UploadGetLastPart(ctx context.Context, uploadID string) (*models.BlobUpload, error)
	// UploadFindAllByUploadID finds all blob upload parts with the specified upload ID
	UploadFindAllByUploadID(ctx context.Context, uploadID string) ([]*models.BlobUpload, error)
	// UploadTotalSizeByUploadID gets the total uploaded size with the specified upload ID
	UploadTotalSizeByUploadID(ctx context.Context, uploadID string) (int64, error)
	// UploadTotalEtagsByUploadID gets uploaded part etags with the specified upload ID
	UploadTotalEtagsByUploadID(ctx context.Context, uploadID string) ([]string, error)
	// UploadDeleteByUploadID deletes all blob upload parts with the specified upload ID
	UploadDeleteByUploadID(ctx context.Context, uploadID string) error
	// FindWithLastPull finds blobs older than the given pull or update time with cursor pagination
	FindWithLastPull(ctx context.Context, before int64, last string, limit int64) ([]*models.Blob, error)
	// FindDeletableWithCursor finds blobs that are not referenced by any live artifact with cursor pagination
	FindDeletableWithCursor(ctx context.Context, before int64, last string, limit int64) ([]*models.Blob, error)
	// FindAssociateWithArtifact finds blob IDs that are still associated with live artifacts
	FindAssociateWithArtifact(ctx context.Context, ids []string) ([]string, error)
	// FindByDigest finds the blob with the specified digest
	FindByDigest(ctx context.Context, digest string) (*models.Blob, error)
	// FindByDigests finds blobs with the specified digests
	FindByDigests(ctx context.Context, digests []string) ([]*models.Blob, error)
	// Exists checks whether the blob with the specified digest exists
	Exists(ctx context.Context, digest string) (bool, error)
	// Incr increases the blob pull count
	Incr(ctx context.Context, id string) error
	// DeleteByID deletes the blob with the specified blob ID
	DeleteByID(ctx context.Context, id string) error
}

var _ BlobRepository = &blobRepository{}

type blobRepository struct {
	tx *query.Query
}

// NewBlobRepository creates a new blob repository with the optional query transaction
func NewBlobRepository(txs ...*query.Query) BlobRepository {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &blobRepository{
		tx: tx,
	}
}

// Create creates a new blob
func (s *blobRepository) Create(ctx context.Context, blob *models.Blob) error {
	return s.tx.Blob.WithContext(ctx).Create(blob)
}

// UploadCreate creates a new blob upload
func (s *blobRepository) UploadCreate(ctx context.Context, blobUpload *models.BlobUpload) error {
	return s.tx.BlobUpload.WithContext(ctx).Create(blobUpload)
}

// UploadGetLastPart gets the last uploaded part with the specified upload ID
func (s *blobRepository) UploadGetLastPart(ctx context.Context, uploadID string) (*models.BlobUpload, error) {
	return s.tx.BlobUpload.WithContext(ctx).
		Where(s.tx.BlobUpload.UploadID.Eq(uploadID)).
		Order(s.tx.BlobUpload.PartNumber.Desc()).First()
}

// UploadFindAllByUploadID finds all blob upload parts with the specified upload ID
func (s *blobRepository) UploadFindAllByUploadID(ctx context.Context, uploadID string) ([]*models.BlobUpload, error) {
	return s.tx.BlobUpload.WithContext(ctx).
		Where(s.tx.BlobUpload.UploadID.Eq(uploadID)).
		Order(s.tx.BlobUpload.PartNumber).Find()
}

// UploadTotalSizeByUploadID gets the total uploaded size with the specified upload ID
func (s *blobRepository) UploadTotalSizeByUploadID(ctx context.Context, uploadID string) (int64, error) {
	blobUploads, err := s.UploadFindAllByUploadID(ctx, uploadID)
	if err != nil {
		return 0, err
	}
	var totalSize int64
	for _, blobUpload := range blobUploads {
		totalSize += blobUpload.Size
	}
	return totalSize, nil
}

// UploadTotalEtagsByUploadID gets uploaded part etags with the specified upload ID
func (s *blobRepository) UploadTotalEtagsByUploadID(ctx context.Context, uploadID string) ([]string, error) {
	blobUploads, err := s.UploadFindAllByUploadID(ctx, uploadID)
	if err != nil {
		return nil, err
	}
	etags := make([]string, 0, len(blobUploads))
	for _, blobUpload := range blobUploads {
		etags = append(etags, blobUpload.Etag)
	}
	if len(etags) == 1 {
		return []string{}, nil
	}
	return etags[1:], nil
}

// UploadDeleteByUploadID deletes all blob upload parts with the specified upload ID
func (s *blobRepository) UploadDeleteByUploadID(ctx context.Context, uploadID string) error {
	_, err := s.tx.BlobUpload.WithContext(ctx).
		Where(s.tx.BlobUpload.UploadID.Eq(uploadID)).
		Delete()
	return err
}

// FindWithLastPull finds blobs older than the given pull or update time with cursor pagination
func (s *blobRepository) FindWithLastPull(ctx context.Context, before int64, last string, limit int64) ([]*models.Blob, error) {
	return s.tx.Blob.WithContext(ctx).
		Where(s.tx.Blob.ID.Gt(last)).
		Where(s.tx.Blob.LastPull.Lt(before)).
		Or(s.tx.Blob.LastPull.IsNull(), s.tx.Blob.UpdatedAt.Lt(before)).
		Limit(int(limit)).Order(s.tx.Blob.ID).Find()
}

// FindDeletableWithCursor finds blobs that are not referenced by any live artifact with cursor pagination
func (s *blobRepository) FindDeletableWithCursor(ctx context.Context, before int64, last string, limit int64) ([]*models.Blob, error) {
	artifactBlobs := artifactBlobsJoinTable{}
	artifactID := field.NewString(artifactBlobs.TableName(), "artifact_id")
	blobID := field.NewString(artifactBlobs.TableName(), "blob_id")
	associateSubQuery := s.tx.Artifact.WithContext(ctx).
		Select(s.tx.Artifact.ID).
		Join(artifactBlobs, s.tx.Artifact.ID.EqCol(artifactID)).
		Where(blobID.EqCol(s.tx.Blob.ID))

	return s.tx.Blob.WithContext(ctx).
		Where(s.tx.Blob.ID.Gt(last)).
		Where(field.Or(
			s.tx.Blob.LastPull.Lt(before),
			field.And(s.tx.Blob.LastPull.IsNull(), s.tx.Blob.UpdatedAt.Lt(before)),
		)).
		Not(gen.Exists(associateSubQuery)).
		Limit(int(limit)).
		Order(s.tx.Blob.ID).
		Find()
}

// FindAssociateWithArtifact finds blob IDs that are still associated with live artifacts
func (s *blobRepository) FindAssociateWithArtifact(ctx context.Context, ids []string) ([]string, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	artifactBlobs := artifactBlobsJoinTable{}
	artifactID := field.NewString(artifactBlobs.TableName(), "artifact_id")
	blobID := field.NewString(artifactBlobs.TableName(), "blob_id")
	associateSubQuery := s.tx.Artifact.WithContext(ctx).
		Select(s.tx.Artifact.ID).
		Join(artifactBlobs, s.tx.Artifact.ID.EqCol(artifactID)).
		Where(blobID.EqCol(s.tx.Blob.ID))

	blobs, err := s.tx.Blob.WithContext(ctx).
		Where(s.tx.Blob.ID.In(ids...)).
		Where(gen.Exists(associateSubQuery)).
		Find()
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(blobs))
	for _, blob := range blobs {
		result = append(result, blob.ID)
	}
	return result, err
}

// FindByDigest finds the blob with the specified digest
func (s *blobRepository) FindByDigest(ctx context.Context, digest string) (*models.Blob, error) {
	return s.tx.Blob.WithContext(ctx).Where(s.tx.Blob.Digest.Eq(digest)).First()
}

// FindByDigests finds blobs with the specified digests
func (s *blobRepository) FindByDigests(ctx context.Context, digests []string) ([]*models.Blob, error) {
	return s.tx.Blob.WithContext(ctx).Where(s.tx.Blob.Digest.In(digests...)).Find()
}

// Exists checks whether the blob with the specified digest exists
func (s *blobRepository) Exists(ctx context.Context, digest string) (bool, error) {
	blob, err := s.tx.Blob.WithContext(ctx).Where(s.tx.Blob.Digest.Eq(digest)).First()
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return blob != nil, err
}

// Incr increases the blob pull count
func (s *blobRepository) Incr(ctx context.Context, id string) error {
	_, err := s.tx.Blob.WithContext(ctx).Where(s.tx.Blob.ID.Eq(id)).
		UpdateColumns(map[string]any{
			"pull_times": gorm.Expr("pull_times + ?", 1),
			"last_pull":  time.Now().UnixMilli(),
		})
	return err
}

// DeleteByID deletes the blob with the specified blob ID
func (s *blobRepository) DeleteByID(ctx context.Context, id string) error {
	matched, err := s.tx.Blob.WithContext(ctx).Where(s.tx.Blob.ID.Eq(id)).Delete()
	if err != nil {
		return err
	}
	if matched.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
