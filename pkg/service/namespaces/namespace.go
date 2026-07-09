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

package namespaces

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"go.uber.org/dig"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	reponamespace "github.com/go-sigma/sigma/pkg/dal/repository/namespace"
	reporegistry "github.com/go-sigma/sigma/pkg/dal/repository/registry"
	"github.com/go-sigma/sigma/pkg/infra/workq"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

//go:generate mockgen -destination=namespace_mocks.go -package=namespaces github.com/go-sigma/sigma/pkg/service/namespaces NamespaceService

// NamespaceService encapsulates namespace-related business logic.
type NamespaceService interface {
	// CreateNamespace creates a namespace (includes: create namespace + add admin member + produce webhook).
	CreateNamespace(ctx context.Context, userID string, req api.PostNamespaceRequest) (*models.Namespace, error)
	// ListNamespaces lists namespaces with auth filtering.
	ListNamespaces(ctx context.Context, userID string, name *string, pagination api.Pagination, sort api.Sortable) ([]*models.Namespace, int64, error)
	// GetNamespace gets a namespace by ID, also returns repository count and tag count.
	GetNamespace(ctx context.Context, id string) (*models.Namespace, int64, int64, error)
	// UpdateNamespace updates a namespace (includes: update + produce webhook).
	UpdateNamespace(ctx context.Context, userID string, id string, req api.UpdateNamespaceRequest) error
	// DeleteNamespace deletes a namespace (includes: delete + produce webhook).
	DeleteNamespace(ctx context.Context, userID string, id string) error
	// HotNamespaces returns hot namespaces for the user.
	HotNamespaces(ctx context.Context, userID string) ([]*models.Namespace, error)

	// --- Namespace Member ---

	// ListNamespaceMembers lists members of a namespace.
	ListNamespaceMembers(ctx context.Context, namespaceID string, name *string, pagination api.Pagination, sort api.Sortable) ([]*models.NamespaceMember, int64, error)
	// AddNamespaceMember adds a member to a namespace (includes: check existing role + check quota + add member + reload policy).
	AddNamespaceMember(ctx context.Context, userID string, namespaceID string, targetUserID string, role enums.NamespaceRole) (*models.NamespaceMember, error)
	// UpdateNamespaceMember updates a member's role (includes: check existing role + update + reload policy).
	UpdateNamespaceMember(ctx context.Context, userID string, namespaceID string, targetUserID string, role enums.NamespaceRole) error
	// DeleteNamespaceMember removes a member from a namespace (includes: delete + reload policy).
	DeleteNamespaceMember(ctx context.Context, userID string, namespaceID string, targetUserID string) error
	// GetNamespaceMember gets a single namespace member.
	GetNamespaceMember(ctx context.Context, namespaceID string, userID string) (*models.NamespaceMember, error)
}

type namespaceService struct {
	namespaceRepository       reponamespace.NamespaceRepository
	namespaceMemberRepository reponamespace.NamespaceMemberRepository
	repositoryRepository      reporegistry.RepositoryRepository
	tagRepository             reporegistry.TagRepository
	producer                  workq.Producer
}

type ServiceParams struct {
	dig.In

	NamespaceRepository       reponamespace.NamespaceRepository
	NamespaceMemberRepository reponamespace.NamespaceMemberRepository
	RepositoryRepository      reporegistry.RepositoryRepository
	TagRepository             reporegistry.TagRepository
	Producer                  workq.Producer
}

func NewService(digCon *dig.Container) error {
	return digCon.Provide(func(params ServiceParams) NamespaceService {
		return &namespaceService{
			namespaceRepository:       params.NamespaceRepository,
			namespaceMemberRepository: params.NamespaceMemberRepository,
			repositoryRepository:      params.RepositoryRepository,
			tagRepository:             params.TagRepository,
			producer:                  params.Producer,
		}
	})
}

