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

package upload

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/distribution/reference"
	"github.com/opencontainers/go-digest"
	"go.uber.org/dig"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	reponamespace "github.com/go-sigma/sigma/pkg/dal/repository/namespace"
	reporegistry "github.com/go-sigma/sigma/pkg/dal/repository/registry"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/storage"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/counter"
	"github.com/go-sigma/sigma/pkg/utils/hash"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

//go:generate go tool mockgen -mock_names Service=MockDistributionUploadService -destination=upload_mocks.go -package=upload github.com/go-sigma/sigma/pkg/service/distribution/upload Service

// Service encapsulates distribution blob upload business logic.
type Service interface {
	// GetNamespaceByName gets a namespace by name (used by handler for auth check).
	GetNamespaceByName(ctx context.Context, name string) (*models.Namespace, error)
	// PostUpload starts a new upload session. If digestStr is non-empty, performs a single-shot upload.
	// Returns the upload ID.
	PostUpload(ctx context.Context, repositoryName string, digestStr string, body io.Reader, contentType string) (string, error)
	// PatchUpload appends data to an upload session. Returns (sizeBefore, sizeUploaded, error).
	PatchUpload(ctx context.Context, uploadID string, body io.Reader, repositoryName string) (int64, int64, error)
	// PutUpload completes an upload session and commits the blob. Returns (sizeBefore, sizeUploaded, error).
	PutUpload(ctx context.Context, uploadID string, digestStr string, body io.Reader, repositoryName string, contentType string, contentLength int64) (int64, int64, error)
	// GetUpload gets upload status.
	GetUpload(ctx context.Context, uploadID string) (*models.BlobUpload, error)
	// DeleteUpload cancels an upload session.
	DeleteUpload(ctx context.Context, uploadID string) error
}

type service struct {
	dig.In

	RepoBlob     reporegistry.BlobRepository
	RepoRegistry reporegistry.RepositoryRepository
	RepoNs       reponamespace.NamespaceRepository
	Storage      storage.StorageDriver
}

func NewService(params service) Service {
	return &params
}

// GetNamespaceByName gets a namespace by name.
func (s *service) GetNamespaceByName(ctx context.Context, name string) (*models.Namespace, error) {
	return s.RepoNs.GetByName(ctx, name)
}

// PostUpload starts a new upload session. If digestStr is non-empty, performs a single-shot upload.
func (s *service) PostUpload(ctx context.Context, repositoryName string, digestStr string, body io.Reader, contentType string) (string, error) {
	// fileID is the filename that upload to the blob_uploads
	fileID := uuid.NewV7String()

	// according to the docker registry api, if the digest is provided, the upload is complete
	if digestStr != "" {
		dgest, err := digest.Parse(digestStr)
		if err != nil {
			slog.Error("parse digest failed", "err", err, "digest", digestStr)
			return "", errcode.DSErrCodeBlobUploadInvalid
		}

		countReader := counter.NewCounter(body)

		srcPath := fmt.Sprintf("%s/%s", consts.BlobUploads, fileID)
		err = s.Storage.Upload(ctx, srcPath, countReader)
		if err != nil {
			slog.Error("upload blob failed", "err", err)
			return "", errcode.DSErrCodeBlobUploadInvalid
		}
		destPath := utils.GenBlobPathByDigest(dgest)
		err = s.Storage.Move(ctx, srcPath, destPath)
		if err != nil {
			slog.Error("move blob failed", "err", err)
			return "", errcode.DSErrCodeUnknown
		}

		err = s.Storage.Delete(ctx, srcPath)
		if err != nil {
			slog.Error("delete blob upload failed", "err", err)
			return "", errcode.DSErrCodeUnknown
		}

		size := countReader.Count()

		err = s.RepoBlob.Create(ctx, &models.Blob{
			ID:          uuid.NewV7String(),
			Digest:      dgest.String(),
			Size:        size,
			ContentType: contentType,
		})
		if err != nil {
			slog.Error("save blob record failed", "err", err)
			return "", errcode.DSErrCodeUnknown
		}
	}

	uploadID, err := s.Storage.CreateUploadID(ctx, fmt.Sprintf("%s/%s", consts.BlobUploads, fileID))
	if err != nil {
		slog.Error("create blob upload id failed", "err", err)
		return "", errcode.DSErrCodeUnknown
	}

	err = s.RepoBlob.UploadCreate(ctx, &models.BlobUpload{
		ID:         uuid.NewV7String(),
		PartNumber: 0,
		UploadID:   uploadID,
		Etag:       "fake",
		Repository: repositoryName,
		FileID:     fileID,
	})
	if err != nil {
		slog.Error("save blob upload record failed", "err", err)
		return "", errcode.DSErrCodeUnknown
	}

	return uploadID, nil
}

