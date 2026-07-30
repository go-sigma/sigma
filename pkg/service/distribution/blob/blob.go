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

package blob

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/opencontainers/go-digest"
	"go.uber.org/dig"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	reponamespace "github.com/go-sigma/sigma/pkg/dal/repository/namespace"
	reporegistry "github.com/go-sigma/sigma/pkg/dal/repository/registry"
	"github.com/go-sigma/sigma/pkg/distribution/clients"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/storage"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

//go:generate mockgen -mock_names Service=MockDistributionBlobService -destination=blob_mocks.go -package=blob github.com/go-sigma/sigma/pkg/service/distribution/blob Service

// Service encapsulates distribution blob business logic.
type Service interface {
	// GetNamespaceByName gets a namespace by name (used by handler for auth check).
	GetNamespaceByName(ctx context.Context, name string) (*models.Namespace, error)
	// DeleteBlob deletes a blob (includes: find blob, check deletable, delete from DB).
	DeleteBlob(ctx context.Context, namespaceID string, digestStr string, userID string) error
	// HeadBlob gets blob metadata without content.
	HeadBlob(ctx context.Context, namespaceID string, digestStr string, requestMethod string, requestPath string) (*models.Blob, error)
	// GetBlob gets blob metadata and content reader.
	// Returns (reader, blob, redirectURL, error). If redirectURL is non-empty,
	// the handler should issue an HTTP redirect instead of streaming the reader.
	GetBlob(ctx context.Context, namespaceID string, digestStr string, requestMethod string, requestPath string) (io.ReadCloser, *models.Blob, string, error)
}

type service struct {
	dig.In

	Config       *config.Configuration
	RepoBlob     reporegistry.BlobRepository
	RepoNs       reponamespace.NamespaceRepository
	RepoRegistry reporegistry.RepositoryRepository
	Storage      storage.StorageDriver
}

func NewService(params service) Service {
	return &params
}

// GetNamespaceByName gets a namespace by name.
func (s *service) GetNamespaceByName(ctx context.Context, name string) (*models.Namespace, error) {
	return s.RepoNs.GetByName(ctx, name)
}

// DeleteBlob deletes a blob (includes: find blob, check deletable, delete from DB).
func (s *service) DeleteBlob(ctx context.Context, namespaceID string, digestStr string, userID string) error {
	blobObj, err := s.RepoBlob.FindByDigest(ctx, digestStr)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("blob not found", "err", err, "digest", digestStr)
			return errcode.DSErrCodeBlobUnknown
		}
		slog.Error("find blob failed", "err", err, "digest", digestStr)
		return errcode.DSErrCodeUnknown
	}
	result, err := s.RepoBlob.FindAssociateWithArtifact(ctx, []string{blobObj.ID})
	if err != nil {
		slog.Error("find associate with artifact failed", "err", err, "digest", digestStr)
		return errcode.DSErrCodeUnknown
	}
	if len(result) > 0 {
		slog.Error("blob associate with artifact", "digest", digestStr)
		return errcode.DSErrCodeBlobAssociated
	}
	err = s.RepoBlob.DeleteByID(ctx, blobObj.ID)
	if err != nil {
		slog.Error("delete blob failed", "err", err, "digest", digestStr)
		return errcode.DSErrCodeUnknown
	}
	return nil
}

// HeadBlob gets blob metadata without content.
func (s *service) HeadBlob(ctx context.Context, namespaceID string, digestStr string, requestMethod string, requestPath string) (*models.Blob, error) {
	blobObj, err := s.RepoBlob.FindByDigest(ctx, digestStr)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) && s.Config.Proxy.Enabled {
			f := clients.NewClientsFactory()
			cli, err := f.New(s.Config)
			if err != nil {
				slog.Error("new proxy server failed", "err", err, "digest", digestStr)
				return nil, errcode.DSErrCodeUnknown
			}
			statusCode, header, _, err := cli.DoRequest(ctx, requestMethod, requestPath, nil)
			if err != nil {
				slog.Error("request proxy server failed", "err", err, "digest", digestStr)
				return nil, errcode.DSErrCodeUnknown
			}
			if statusCode != http.StatusOK {
				slog.Error("request proxy server failed", "statusCode", statusCode, "digest", digestStr)
				return nil, errcode.DSErrCodeUnknown
			}
			contentLength, err := strconv.ParseInt(header.Get(consts.HeaderContentLength), 10, 64)
			if err != nil {
				slog.Error("parse content length failed", "err", err, "digest", digestStr)
				return nil, errcode.DSErrCodeUnknown
			}
			return &models.Blob{
				ID:          uuid.NewV7String(),
				Digest:      digestStr,
				Size:        contentLength,
				ContentType: header.Get(consts.HeaderContentType),
			}, nil
		}
		slog.Error("check blob exist failed", "err", err, "digest", digestStr)
		return nil, errcode.DSErrCodeBlobUnknown
	}
	return blobObj, nil
}

