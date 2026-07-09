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

//go:generate mockgen -destination=tags_mocks.go -package=tags github.com/go-sigma/sigma/pkg/service/tags TagService

// TagService encapsulates tag-related business logic.
type TagService interface {
	// ListTags lists tags for a repository (includes: validate namespace/repository match).
	ListTags(ctx context.Context, namespaceID, repositoryID string, name *string, artifactTypes []enums.ArtifactType, pagination api.Pagination, sort api.Sortable) ([]*models.Tag, int64, error)
	// GetTag gets a tag by ID.
	GetTag(ctx context.Context, id string) (*models.Tag, error)
	// GetArtifactRaw gets the manifest raw content for an artifact digest.
	GetArtifactRaw(ctx context.Context, digestStr string) ([]byte, error)
	// DeleteTag deletes a tag by ID (includes: validate namespace/repository match).
	DeleteTag(ctx context.Context, namespaceID, repositoryID, id string) error
}

type tagService struct {
	namespaceRepository  reponamespace.NamespaceRepository
	repositoryRepository reporegistry.RepositoryRepository
	tagRepository        reporegistry.TagRepository
	storageDriver        storage.StorageDriver
}

type ServiceParams struct {
	dig.In

	NamespaceRepository  reponamespace.NamespaceRepository
	RepositoryRepository reporegistry.RepositoryRepository
	TagRepository        reporegistry.TagRepository
	StorageDriver        storage.StorageDriver
}

func NewService(digCon *dig.Container) error {
	return digCon.Provide(func(params ServiceParams) TagService {
		return &tagService{
			namespaceRepository:  params.NamespaceRepository,
			repositoryRepository: params.RepositoryRepository,
			tagRepository:        params.TagRepository,
			storageDriver:        params.StorageDriver,
		}
	})
}

func (s *tagService) ListTags(ctx context.Context, namespaceID, repositoryID string, name *string, artifactTypes []enums.ArtifactType, pagination api.Pagination, sort api.Sortable) ([]*models.Tag, int64, error) {
	namespaceRepository := s.namespaceRepository
	namespaceObj, err := namespaceRepository.Get(ctx, namespaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Namespace(%s) not found: %v", namespaceID, err))
		}
		return nil, 0, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Namespace(%s) find failed: %v", namespaceID, err))
	}

	repositoryRepository := s.repositoryRepository
	repositoryObj, err := repositoryRepository.Get(ctx, repositoryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Repository(%s) not found: %v", repositoryID, err))
		}
		return nil, 0, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Repository(%s) find failed: %v", repositoryID, err))
	}
	if repositoryObj.NamespaceID != namespaceObj.ID {
		return nil, 0, errcode.HTTPErrCodeNotFound.Detail("Repository's namespace ref id not equal namespace id")
	}

	tagRepository := s.tagRepository
	tags, total, err := tagRepository.ListTag(ctx, repositoryObj.ID, name, artifactTypes, pagination, sort)
	if err != nil {
		return nil, 0, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("List tag from db failed: %v", err))
	}
	return tags, total, nil
}

func (s *tagService) GetTag(ctx context.Context, id string) (*models.Tag, error) {
	tagRepository := s.tagRepository
	tag, err := tagRepository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Tag not found: %v", err))
		}
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get tag from db failed: %v", err))
	}
	return tag, nil
}

func (s *tagService) GetArtifactRaw(ctx context.Context, digestStr string) ([]byte, error) {
	dgst, err := digest.Parse(digestStr)
	if err != nil {
		slog.Error("parse artifact digest failed", "err", err, "digest", digestStr)
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Parse artifact digest failed: %v", err))
	}
	reader, err := s.storageDriver.Reader(ctx, utils.GenManifestPathByDigest(dgst))
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

func (s *tagService) DeleteTag(ctx context.Context, namespaceID, repositoryID, id string) error {
	namespaceRepository := s.namespaceRepository
	namespaceObj, err := namespaceRepository.Get(ctx, namespaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Namespace(%s) not found: %v", namespaceID, err))
		}
		return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Namespace(%s) find failed: %v", namespaceID, err))
	}

	repositoryRepository := s.repositoryRepository
	repositoryObj, err := repositoryRepository.Get(ctx, repositoryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Repository(%s) not found: %v", repositoryID, err))
		}
		return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Repository(%s) find failed: %v", repositoryID, err))
	}
	if repositoryObj.NamespaceID != namespaceObj.ID {
		return errcode.HTTPErrCodeNotFound.Detail("Repository's namespace ref id not equal namespace id")
	}

	tagRepository := s.tagRepository
	err = tagRepository.DeleteByID(ctx, id)
	if err != nil {
		return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Delete tag failed: %v", err))
	}
	return nil
}