func (s *namespaceService) CreateNamespace(ctx context.Context, userID string, req api.PostNamespaceRequest) (*models.Namespace, error) {
	namespaceRepository := s.namespaceRepository
	_, err := namespaceRepository.GetByName(ctx, req.Name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.Error("get namespace by name failed", "err", err)
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get namespace by name failed: %v", err))
	}
	if err == nil {
		return nil, errcode.HTTPErrCodeConflict.Detail("Namespace already exists")
	}

	namespaceObj := &models.Namespace{
		ID:          uuid.NewV7String(),
		Name:        req.Name,
		Description: req.Description,
	}
	if req.Visibility != nil {
		namespaceObj.Visibility = ptr.To(req.Visibility)
	}
	if ptr.To(req.SizeLimit) > 0 {
		namespaceObj.SizeLimit = ptr.To(req.SizeLimit)
	}
	if ptr.To(req.RepositoryLimit) > 0 {
		namespaceObj.RepositoryLimit = ptr.To(req.RepositoryLimit)
	}
	if ptr.To(req.TagLimit) > 0 {
		namespaceObj.TagLimit = ptr.To(req.TagLimit)
	}

	err = query.Q.Transaction(func(tx *query.Query) error {
		nsRepo := reponamespace.NewNamespaceRepository(tx)
		err = nsRepo.Create(ctx, namespaceObj)
		if err != nil {
			slog.Error("create namespace failed", "err", err)
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create namespace failed: %v", err))
		}
		nmRepo := reponamespace.NewNamespaceMemberRepository(tx)
		_, err = nmRepo.AddNamespaceMember(ctx, userID, ptr.To(namespaceObj), enums.NamespaceRoleAdmin)
		if err != nil {
			slog.Error("add namespace member failed", "err", err)
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Add namespace member failed: %v", err))
		}
		err = s.producer.Produce(ctx, enums.DaemonWebhook, api.DaemonWebhookPayload{
			NamespaceID:  new(namespaceObj.ID),
			Action:       enums.WebhookActionCreate,
			ResourceType: enums.WebhookResourceTypeNamespace,
			Payload:      utils.MustMarshal(namespaceObj),
		})
		if err != nil {
			slog.Error("webhook event produce failed", "err", err)
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Webhook event produce failed: %v", err))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.seedNamespaceCache(ctx, namespaceObj)
	return namespaceObj, nil
}

func (s *namespaceService) ListNamespaces(ctx context.Context, userID string, name *string, pagination api.Pagination, sort api.Sortable) ([]*models.Namespace, int64, error) {
	pagination = utils.NormalizePagination(pagination)
	namespaceRepository := s.namespaceRepository
	return namespaceRepository.ListNamespaceWithAuth(ctx, userID, name, pagination, sort)
}

func (s *namespaceService) GetNamespace(ctx context.Context, id string) (*models.Namespace, int64, int64, error) {
	namespaceRepository := s.namespaceRepository
	namespaceObj, err := namespaceRepository.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, 0, errcode.HTTPErrCodeNotFound.Detail(err.Error())
		}
		return nil, 0, 0, errcode.HTTPErrCodeInternalError.Detail(err.Error())
	}

	repositoryRepository := s.repositoryRepository
	repositoryMapCount, err := repositoryRepository.CountByNamespace(ctx, []string{namespaceObj.ID})
	if err != nil {
		return nil, 0, 0, errcode.HTTPErrCodeInternalError.Detail(err.Error())
	}

	tagRepository := s.tagRepository
	tagMapCount, err := tagRepository.CountByNamespace(ctx, []string{namespaceObj.ID})
	if err != nil {
		return nil, 0, 0, errcode.HTTPErrCodeInternalError.Detail(err.Error())
	}

	return namespaceObj, repositoryMapCount[namespaceObj.ID], tagMapCount[namespaceObj.ID], nil
}

func (s *namespaceService) UpdateNamespace(ctx context.Context, userID string, id string, req api.UpdateNamespaceRequest) error {
	namespaceRepository := s.namespaceRepository
	namespaceObj, err := namespaceRepository.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.HTTPErrCodeNotFound.Detail(err.Error())
		}
		return errcode.HTTPErrCodeInternalError.Detail(err.Error())
	}

	if req.SizeLimit != nil && namespaceObj.SizeLimit > ptr.To(req.SizeLimit) {
		return errcode.HTTPErrCodeBadRequest.Detail("Namespace quota is less than the before limit")
	}

	updates := make(map[string]any, 5)
	if req.SizeLimit != nil {
		updates[query.Namespace.SizeLimit.ColumnName().String()] = ptr.To(req.SizeLimit)
	}
	if req.RepositoryLimit != nil {
		updates[query.Namespace.RepositoryLimit.ColumnName().String()] = ptr.To(req.RepositoryLimit)
	}
	if req.TagLimit != nil {
		updates[query.Namespace.TagLimit.ColumnName().String()] = ptr.To(req.TagLimit)
	}
	if req.Description != nil {
		updates[query.Namespace.Description.ColumnName().String()] = ptr.To(req.Description)
	}
	if req.Visibility != nil {
		updates[query.Namespace.Visibility.ColumnName().String()] = ptr.To(req.Visibility)
	}
	if req.Overview != nil {
		updates[query.Namespace.Overview.ColumnName().String()] = []byte(ptr.To(req.Overview))
	}

	if len(updates) == 0 {
		return nil
	}

	err = query.Q.Transaction(func(tx *query.Query) error {
		nsRepo := reponamespace.NewNamespaceRepository(tx)
		err = nsRepo.UpdateByID(ctx, namespaceObj.ID, updates)
		if err != nil {
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Update namespace failed: %v", err))
		}
		err = s.producer.Produce(ctx, enums.DaemonWebhook, api.DaemonWebhookPayload{
			NamespaceID:  new(namespaceObj.ID),
			Action:       enums.WebhookActionUpdate,
			ResourceType: enums.WebhookResourceTypeNamespace,
			Payload:      utils.MustMarshal(req),
		})
		if err != nil {
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Webhook event produce failed: %v", err))
		}
		return nil
	})
	if err != nil {
		return err
	}
	s.invalidateNamespaceCache(ctx, namespaceObj)
	return nil
}

