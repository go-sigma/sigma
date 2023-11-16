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

	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
)

//go:generate mockgen -destination=mocks/blobupload.go -package=mocks github.com/go-sigma/sigma/pkg/dal/dao BlobUploadService
//go:generate mockgen -destination=mocks/blobupload_factory.go -package=mocks github.com/go-sigma/sigma/pkg/dal/dao BlobUploadServiceFactory

// BlobUploadService is the interface for the blob upload service.
type BlobUploadService interface {
	// Create creates a new blob upload.
	Create(ctx context.Context, blobUpload *models.BlobUpload) error
	// GetLastPart gets the blob upload with the specified blob upload ID.
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

// BlobUploadServiceFactory is the interface for the blob upload service factory.
type BlobUploadServiceFactory interface {
	New(txs ...*query.Query) BlobUploadService
}

type blobUploadServiceFactory struct{}

// NewBlobUploadServiceFactory creates a new blob upload service factory.
func NewBlobUploadServiceFactory() BlobUploadServiceFactory {
	return &blobUploadServiceFactory{}
}

// New creates a new blob upload service.
func (s *blobUploadServiceFactory) New(txs ...*query.Query) BlobUploadService {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &blobUploadService{
		tx: tx,
	}
}

// Create creates a new blob upload.
func (s *blobUploadService) Create(ctx context.Context, blobUpload *models.BlobUpload) error {
	return s.tx.BlobUpload.WithContext(ctx).Create(blobUpload)
}

// GetLastPart gets the blob upload with the specified blob upload ID.
func (s *blobUploadService) GetLastPart(ctx context.Context, uploadID string) (*models.BlobUpload, error) {
	return s.tx.BlobUpload.WithContext(ctx).
		Where(s.tx.BlobUpload.UploadID.Eq(uploadID)).
		Order(s.tx.BlobUpload.PartNumber.Desc()).First()
}

// FindAllByUploadID find all blob uploads with the specified upload ID.
func (s *blobUploadService) FindAllByUploadID(ctx context.Context, uploadID string) ([]*models.BlobUpload, error) {
	return s.tx.BlobUpload.WithContext(ctx).
		Where(s.tx.BlobUpload.UploadID.Eq(uploadID)).
		Order(s.tx.BlobUpload.PartNumber).Find()
}

// TotalSizeByUploadID gets the total size of the blob uploads with the specified upload ID.
func (s *blobUploadService) TotalSizeByUploadID(ctx context.Context, uploadID string) (int64, error) {
	blobUploads, err := s.FindAllByUploadID(ctx, uploadID)
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
func (s *blobUploadService) TotalEtagsByUploadID(ctx context.Context, uploadID string) ([]string, error) {
	blobUploads, err := s.FindAllByUploadID(ctx, uploadID)
	if err != nil {
		return nil, err
	}
	var etags = make([]string, 0, len(blobUploads))
	for _, blobUpload := range blobUploads {
		etags = append(etags, blobUpload.Etag)
	}
	if len(etags) == 1 {
		return []string{}, nil
	}
	return etags[1:], nil
}

// DeleteByUploadID deletes all blob uploads with the specified upload ID.
func (s *blobUploadService) DeleteByUploadID(ctx context.Context, uploadID string) error {
	_, err := s.tx.BlobUpload.WithContext(ctx).
		Where(s.tx.BlobUpload.UploadID.Eq(uploadID)).
		Delete()
	return err
}
