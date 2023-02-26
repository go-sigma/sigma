// The MIT License (MIT)
//
// Copyright © 2023 Tosone <i@tosone.cn>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package dao

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/ximager/ximager/pkg/dal/models"
	"github.com/ximager/ximager/pkg/dal/query"
)

// BlobService defines the operations related to blobs
type BlobService interface {
	// Create creates a new blob.
	Create(ctx context.Context, blob *models.Blob) (*models.Blob, error)
	// FindByDigest finds the blob with the specified digest.
	FindByDigest(ctx context.Context, digests string) (*models.Blob, error)
	// FindByDigests finds the blobs with the specified digests.
	FindByDigests(ctx context.Context, digests []string) ([]*models.Blob, error)
	// Exists checks if the blob with the specified digest exists.
	Exists(ctx context.Context, digest string) (bool, error)
	// Incr increases the pull times of the artifact.
	Incr(ctx context.Context, id uint64) error
	// DeleteByID deletes the blob with the specified blob ID.
	DeleteByID(ctx context.Context, id uint64) error
}

var _ BlobService = &blobService{}

type blobService struct {
	tx *query.Query
}

// NewBlobService creates a new blob upload service.
func NewBlobService(txs ...*query.Query) BlobService {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &blobService{
		tx: tx,
	}
}

// Create creates a new blob.
func (b *blobService) Create(ctx context.Context, blob *models.Blob) (*models.Blob, error) {
	err := b.tx.Blob.WithContext(ctx).Create(blob)
	if err != nil {
		return nil, err
	}
	return blob, nil
}

func (b *blobService) FindByDigest(ctx context.Context, digest string) (*models.Blob, error) {
	blob, err := b.tx.Blob.WithContext(ctx).Where(b.tx.Blob.Digest.Eq(digest)).First()
	if err != nil {
		return nil, err
	}
	return blob, nil
}

// FindByDigests finds the blobs with the specified digests.
func (b *blobService) FindByDigests(ctx context.Context, digests []string) ([]*models.Blob, error) {
	blobs, err := b.tx.Blob.WithContext(ctx).Where(b.tx.Blob.Digest.In(digests...)).Find()
	if err != nil {
		return nil, err
	}
	return blobs, nil
}

// Exists checks if the blob with the specified digest exists.
func (b *blobService) Exists(ctx context.Context, digest string) (bool, error) {
	blob, err := b.tx.Blob.WithContext(ctx).Where(b.tx.Blob.Digest.Eq(digest)).First()
	if err != nil {
		return false, err
	}
	return blob != nil, nil
}

// Incr increases the pull times of the artifact.
func (s *blobService) Incr(ctx context.Context, id uint64) error {
	_, err := s.tx.Blob.WithContext(ctx).Where(s.tx.Tag.ID.Eq(id)).
		UpdateColumns(map[string]interface{}{
			"pull_times": gorm.Expr("pull_times + ?", 1),
			"last_pull":  time.Now(),
		})
	if err != nil {
		return err
	}
	return nil
}

// DeleteByID deletes the blob with the specified blob ID.
func (s *blobService) DeleteByID(ctx context.Context, id uint64) error {
	matched, err := s.tx.Blob.WithContext(ctx).Where(s.tx.Blob.ID.Eq(id)).Delete()
	if err != nil {
		return err
	}
	if matched.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
