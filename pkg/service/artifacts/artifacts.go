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

package artifacts

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/dig"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/dal/models"
	reporegistry "github.com/go-sigma/sigma/pkg/dal/repository/registry"
	"github.com/go-sigma/sigma/pkg/server/errcode"
)

//go:generate go tool mockgen -mock_names Service=MockArtifactService -destination=artifacts_mocks.go -package=artifacts github.com/go-sigma/sigma/pkg/service/artifacts Service

// Service encapsulates artifact-related business logic.
type Service interface {
	// ListArtifacts lists artifacts and returns the total count.
	ListArtifacts(ctx context.Context, req api.ListArtifactRequest) ([]*models.Artifact, int64, error)
	// GetArtifact gets an artifact by repository name and digest.
	GetArtifact(ctx context.Context, repository string, digest string) (*models.Artifact, error)
	// DeleteArtifact deletes an artifact by repository name and digest (includes: cascade delete blobs and tags).
	DeleteArtifact(ctx context.Context, repository string, digest string) error
}

type service struct {
	dig.In

	RepoRegistry reporegistry.RepositoryRepository
	RepoArtifact reporegistry.ArtifactRepository
}

func NewService(params service) Service {
	return &params
}

func (s *service) ListArtifacts(ctx context.Context, req api.ListArtifactRequest) ([]*models.Artifact, int64, error) {
	artifactObjs, err := s.RepoArtifact.ListArtifact(ctx, req)
	if err != nil {
		return nil, 0, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("List artifact from db failed: %v", err))
	}
	total, err := s.RepoArtifact.CountArtifact(ctx, req)
	if err != nil {
		return nil, 0, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Count artifact from db failed: %v", err))
	}
	return artifactObjs, total, nil
}

func (s *service) GetArtifact(ctx context.Context, repositoryName string, digest string) (*models.Artifact, error) {
	repositoryObj, err := s.RepoRegistry.GetByName(ctx, repositoryName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Cannot find repository: %v", err))
		}
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get repository failed: %v", err))
	}

	artifactObj, err := s.RepoArtifact.GetByDigest(ctx, repositoryObj.ID, digest)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Artifact not found: %v", err))
		}
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get artifact failed: %v", err))
	}
	return artifactObj, nil
}

func (s *service) DeleteArtifact(ctx context.Context, repositoryName string, digest string) error {
	err := s.RepoArtifact.DeleteByDigest(ctx, repositoryName, digest)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Artifact not found: %v", err))
		}
		return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Delete artifact failed: %v", err))
	}
	return nil
}
