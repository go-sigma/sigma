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

package repositories

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"go.uber.org/dig"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	repobuilder "github.com/go-sigma/sigma/pkg/dal/repository/builder"
	reponamespace "github.com/go-sigma/sigma/pkg/dal/repository/namespace"
	reporegistry "github.com/go-sigma/sigma/pkg/dal/repository/registry"
	"github.com/go-sigma/sigma/pkg/distribution/reference"
	"github.com/go-sigma/sigma/pkg/infra/workq"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

//go:generate mockgen -mock_names Service=MockRepositoryService -destination=repositories_mocks.go -package=repositories github.com/go-sigma/sigma/pkg/service/repositories Service

// Service encapsulates repository-related business logic.
type Service interface {
	// CreateRepository creates a repository and produces its webhook event.
	CreateRepository(ctx context.Context, userID string, req api.CreateRepositoryRequest) (*models.Repository, error)
	// ListRepositories lists repositories with auth filtering, also returns builders keyed by repository id.
	ListRepositories(ctx context.Context, userID string, namespaceID string, name *string, pagination api.Pagination, sort api.Sortable) ([]*models.Repository, map[string]*models.Builder, int64, error)
	// GetRepository gets a repository by ID.
	GetRepository(ctx context.Context, id string) (*models.Repository, error)
	// GetRepositoryByName gets a repository by name (used by distribution base handler).
	GetRepositoryByName(ctx context.Context, name string) (*models.Repository, error)
	// UpdateRepository updates a repository (includes: validate namespace match + update columns).
	UpdateRepository(ctx context.Context, userID string, req api.UpdateRepositoryRequest) error
	// DeleteRepository deletes a repository (includes: validate namespace match + delete in transaction).
	DeleteRepository(ctx context.Context, userID, namespaceID, id string) error
}

type service struct {
	dig.In

	Config       *config.Configuration
	RepoNs       reponamespace.NamespaceRepository
	RepoRegistry reporegistry.RepositoryRepository
	RepoBuilder  repobuilder.BuilderRepository
	Producer     workq.Producer
}

func NewService(digCon *dig.Container) error {
	return digCon.Provide(func(params service) Service {
		return &params
	})
}

func (s *service) CreateRepository(ctx context.Context, userID string, req api.CreateRepositoryRequest) (*models.Repository, error) {
	namespaceObj, err := s.RepoNs.Get(ctx, req.NamespaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Namespace(%s) not found: %v", req.NamespaceID, err))
		}
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Namespace(%s) find failed: %v", req.NamespaceID, err))
	}

	_, namespaceName, _, _, err := reference.Parse(req.Name)
	if err != nil {
		return nil, errcode.HTTPErrCodeBadRequest.Detail(fmt.Sprintf("Repository name is invalid: %v", err))
	}
	if namespaceName != namespaceObj.Name {
		return nil, errcode.HTTPErrCodeBadRequest.Detail("Repository namespace does not match the request namespace")
	}

	existingRepositoryObj, err := s.RepoRegistry.GetByName(ctx, req.Name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get repository by name failed: %v", err))
	}
	if err == nil {
		return existingRepositoryObj, nil
	}

	repositoryObj := &models.Repository{
		ID:          uuid.NewV7String(),
		NamespaceID: namespaceObj.ID,
		Name:        req.Name,
		Description: req.Description,
		Overview:    []byte(ptr.To(req.Overview)),
		TagLimit:    ptr.To(req.TagLimit),
		SizeLimit:   ptr.To(req.SizeLimit),
	}
	err = query.Q.Transaction(func(tx *query.Query) error {
		repoRegistry := reporegistry.NewRepositoryRepository(tx)
		err = repoRegistry.Create(ctx, repositoryObj)
		if err != nil {
			slog.Error("repository create failed", "err", err, "repositoryObj", repositoryObj)
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create repository failed: %v", err))
		}
		err = s.produceRepositoryCreateWebhook(ctx, namespaceObj.ID, repositoryObj)
		if err != nil {
			slog.Error("webhook event produce failed", "err", err)
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Webhook event produce failed: %v", err))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.seedRepositoryCache(ctx, repositoryObj)
	return repositoryObj, nil
}

func (s *service) ListRepositories(ctx context.Context, userID string, namespaceID string, name *string, pagination api.Pagination, sort api.Sortable) ([]*models.Repository, map[string]*models.Builder, int64, error) {
	if namespaceID != "" {
		_, err := s.RepoNs.Get(ctx, namespaceID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, nil, 0, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Namespace(%s) not found: %v", namespaceID, err))
			}
			return nil, nil, 0, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Namespace(%s) find failed: %v", namespaceID, err))
		}
	}

	repositoryObjs, total, err := s.RepoRegistry.ListRepositoryWithAuth(ctx, namespaceID, userID, name, pagination, sort)
	if err != nil {
		return nil, nil, 0, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("List repository failed: %v", err))
	}

	repositoryIDs := make([]string, 0, len(repositoryObjs))
	for _, r := range repositoryObjs {
		repositoryIDs = append(repositoryIDs, r.ID)
	}
	builderMap, err := s.RepoBuilder.GetByRepositoryIDs(ctx, repositoryIDs)
	if err != nil {
		return nil, nil, 0, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Find builders with repository failed: %v", err))
	}
	return repositoryObjs, builderMap, total, nil
}