// PatchUpload appends data to an upload session. Returns (sizeBefore, sizeUploaded, error).
func (s *service) PatchUpload(ctx context.Context, uploadID string, body io.Reader, repositoryName string) (int64, int64, error) {
	uploadObj, err := s.RepoBlob.UploadGetLastPart(ctx, uploadID)
	if err != nil {
		slog.Error("get blob upload record failed", "err", err)
		return 0, 0, errcode.DSErrCodeUnknown
	}

	sizeBefore, err := s.RepoBlob.UploadTotalSizeByUploadID(ctx, uploadID)
	if err != nil {
		slog.Error("get blob upload record failed", "err", err)
		return 0, 0, errcode.DSErrCodeUnknown
	}

	counterReader := counter.NewCounter(body)

	path := fmt.Sprintf("%s/%s", consts.BlobUploads, uploadObj.FileID)
	etag, err := s.Storage.UploadPart(ctx, path, uploadObj.UploadID, uploadObj.PartNumber+1, counterReader)
	if err != nil {
		slog.Error("upload part failed", "err", err)
		return 0, 0, errcode.DSErrCodeUnknown
	}

	size := counterReader.Count()
	err = s.RepoBlob.UploadCreate(ctx, &models.BlobUpload{
		ID:         uuid.NewV7String(),
		PartNumber: uploadObj.PartNumber + 1,
		UploadID:   uploadID,
		Etag:       strings.Trim(etag, "\""),
		Repository: repositoryName,
		FileID:     uploadObj.FileID,
		Size:       size,
	})
	if err != nil {
		slog.Error("save blob upload record failed", "err", err)
		return 0, 0, errcode.DSErrCodeUnknown
	}

	return sizeBefore, size, nil
}

