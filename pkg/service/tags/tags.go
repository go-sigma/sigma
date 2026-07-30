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

package tags

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/opencontainers/go-digest"
	"go.uber.org/dig"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/dal/models"
	reponamespace "github.com/go-sigma/sigma/pkg/dal/repository/namespace"
	reporegistry "github.com/go-sigma/sigma/pkg/dal/repository/registry"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/storage"
	"github.com/go-sigma/sigma/pkg/utils"
)

//go:generate mockgen -mock_names Service=MockTagService -destination=tags_mocks.go -package=tags github.com/go-sigma/sigma/pkg/service/tags Service

// Service encapsulates tag-related business logic.
type Service interface {
	// ListTags lists tags for a repository (includes: validate namespace/repository match).
	ListTags(ctx context.Context, namespaceID, repositoryID string, name *string, artifactTypes []enums.ArtifactType, pagination api.Pagination, sort api.Sortable) ([]*models.Tag, int64, error)
	// GetTag gets a tag by ID.
	GetTag(ctx context.Context, id string) (*models.Tag, error)
	// GetArtifactRaw gets the manifest raw content for an artifact digest.
	GetArtifactRaw(ctx context.Context, digestStr string) ([]byte, error)
	// DeleteTag deletes a tag by ID (includes: validate namespace/repository match).
	DeleteTag(ctx context.Context, namespaceID, repositoryID, id string) error
}

type service struct {
	dig.In

	RepoNs       reponamespace.NamespaceRepository
	RepoRegistry reporegistry.RepositoryRepository
	RepoTag      reporegistry.TagRepository
	Storage      storage.StorageDriver
}

func NewService(params service) Service {
	return &params
}

func (s *service) ListTags(ctx context.Context, namespaceID, repositoryID string, name *string, artifactTypes []enums.ArtifactType, pagination api.Pagination, sort api.Sortable) ([]*models.Tag, int64, error) {
	namespaceObj, err := s.RepoNs.Get(ctx, namespaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Namespace(%s) not found: %v", namespaceID, err))
		}
		return nil, 0, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Namespace(%s) find failed: %v", namespaceID, err))
	}

	repositoryObj, err := s.RepoRegistry.Get(ctx, repositoryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Repository(%s) not found: %v", repositoryID, err))
		}
		return nil, 0, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Repository(%s) find failed: %v", repositoryID, err))
	}
	if repositoryObj.NamespaceID != namespaceObj.ID {
		return nil, 0, errcode.HTTPErrCodeNotFound.Detail("Repository's namespace ref id not equal namespace id")
	}

	tags, total, err := s.RepoTag.ListTag(ctx, repositoryObj.ID, name, artifactTypes, pagination, sort)
	if err != nil {
		return nil, 0, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("List tag from db failed: %v", err))
	}
	return tags, total, nil
}

func (s *service) GetTag(ctx context.Context, id string) (*models.Tag, error) {
	tag, err := s.RepoTag.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Tag not found: %v", err))
		}
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get tag from db failed: %v", err))
	}
	return tag, nil
}

func (s *service) GetArtifactRaw(ctx context.Context, digestStr string) ([]byte, error) {
	dgst, err := digest.Parse(digestStr)
	if err != nil {
		slog.Error("parse artifact digest failed", "err", err, "digest", digestStr)
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Parse artifact digest failed: %v", err))
	}
	reader, err := s.Storage.Reader(ctx, utils.GenManifestPathByDigest(dgst))
	if err != nil {
		slog.Error("read artifact raw failed", "err", err, "digest", digestStr)
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Read artifact raw failed: %v", err))
	}
	defer reader.Close()
	raw, err := io.ReadAll(reader)
	if err != nil {
		slog.Error("read artifact raw failed", "err", err, "digest", digestStr)
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Read artifact raw failed: %v", err))
	}
	return raw, nil
}

func (s *service) DeleteTag(ctx context.Context, namespaceID, repositoryID, id string) error {
	namespaceObj, err := s.RepoNs.Get(ctx, namespaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Namespace(%s) not found: %v", namespaceID, err))
		}
		return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Namespace(%s) find failed: %v", namespaceID, err))
	}

	repositoryObj, err := s.RepoRegistry.Get(ctx, repositoryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Repository(%s) not found: %v", repositoryID, err))
		}
		return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Repository(%s) find failed: %v", repositoryID, err))
	}
	if repositoryObj.NamespaceID != namespaceObj.ID {
		return errcode.HTTPErrCodeNotFound.Detail("Repository's namespace ref id not equal namespace id")
	}

	err = s.RepoTag.DeleteByID(ctx, id)
	if err != nil {
		return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Delete tag failed: %v", err))
	}
	return nil
}