func (s *namespaceService) DeleteNamespace(ctx context.Context, userID string, id string) error {
	namespaceRepository := s.namespaceRepository
	namespaceObj, err := namespaceRepository.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Namespace(%s) not found", id))
		}
		return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get namespace(%s) failed", id))
	}

	err = query.Q.Transaction(func(tx *query.Query) error {
		nsRepo := reponamespace.NewNamespaceRepository(tx)
		err = nsRepo.DeleteByID(ctx, id)
		if err != nil {
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Namespace(%s) delete failed: %v", id, err))
		}
		err = s.producer.Produce(ctx, enums.DaemonWebhook, api.DaemonWebhookPayload{
			NamespaceID:  new(namespaceObj.ID),
			Action:       enums.WebhookActionDelete,
			ResourceType: enums.WebhookResourceTypeNamespace,
			Payload:      utils.MustMarshal(namespaceObj),
		})
		if err != nil {
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Webhook event produce failed: %v", err))
		}
		return nil
	})
	if err != nil {
		return err
	}
	s.invalidateNamespaceCache(ctx, namespaceObj)
	return nil
}

func (s *namespaceService) HotNamespaces(ctx context.Context, userID string) ([]*models.Namespace, error) {
	return []*models.Namespace{}, nil
}

func (s *namespaceService) ListNamespaceMembers(ctx context.Context, namespaceID string, name *string, pagination api.Pagination, sort api.Sortable) ([]*models.NamespaceMember, int64, error) {
	namespaceMemberRepository := s.namespaceMemberRepository
	return namespaceMemberRepository.ListNamespaceMembers(ctx, namespaceID, name, pagination, sort)
}