func (s *service) GetRepository(ctx context.Context, id string) (*models.Repository, error) {
	repositoryObj, err := s.RepoRegistry.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get repository by id not found: %v", err))
		}
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get repository by id failed: %v", err))
	}
	return repositoryObj, nil
}

func (s *service) GetRepositoryByName(ctx context.Context, name string) (*models.Repository, error) {
	repositoryObj, err := s.RepoRegistry.GetByName(ctx, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get repository by name not found: %v", err))
		}
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get repository by name failed: %v", err))
	}
	return repositoryObj, nil
}

func (s *service) UpdateRepository(ctx context.Context, userID string, req api.UpdateRepositoryRequest) error {
	namespaceObj, err := s.RepoNs.Get(ctx, req.NamespaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.HTTPErrCodeNotFound.Detail(err.Error())
		}
		return errcode.HTTPErrCodeInternalError.Detail(err.Error())
	}

	repositoryObj, err := s.RepoRegistry.Get(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.HTTPErrCodeNotFound.Detail(err.Error())
		}
		return errcode.HTTPErrCodeInternalError.Detail(err.Error())
	}
	if repositoryObj.NamespaceID != namespaceObj.ID {
		return errcode.HTTPErrCodeNotFound.Detail("Repository's namespace ref id not equal namespace id")
	}

	updates := make(map[string]any, 5)
	if req.SizeLimit != nil {
		updates[query.Repository.SizeLimit.ColumnName().String()] = ptr.To(req.SizeLimit)
	}
	if req.TagLimit != nil {
		updates[query.Repository.TagLimit.ColumnName().String()] = ptr.To(req.TagLimit)
	}
	if req.Description != nil {
		updates[query.Repository.Description.ColumnName().String()] = ptr.To(req.Description)
	}
	if req.Overview != nil {
		updates[query.Repository.Overview.ColumnName().String()] = []byte(ptr.To(req.Overview))
	}

	if len(updates) > 0 {
		err = s.RepoRegistry.UpdateRepository(ctx, repositoryObj.ID, updates)
		if err != nil {
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Repository update failed: %v", err))
		}
	}
	return nil
}

func (s *service) DeleteRepository(ctx context.Context, userID, namespaceID, id string) error {
	namespaceObj, err := s.RepoNs.Get(ctx, namespaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Namespace(%s) not found: %v", namespaceID, err))
		}
		return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Namespace(%s) find failed: %v", namespaceID, err))
	}

	repositoryObj, err := s.RepoRegistry.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.HTTPErrCodeNotFound.Detail(err.Error())
		}
		return errcode.HTTPErrCodeInternalError.Detail(err.Error())
	}
	if repositoryObj.NamespaceID != namespaceObj.ID {
		return errcode.HTTPErrCodeNotFound.Detail("Repository's namespace ref id not equal namespace id")
	}

	err = query.Q.Transaction(func(tx *query.Query) error {
		repoRegistry := reporegistry.NewRepositoryRepository(tx)
		err = repoRegistry.DeleteByID(ctx, id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Delete repository by id not found: %v", err))
			}
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Delete repository by id failed: %v", err))
		}
		return nil
	})
	if err != nil {
		return err
	}
	s.invalidateRepositoryCache(ctx, repositoryObj)
	return nil
}

func (s *service) invalidateRepositoryCache(ctx context.Context, repositoryObj *models.Repository) {
	invalidator, ok := s.RepoRegistry.(reporegistry.RepositoryCacheInvalidator)
	if !ok {
		return
	}
	if err := invalidator.InvalidateRepository(ctx, repositoryObj); err != nil {
		slog.Warn("invalidate repository cache failed", "err", err, "repository_id", repositoryObj.ID)
	}
}

func (s *service) seedRepositoryCache(ctx context.Context, repositoryObj *models.Repository) {
	seeder, ok := s.RepoRegistry.(reporegistry.RepositoryCacheSeeder)
	if !ok {
		return
	}
	if err := seeder.SeedRepository(ctx, repositoryObj); err != nil {
		slog.Warn("seed repository cache failed", "err", err, "repository_id", repositoryObj.ID)
	}
}

func (s *service) produceRepositoryCreateWebhook(ctx context.Context, namespaceID string, repositoryObj *models.Repository) error {
	if s.Producer == nil {
		return nil
	}
	return s.Producer.Produce(ctx, enums.DaemonWebhook, api.DaemonWebhookPayload{
		NamespaceID:  &namespaceID,
		Action:       enums.WebhookActionCreate,
		ResourceType: enums.WebhookResourceTypeRepository,
		Payload:      utils.MustMarshal(repositoryObj),
	})
}
