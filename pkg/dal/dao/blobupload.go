// Copyright 2023 XImager
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
	"fmt"

	"github.com/ximager/ximager/pkg/dal/models"
	"github.com/ximager/ximager/pkg/dal/query"
)

//go:generate mockgen -destination=mocks/blobupload.go -package=mocks github.com/ximager/ximager/pkg/dal/dao BlobUploadService
//go:generate mockgen -destination=mocks/blobupload_factory.go -package=mocks github.com/ximager/ximager/pkg/dal/dao BlobUploadServiceFactory

// BlobUploadService is the interface for the blob upload service.
type BlobUploadService interface {
	// Create creates a new blob upload.
	Create(ctx context.Context, blobUpload *models.BlobUpload) error
	// Get gets the blob upload with the specified blob upload ID.
	GetLastPart(ctx context.Context, uploadID string) (*models.BlobUpload, error)
	// FindAllByUploadID find all blob uploads with the specified upload ID.
	FindAllByUploadID(ctx context.Context, uploadID string) ([]*models.BlobUpload, error)
	// TotalSizeByUploadID gets the total size of the blob uploads with the specified upload ID.
	TotalSizeByUploadID(ctx context.Context, uploadID string) (uint64, error)
	// TotalEtagsByUploadID gets the total etags of the blob uploads with the specified upload ID.
	TotalEtagsByUploadID(ctx context.Context, uploadID string) ([]string, error)
	// DeleteByUploadID deletes all blob uploads with the specified upload ID.
	DeleteByUploadID(ctx context.Context, uploadID string) error
}

var _ BlobUploadService = &blobUploadService{}

type blobUploadService struct {
	tx *query.Query
}

// BlobUploadServiceFactory is the interface for the blob upload service factory.
type BlobUploadServiceFactory interface {
	New(txs ...*query.Query) BlobUploadService
}

type blobUploadServiceFactory struct{}

// NewBlobUploadServiceFactory creates a new blob upload service factory.
func NewBlobUploadServiceFactory() BlobUploadServiceFactory {
	return &blobUploadServiceFactory{}
}

func (f *blobUploadServiceFactory) New(txs ...*query.Query) BlobUploadService {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &blobUploadService{
		tx: tx,
	}
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
func (b *blobUploadService) Create(ctx context.Context, blobUpload *models.BlobUpload) error {
	return b.tx.BlobUpload.WithContext(ctx).Create(blobUpload)
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
	return b.tx.BlobUpload.WithContext(ctx).
		Where(b.tx.BlobUpload.UploadID.Eq(uploadID)).
		Order(b.tx.BlobUpload.PartNumber).Find()
}

// TotalSizeByUploadID gets the total size of the blob uploads with the specified upload ID.
func (b *blobUploadService) TotalSizeByUploadID(ctx context.Context, uploadID string) (uint64, error) {
	blobUploads, err := b.FindAllByUploadID(ctx, uploadID)
	if err != nil {
		return 0, err
	}
	var totalSize uint64
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
	var etags = make([]string, 0, len(blobUploads))
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