func (s *namespaceService) AddNamespaceMember(ctx context.Context, userID string, namespaceID string, targetUserID string, role enums.NamespaceRole) (*models.NamespaceMember, error) {
	namespaceRepository := s.namespaceRepository
	namespaceObj, err := namespaceRepository.Get(ctx, namespaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.HTTPErrCodeNotFound.Detail(err.Error())
		}
		return nil, errcode.HTTPErrCodeInternalError.Detail(err.Error())
	}

	namespaceMemberRepository := s.namespaceMemberRepository
	_, err = namespaceMemberRepository.GetNamespaceMember(ctx, namespaceID, targetUserID)
	if err == nil {
		return nil, errcode.HTTPErrCodeConflict.Detail("User already have role in namespace")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get namespace member failed: %v", err))
	}

	roleCount, err := namespaceMemberRepository.CountNamespaceMember(ctx, targetUserID, namespaceID)
	if err != nil {
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Count namespace role failed: %v", err))
	}
	if roleCount >= consts.MaxNamespaceMember {
		return nil, errcode.HTTPErrCodeBadRequest.Detail("Max namespace role quota exceeds")
	}

	var namespaceMemberObj *models.NamespaceMember
	err = query.Q.Transaction(func(tx *query.Query) error {
		nmRepo := reponamespace.NewNamespaceMemberRepository(tx)
		namespaceMemberObj, err = nmRepo.AddNamespaceMember(ctx, targetUserID, ptr.To(namespaceObj), role)
		if err != nil {
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Add namespace role for user failed: %v", err))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return namespaceMemberObj, nil
}

func (s *namespaceService) UpdateNamespaceMember(ctx context.Context, userID string, namespaceID string, targetUserID string, role enums.NamespaceRole) error {
	namespaceRepository := s.namespaceRepository
	namespaceObj, err := namespaceRepository.Get(ctx, namespaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Namespace not found: %v", err))
		}
		return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Find namespace failed: %v", err))
	}

	namespaceMemberRepository := s.namespaceMemberRepository
	existingMember, err := namespaceMemberRepository.GetNamespaceMember(ctx, namespaceID, targetUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.HTTPErrCodeNotFound.Detail("User not have role in namespace")
		}
		return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get namespace member failed: %v", err))
	}

	if role == existingMember.Role {
		return nil
	}

	err = query.Q.Transaction(func(tx *query.Query) error {
		nmRepo := reponamespace.NewNamespaceMemberRepository(tx)
		err = nmRepo.UpdateNamespaceMember(ctx, targetUserID, ptr.To(namespaceObj), role)
		if err != nil {
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Update namespace role for user failed: %v", err))
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *namespaceService) DeleteNamespaceMember(ctx context.Context, userID string, namespaceID string, targetUserID string) error {
	namespaceRepository := s.namespaceRepository
	namespaceObj, err := namespaceRepository.Get(ctx, namespaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Namespace not found: %v", err))
		}
		return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Find namespace failed: %v", err))
	}

	err = query.Q.Transaction(func(tx *query.Query) error {
		nmRepo := reponamespace.NewNamespaceMemberRepository(tx)
		err = nmRepo.DeleteNamespaceMember(ctx, targetUserID, ptr.To(namespaceObj))
		if err != nil {
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Delete namespace role for user failed: %v", err))
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *namespaceService) GetNamespaceMember(ctx context.Context, namespaceID string, userID string) (*models.NamespaceMember, error) {
	namespaceMemberRepository := s.namespaceMemberRepository
	member, err := namespaceMemberRepository.GetNamespaceMember(ctx, namespaceID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get namespace role from db not found: %v", err))
		}
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get namespace role from db failed: %v", err))
	}
	return member, nil
}

func (s *namespaceService) seedNamespaceCache(ctx context.Context, namespaceObj *models.Namespace) {
	seeder, ok := s.namespaceRepository.(reponamespace.NamespaceCacheSeeder)
	if !ok {
		return
	}
	if err := seeder.SeedNamespace(ctx, namespaceObj); err != nil {
		slog.Warn("seed namespace cache failed", "err", err, "namespace_id", namespaceObj.ID)
	}
}

func (s *namespaceService) invalidateNamespaceCache(ctx context.Context, namespaceObj *models.Namespace) {
	invalidator, ok := s.namespaceRepository.(reponamespace.NamespaceCacheInvalidator)
	if !ok {
		return
	}
	if err := invalidator.InvalidateNamespace(ctx, namespaceObj); err != nil {
		slog.Warn("invalidate namespace cache failed", "err", err, "namespace_id", namespaceObj.ID)
	}
}
