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
	"fmt"

	"github.com/ximager/ximager/pkg/dal/models"
	"github.com/ximager/ximager/pkg/dal/query"
)

// BlobUploadService is the interface for the blob upload service.
type BlobUploadService interface {
	// Create creates a new blob upload.
	Create(ctx context.Context, blobUpload *models.BlobUpload) (*models.BlobUpload, error)
	// Get gets the blob upload with the specified blob upload ID.
	GetLastPart(ctx context.Context, uploadID string) (*models.BlobUpload, error)
	// FindAllByUploadID find all blob uploads with the specified upload ID.
	FindAllByUploadID(ctx context.Context, uploadID string) ([]*models.BlobUpload, error)
	// TotalSizeByUploadID gets the total size of the blob uploads with the specified upload ID.
	TotalSizeByUploadID(ctx context.Context, uploadID string) (int64, error)
	// TotalEtagsByUploadID gets the total etags of the blob uploads with the specified upload ID.
	TotalEtagsByUploadID(ctx context.Context, uploadID string) ([]string, error)
	// DeleteByUploadID deletes all blob uploads with the specified upload ID.
	DeleteByUploadID(ctx context.Context, uploadID string) error
}

var _ BlobUploadService = &blobUploadService{}

type blobUploadService struct {
	tx *query.Query
}

// NewBlobUploadService creates a new blob upload service.
func NewBlobUploadService(txs ...*query.Query) BlobUploadService {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &blobUploadService{
		tx: tx,
	}
}

// Create creates a new blob upload.
func (b *blobUploadService) Create(ctx context.Context, blobUpload *models.BlobUpload) (*models.BlobUpload, error) {
	err := b.tx.BlobUpload.WithContext(ctx).Create(blobUpload)
	if err != nil {
		return nil, err
	}
	return blobUpload, nil
}

// GetLastPart gets the blob upload with the specified blob upload ID.
func (b *blobUploadService) GetLastPart(ctx context.Context, uploadID string) (*models.BlobUpload, error) {
	blobUpload, err := b.tx.BlobUpload.WithContext(ctx).
		Where(b.tx.BlobUpload.UploadID.Eq(uploadID)).
		Order(b.tx.BlobUpload.PartNumber.Desc()).First()
	if err != nil {
		return nil, err
	}
	return blobUpload, nil
}

// FindAllByUploadID find all blob uploads with the specified upload ID.
func (b *blobUploadService) FindAllByUploadID(ctx context.Context, uploadID string) ([]*models.BlobUpload, error) {
	blobUploads, err := b.tx.BlobUpload.WithContext(ctx).
		Where(b.tx.BlobUpload.UploadID.Eq(uploadID)).
		Order(b.tx.BlobUpload.PartNumber).Find()
	if err != nil {
		return nil, err
	}
	return blobUploads, nil
}

// TotalSizeByUploadID gets the total size of the blob uploads with the specified upload ID.
func (b *blobUploadService) TotalSizeByUploadID(ctx context.Context, uploadID string) (int64, error) {
	blobUploads, err := b.FindAllByUploadID(ctx, uploadID)
	if err != nil {
		return 0, err
	}
	var totalSize int64
	for _, blobUpload := range blobUploads {
		totalSize += blobUpload.Size
	}
	return totalSize, nil
}

// TotalEtagsByUploadID gets the total etags of the blob uploads with the specified upload ID.
func (b *blobUploadService) TotalEtagsByUploadID(ctx context.Context, uploadID string) ([]string, error) {
	blobUploads, err := b.FindAllByUploadID(ctx, uploadID)
	if err != nil {
		return nil, err
	}
	var etags []string
	for _, blobUpload := range blobUploads {
		etags = append(etags, blobUpload.Etag)
	}
	if len(etags) == 1 {
		return nil, fmt.Errorf("cannot find valid etags")
	}
	return etags[1:], nil
}

// DeleteByUploadID deletes all blob uploads with the specified upload ID.
func (b *blobUploadService) DeleteByUploadID(ctx context.Context, uploadID string) error {
	_, err := b.tx.BlobUpload.WithContext(ctx).
		Where(b.tx.BlobUpload.UploadID.Eq(uploadID)).
		Delete()
	return err
}