// GetBlob gets blob metadata and content reader.
func (s *service) GetBlob(ctx context.Context, namespaceID string, digestStr string, requestMethod string, requestPath string) (io.ReadCloser, *models.Blob, string, error) {
	dgest, err := digest.Parse(digestStr)
	if err != nil {
		slog.Error("parse digest failed", "err", err, "digest", digestStr)
		return nil, nil, "", errcode.DSErrCodeDigestInvalid
	}

	blob, err := s.RepoBlob.FindByDigest(ctx, digestStr)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) && s.Config.Proxy.Enabled {
			f := clients.NewClientsFactory()
			cli, err := f.New(s.Config)
			if err != nil {
				slog.Error("new proxy server failed", "err", err, "digest", digestStr)
				return nil, nil, "", errcode.DSErrCodeUnknown
			}
			statusCode, header, bodyReader, err := cli.DoRequest(ctx, requestMethod, requestPath, nil)
			if err != nil {
				slog.Error("request proxy server failed", "err", err, "digest", digestStr)
				return nil, nil, "", errcode.DSErrCodeUnknown
			}
			if statusCode != http.StatusOK {
				slog.Error("request proxy server failed", "statusCode", statusCode, "digest", digestStr)
				return nil, nil, "", errcode.DSErrCodeUnknown
			}
			contentType := header.Get(consts.HeaderContentType)
			blobSize, err := strconv.ParseInt(header.Get(consts.HeaderContentLength), 10, 64)
			if err != nil {
				slog.Error("parse content length failed", "err", err, "digest", digestStr)
				return nil, nil, "", errcode.DSErrCodeUnknown
			}
			slog.Info("proxy blob", "digest", digestStr, "length", blobSize)
			pipeReader, pipeWriter := io.Pipe()
			newBodyReader := io.TeeReader(bodyReader, pipeWriter)
			go func() {
				uploadCtx := context.WithoutCancel(ctx)
				uploadErr := s.Storage.Upload(uploadCtx, utils.GenBlobPathByDigest(dgest), io.LimitReader(pipeReader, blobSize))
				if uploadErr != nil {
					slog.Error("upload blob failed", "err", uploadErr, "digest", digestStr)
					return
				}
				// Note: the blob exist in the storage, but not in the database,
				// so gc should delete the file directly.
				createErr := s.RepoBlob.Create(uploadCtx, &models.Blob{ID: uuid.NewV7String(), Digest: digestStr, Size: blobSize, ContentType: contentType, PushedAt: time.Now().UnixMilli()})
				if createErr != nil {
					slog.Error("create blob failed", "err", createErr, "digest", digestStr)
					return
				}
			}()
			blob = &models.Blob{
				ID:          uuid.NewV7String(),
				Digest:      digestStr,
				Size:        blobSize,
				ContentType: contentType,
			}
			return &proxyReadCloser{Reader: newBodyReader, body: bodyReader}, blob, "", nil
		}
		slog.Error("check blob exist failed", "err", err, "digest", digestStr)
		return nil, nil, "", errcode.DSErrCodeBlobUnknown
	}

	if s.Config.Storage.Redirect && s.Config.Storage.Type != enums.StorageTypeFilesystem {
		redirectUrl, err := s.Storage.Redirect(ctx, utils.GenBlobPathByDigest(dgest))
		if err != nil {
			slog.Error("get blob redirect url failed", "err", err, "digest", digestStr)
			return nil, nil, "", errcode.DSErrCodeUnknown
		}
		return nil, blob, redirectUrl, nil
	}

	reader, err := s.Storage.Reader(ctx, utils.GenBlobPathByDigest(dgest))
	if err != nil {
		slog.Error("get blob reader failed", "err", err, "digest", digestStr)
		return nil, nil, "", errcode.DSErrCodeUnknown
	}
	return reader, blob, "", nil
}

// proxyReadCloser wraps a TeeReader and the underlying body reader so that
// closing it also closes the proxied response body.
type proxyReadCloser struct {
	io.Reader
	body io.ReadCloser
}

func (p *proxyReadCloser) Close() error {
	return p.body.Close()
}