// PutUpload completes an upload session and commits the blob. Returns (sizeBefore, sizeUploaded, error).
func (s *service) PutUpload(ctx context.Context, uploadID string, digestStr string, body io.Reader, repositoryName string, contentType string, contentLength int64) (int64, int64, error) {
	dgest, err := digest.Parse(digestStr)
	if err != nil {
		slog.Error("parse digest failed", "err", err, "digest", digestStr)
		return 0, 0, errcode.DSErrCodeDigestInvalid
	}

	uploadObj, err := s.RepoBlob.UploadGetLastPart(ctx, uploadID)
	if err != nil {
		slog.Error("get blob upload record failed", "err", err)
		return 0, 0, errcode.DSErrCodeUnknown
	}
	srcPath := fmt.Sprintf("%s/%s", consts.BlobUploads, uploadObj.FileID)

	exist, err := s.RepoBlob.Exists(ctx, dgest.String())
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.Error("check blob exist failed", "err", err, "digest", dgest.String())
		return 0, 0, errcode.DSErrCodeUnknown
	}
	if exist {
		err = s.Storage.AbortUpload(ctx, srcPath, uploadObj.UploadID)
		if err != nil {
			slog.Error("abort upload failed", "err", err)
			return 0, 0, errcode.DSErrCodeUnknown
		}
		return 0, 0, nil
	}

	if !reference.NameRegexp.MatchString(repositoryName) {
		slog.Error("invalid repository name", "repository", repositoryName)
		return 0, 0, errcode.DSErrCodeNameInvalid
	}

	etags, err := s.RepoBlob.UploadTotalEtagsByUploadID(ctx, uploadID)
	if err != nil {
		slog.Error("get blob upload etags failed", "err", err)
		return 0, 0, errcode.DSErrCodeUnknown
	}

	sizeBefore, err := s.RepoBlob.UploadTotalSizeByUploadID(ctx, uploadID)
	if err != nil {
		slog.Error("get blob upload size failed", "err", err)
		return 0, 0, errcode.DSErrCodeUnknown
	}

	var sizeUploaded int64
	if contentLength != 0 {
		counterReader := counter.NewCounter(body)
		etag, err := s.Storage.UploadPart(ctx, srcPath, uploadObj.UploadID, uploadObj.PartNumber+1, counterReader)
		if err != nil {
			slog.Error("upload part failed", "err", err, "uploadID", uploadObj.UploadID)
			return 0, 0, errcode.DSErrCodeUnknown
		}
		size := counterReader.Count()
		etags = append(etags, strings.Trim(etag, "\""))
		err = s.RepoBlob.UploadCreate(ctx, &models.BlobUpload{
			ID:         uuid.NewV7String(),
			PartNumber: uploadObj.PartNumber + 1,
			UploadID:   uploadID,
			Etag:       strings.Trim(etag, "\""),
			Repository: repositoryName,
			FileID:     uploadObj.FileID,
			Size:       size,
		})
		if err != nil {
			slog.Error("create blob upload record failed", "err", err)
			return 0, 0, errcode.DSErrCodeUnknown
		}
		sizeUploaded = size
	}

	slog.Info("committing upload", "uploadID", uploadID, "etags", etags)
	err = s.Storage.CommitUpload(ctx, srcPath, uploadID, etags)
	if err != nil {
		slog.Error("commit upload failed", "err", err, "id", uploadID, "etags", etags)
		return 0, 0, errcode.DSErrCodeUnknown
	}

	srcPathReader, err := s.Storage.Reader(ctx, srcPath)
	if err != nil {
		slog.Error("get blob upload failed", "err", err, "srcPath", srcPath)
		return 0, 0, errcode.DSErrCodeUnknown
	}
	srcPathHash, err := hash.Reader(srcPathReader, dgest.Algorithm().String())
	if err != nil {
		slog.Error("hash blob upload failed", "err", err, "srcPath", srcPath)
		return 0, 0, errcode.DSErrCodeUnknown
	}
	if fmt.Sprintf("%s:%s", dgest.Algorithm().String(), srcPathHash) != dgest.String() {
		slog.Error("hash blob upload mismatch", "srcPath", srcPath, "srcPathHash", fmt.Sprintf("%s:%s", dgest.Algorithm().String(), srcPathHash), "targetHash", dgest.String())
		return 0, 0, errcode.DSErrCodeBlobUploadDigestMismatch
	}

	destPath := utils.GenBlobPathByDigest(dgest)
	err = s.Storage.Move(ctx, srcPath, destPath)
	if err != nil {
		slog.Error("move blob failed", "err", err, "path", srcPath, "digest", dgest.String(), "dest", destPath)
		return 0, 0, errcode.DSErrCodeUnknown
	}

	err = s.Storage.Delete(ctx, srcPath)
	if err != nil {
		slog.Error("delete blob upload failed", "err", err)
		return 0, 0, errcode.DSErrCodeUnknown
	}

	err = s.RepoBlob.UploadDeleteByUploadID(ctx, uploadID)
	if err != nil {
		slog.Error("delete blob upload record failed", "err", err)
		return 0, 0, errcode.DSErrCodeUnknown
	}

	err = s.RepoBlob.Create(ctx, &models.Blob{
		ID:          uuid.NewV7String(),
		Digest:      dgest.String(),
		Size:        sizeBefore + contentLength,
		ContentType: contentType,
		PushedAt:    time.Now().UnixMilli(),
	})
	if err != nil {
		slog.Error("create blob record failed", "err", err)
		return 0, 0, errcode.DSErrCodeUnknown
	}

	return sizeBefore, sizeUploaded, nil
}

// GetUpload gets upload status.
func (s *service) GetUpload(ctx context.Context, uploadID string) (*models.BlobUpload, error) {
	return s.RepoBlob.UploadGetLastPart(ctx, uploadID)
}

// DeleteUpload cancels an upload session.
func (s *service) DeleteUpload(ctx context.Context, uploadID string) error {
	uploadObj, err := s.RepoBlob.UploadGetLastPart(ctx, uploadID)
	if err != nil {
		slog.Error("get blob upload record failed", "err", err)
		return errcode.DSErrCodeUnknown
	}
	srcPath := fmt.Sprintf("%s/%s", consts.BlobUploads, uploadObj.FileID)
	err = s.Storage.AbortUpload(ctx, srcPath, uploadObj.UploadID)
	if err != nil {
		slog.Error("abort upload failed", "err", err)
		return errcode.DSErrCodeUnknown
	}
	err = s.RepoBlob.UploadDeleteByUploadID(ctx, uploadID)
	if err != nil {
		slog.Error("delete blob upload record failed", "err", err)
		return errcode.DSErrCodeUnknown
	}
	return nil
}
