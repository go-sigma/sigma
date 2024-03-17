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
	"errors"
	"fmt"

	"github.com/jinzhu/copier"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/modules/workq/definition"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/imagerefs"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

//go:generate mockgen -destination=mocks/repository.go -package=mocks github.com/go-sigma/sigma/pkg/dal/dao RepositoryService
//go:generate mockgen -destination=mocks/repository_factory.go -package=mocks github.com/go-sigma/sigma/pkg/dal/dao RepositoryServiceFactory

// RepositoryService is the interface that provides the repository service methods.
type RepositoryService interface {
	// Create saves the repository.
	Create(ctx context.Context, repositoryObj *models.Repository, autoCreateNamespace AutoCreateNamespace) error
	// FindAll ...
	FindAll(ctx context.Context, namespaceID, limit, last int64) ([]*models.Repository, error)
	// Get gets the repository with the specified repository ID.
	Get(ctx context.Context, repositoryID int64) (*models.Repository, error)
	// GetByName gets the repository with the specified repository name.
	GetByName(context.Context, string) (*models.Repository, error)
	// ListByDtPagination lists the repositories by the pagination.
	ListByDtPagination(ctx context.Context, limit int, lastID ...int64) ([]*models.Repository, error)
	// ListWithScrollable list the repository with scrollable last id
	ListWithScrollable(ctx context.Context, namespaceID, userID int64, name *string, limit int, lastID int64) ([]*models.Repository, error)
	// ListRepository lists all repositories.
	ListRepository(ctx context.Context, namespaceID int64, name *string, pagination types.Pagination, sort types.Sortable) ([]*models.Repository, int64, error)
	// ListRepository lists all repositories with auth.
	ListRepositoryWithAuth(ctx context.Context, namespaceID, userID int64, name *string, pagination types.Pagination, sort types.Sortable) ([]*models.Repository, int64, error)
	// CountRepository counts all repositories.
	CountRepository(ctx context.Context, namespaceID int64, name *string) (int64, error)
	// UpdateRepository update specific repository
	UpdateRepository(ctx context.Context, id int64, updates map[string]any) error
	// CountByNamespace counts the repositories by the namespace ID.
	CountByNamespace(ctx context.Context, namespaceIDs []int64) (map[int64]int64, error)
	// DeleteByID deletes the repository with the specified repository ID.
	DeleteByID(ctx context.Context, id int64) error
	// DeleteEmpty delete all of empty repository
	DeleteEmpty(ctx context.Context, namespaceID *int64) ([]string, error)
}

type repositoryService struct {
	tx *query.Query
}

// RepositoryServiceFactory is the interface that provides the repository service factory methods.
type RepositoryServiceFactory interface {
	New(txs ...*query.Query) RepositoryService
}

type repositoryServiceFactory struct{}

// NewRepositoryServiceFactory creates a new repository service factory.
func NewRepositoryServiceFactory() RepositoryServiceFactory {
	return &repositoryServiceFactory{}
}

// New creates a new repository service.
func (s *repositoryServiceFactory) New(txs ...*query.Query) RepositoryService {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &repositoryService{
		tx: tx,
	}
}

// AutoCreateNamespace ...
type AutoCreateNamespace struct {
	AutoCreate     bool
	Visibility     enums.Visibility
	UserID         int64
	ProducerClient definition.WorkQueueProducer
}

// Create creates a new repository.
func (s *repositoryService) Create(ctx context.Context, repositoryObj *models.Repository, autoCreateNamespace AutoCreateNamespace) error {
	_, ns, _, _, err := imagerefs.Parse(repositoryObj.Name)
	if err != nil {
		return err
	}
	namespaceObj, err := s.tx.Namespace.WithContext(ctx).Where(s.tx.Namespace.Name.Eq(ns)).First()
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if !autoCreateNamespace.AutoCreate {
			return fmt.Errorf("namespace %s not found", ns)
		}
		namespaceObj = &models.Namespace{
			Name:       ns,
			Visibility: autoCreateNamespace.Visibility,
		}
		if !namespaceObj.Visibility.IsValid() {
			namespaceObj.Visibility = enums.VisibilityPrivate
		}
		err = s.tx.Namespace.WithContext(ctx).Create(namespaceObj)
		if err != nil {
			return err
		}
		err = s.tx.CasbinRule.WithContext(ctx).Create(&models.CasbinRule{
			PType: ptr.Of("g"),
			V0:    ptr.Of(fmt.Sprintf("%d", autoCreateNamespace.UserID)),
			V1:    ptr.Of(enums.NamespaceRoleAdmin.String()),
			V2:    ptr.Of(namespaceObj.Name),
			V3:    ptr.Of(""),
			V4:    ptr.Of(""),
			V5:    ptr.Of(""),
		})
		if err != nil {
			return err
		}
		namespaceMember := &models.NamespaceMember{UserID: autoCreateNamespace.UserID, NamespaceID: namespaceObj.ID, Role: enums.NamespaceRoleAdmin}
		err = s.tx.NamespaceMember.WithContext(ctx).Create(namespaceMember)
		if err != nil {
			return err
		}

		err = s.tx.Audit.WithContext(ctx).Create(&models.Audit{
			UserID:       autoCreateNamespace.UserID,
			NamespaceID:  ptr.Of(namespaceObj.ID),
			Action:       enums.AuditActionCreate,
			ResourceType: enums.AuditResourceTypeNamespace,
			Resource:     namespaceObj.Name,
		})
		if err != nil {
			return err
		}

		if autoCreateNamespace.ProducerClient != nil {
			err = autoCreateNamespace.ProducerClient.Produce(ctx, enums.DaemonWebhook.String(), types.DaemonWebhookPayload{
				NamespaceID:  ptr.Of(namespaceObj.ID),
				Action:       enums.WebhookActionCreate,
				ResourceType: enums.WebhookResourceTypeNamespace,
				Payload:      utils.MustMarshal(namespaceObj),
			}, definition.ProducerOption{Tx: s.tx})
			if err != nil {
				log.Error().Err(err).Msg("Webhook event produce failed")
				return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Webhook event produce failed: %v", err))
			}
		}
	}
	repositoryObj.NamespaceID = namespaceObj.ID

	findRepositoryObj, err := s.GetByName(ctx, repositoryObj.Name)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		err = s.tx.Repository.WithContext(ctx).Create(repositoryObj)
		if err != nil {
			return err
		}
		err = s.tx.Audit.WithContext(ctx).Create(&models.Audit{
			UserID:       autoCreateNamespace.UserID,
			NamespaceID:  ptr.Of(namespaceObj.ID),
			Action:       enums.AuditActionCreate,
			ResourceType: enums.AuditResourceTypeRepository,
			Resource:     repositoryObj.Name,
			ReqRaw:       utils.MustMarshal(repositoryObj),
		})
		if err != nil {
			return err
		}
		if autoCreateNamespace.ProducerClient != nil {
			err = autoCreateNamespace.ProducerClient.Produce(ctx, enums.DaemonWebhook.String(), types.DaemonWebhookPayload{
				NamespaceID:  ptr.Of(namespaceObj.ID),
				Action:       enums.WebhookActionCreate,
				ResourceType: enums.WebhookResourceTypeRepository,
				Payload:      utils.MustMarshal(repositoryObj),
			}, definition.ProducerOption{Tx: s.tx})
			if err != nil {
				log.Error().Err(err).Msg("Webhook event produce failed")
				return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Webhook event produce failed: %v", err))
			}
		}
		return nil
	}
	return copier.Copy(repositoryObj, findRepositoryObj)
}

// FindAll ...
func (s *repositoryService) FindAll(ctx context.Context, namespaceID, limit, last int64) ([]*models.Repository, error) {
	return s.tx.Repository.WithContext(ctx).
		Where(s.tx.Repository.ID.Gt(last), s.tx.Repository.NamespaceID.Eq(namespaceID)).
		Limit(int(limit)).Order(s.tx.Repository.ID).Find()
}

// Get gets the repository with the specified repository ID.
func (s *repositoryService) Get(ctx context.Context, repositoryID int64) (*models.Repository, error) {
	return s.tx.Repository.WithContext(ctx).
		Where(s.tx.Repository.ID.Eq(repositoryID)).
		Preload(s.tx.Repository.Builder.CodeRepository).
		Preload(s.tx.Repository.Builder.CodeRepository.User3rdParty).
		Preload(s.tx.Repository.Builder).First()
}

// GetByName gets the repository with the specified repository name.
func (s *repositoryService) GetByName(ctx context.Context, name string) (*models.Repository, error) {
	return s.tx.Repository.WithContext(ctx).Where(s.tx.Repository.Name.Eq(name)).First()
}

// ListByDtPagination lists the repositories by the pagination.
func (s *repositoryService) ListByDtPagination(ctx context.Context, limit int, lastID ...int64) ([]*models.Repository, error) {
	do := s.tx.Repository.WithContext(ctx)
	if len(lastID) > 0 {
		do = do.Where(s.tx.Repository.ID.Gt(lastID[0]))
	}
	return do.Order(s.tx.Repository.ID).Limit(limit).Find()
}

// ListWithScrollable list the repository with scrollable last id
func (s *repositoryService) ListWithScrollable(ctx context.Context, namespaceID, userID int64, name *string, limit int, lastID int64) ([]*models.Repository, error) {
	q := s.tx.Repository.WithContext(ctx)
	if namespaceID != 0 {
		q = q.Where(s.tx.Repository.NamespaceID.Eq(namespaceID))
	}
	userObj, err := s.tx.User.WithContext(ctx).Where(s.tx.User.ID.Eq(userID)).First()
	if err != nil {
		return nil, err
	}
	if !(userObj.Role == enums.UserRoleAdmin || userObj.Role == enums.UserRoleRoot) {
		q = q.LeftJoin(s.tx.NamespaceMember, s.tx.Repository.NamespaceID.EqCol(s.tx.NamespaceMember.NamespaceID), s.tx.NamespaceMember.UserID.Eq(userID)).
			Where(s.tx.NamespaceMember.ID.IsNotNull())
	}
	if name != nil {
		q = q.Where(s.tx.Repository.Name.Like(fmt.Sprintf("%s%%", ptr.To(name))))
	}
	if lastID > 0 {
		q = q.Where(s.tx.Repository.ID.Gt(lastID))
	}
	return q.Order(s.tx.Repository.ID).Limit(limit).Find()
}

// ListRepository lists all repositories with auth.
func (s *repositoryService) ListRepositoryWithAuth(ctx context.Context, namespaceID, userID int64, name *string, pagination types.Pagination, sort types.Sortable) ([]*models.Repository, int64, error) {
	pagination = utils.NormalizePagination(pagination)
	q := s.tx.Repository.WithContext(ctx)
	if namespaceID != 0 {
		q = q.Where(s.tx.Repository.NamespaceID.Eq(namespaceID))
	}
	if name != nil {
		q = q.Where(s.tx.Repository.Name.Like(fmt.Sprintf("%s%%", ptr.To(name))))
	}
	userObj, err := s.tx.User.WithContext(ctx).Where(s.tx.User.ID.Eq(userID)).First()
	if err != nil {
		return nil, 0, err
	}
	if !(userObj.Role == enums.UserRoleAdmin || userObj.Role == enums.UserRoleRoot) {
		q = q.LeftJoin(s.tx.NamespaceMember, s.tx.Repository.NamespaceID.EqCol(s.tx.NamespaceMember.NamespaceID), s.tx.NamespaceMember.UserID.Eq(userID)).
			Where(s.tx.NamespaceMember.ID.IsNotNull())
	}
	field, ok := s.tx.Repository.GetFieldByName(ptr.To(sort.Sort))
	if ok {
		switch ptr.To(sort.Method) {
		case enums.SortMethodDesc:
			q = q.Order(field.Desc())
		case enums.SortMethodAsc:
			q = q.Order(field)
		default:
			q = q.Order(s.tx.Repository.UpdatedAt.Desc())
		}
	} else {
		q = q.Order(s.tx.Repository.UpdatedAt.Desc())
	}
	q = q.Preload(s.tx.Repository.Builder.CodeRepository).
		Preload(s.tx.Repository.Builder.CodeRepository.User3rdParty)
	return q.FindByPage(ptr.To(pagination.Limit)*(ptr.To(pagination.Page)-1), ptr.To(pagination.Limit))
}

// ListRepository lists all repositories.
func (s *repositoryService) ListRepository(ctx context.Context, namespaceID int64, name *string, pagination types.Pagination, sort types.Sortable) ([]*models.Repository, int64, error) {
	pagination = utils.NormalizePagination(pagination)
	q := s.tx.Repository.WithContext(ctx).Where(s.tx.Repository.NamespaceID.Eq(namespaceID))
	if name != nil {
		q = q.Where(s.tx.Repository.Name.Like(fmt.Sprintf("%s%%", ptr.To(name))))
	}
	field, ok := s.tx.Repository.GetFieldByName(ptr.To(sort.Sort))
	if ok {
		switch ptr.To(sort.Method) {
		case enums.SortMethodDesc:
			q = q.Order(field.Desc())
		case enums.SortMethodAsc:
			q = q.Order(field)
		default:
			q = q.Order(s.tx.Repository.UpdatedAt.Desc())
		}
	} else {
		q = q.Order(s.tx.Repository.UpdatedAt.Desc())
	}
	return q.FindByPage(ptr.To(pagination.Limit)*(ptr.To(pagination.Page)-1), ptr.To(pagination.Limit))
}

// UpdateRepository ...
func (s *repositoryService) UpdateRepository(ctx context.Context, id int64, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	_, err := s.tx.Repository.WithContext(ctx).Where(s.tx.Repository.ID.Eq(id)).UpdateColumns(updates)
	if err != nil {
		return err
	}
	return nil
}

// CountRepository counts all repositories.
func (s *repositoryService) CountRepository(ctx context.Context, namespaceID int64, name *string) (int64, error) {
	q := s.tx.Repository.WithContext(ctx).Where(s.tx.Repository.NamespaceID.Eq(namespaceID))
	if name != nil {
		q = q.Where(s.tx.Repository.Name.Like(fmt.Sprintf("%s%%", ptr.To(name))))
	}
	return q.Count()
}

// DeleteByID deletes the repository with the specified repository ID.
func (s *repositoryService) DeleteByID(ctx context.Context, id int64) error {
	_, err := s.tx.Repository.WithContext(ctx).Where(s.tx.Repository.ID.Eq(id)).Delete()
	if err != nil {
		return err
	}
	return nil
}

// CountByNamespace counts the repositories by the namespace IDs.
func (s *repositoryService) CountByNamespace(ctx context.Context, namespaceIDs []int64) (map[int64]int64, error) {
	tagCount := make(map[int64]int64)
	var count []struct {
		NamespaceID int64 `gorm:"column:namespace_id"`
		Count       int64 `gorm:"column:count"`
	}
	err := s.tx.Repository.WithContext(ctx).
		Where(s.tx.Repository.NamespaceID.In(namespaceIDs...)).
		Group(s.tx.Repository.NamespaceID).
		Select(s.tx.Repository.NamespaceID, s.tx.Repository.ID.Count().As("count")).
		Scan(&count)
	if err != nil {
		return nil, err
	}
	for _, c := range count {
		tagCount[c.NamespaceID] = c.Count
	}
	return tagCount, nil
}

// DeleteEmpty delete all of empty repository
func (s *repositoryService) DeleteEmpty(ctx context.Context, namespaceID *int64) ([]string, error) {
	q := s.tx.Repository.WithContext(ctx).
		LeftJoin(s.tx.Artifact, s.tx.Repository.ID.EqCol(s.tx.Artifact.RepositoryID)).
		LeftJoin(s.tx.Tag, s.tx.Repository.ID.EqCol(s.tx.Tag.RepositoryID)).
		Where(s.tx.Artifact.RepositoryID.IsNull(), s.tx.Tag.RepositoryID.IsNull())
	if namespaceID != nil {
		q = q.Where(s.tx.Repository.NamespaceID.Eq(ptr.To(namespaceID)))
	}
	repositoryObjs, err := q.Find()
	if err != nil {
		return nil, err
	}
	IDs := make([]int64, 0, len(repositoryObjs))
	result := make([]string, 0, len(repositoryObjs))
	for _, r := range repositoryObjs {
		result = append(result, r.Name)
		IDs = append(IDs, r.ID)
	}
	_, err = s.tx.Repository.WithContext(ctx).Where(s.tx.Repository.ID.In(IDs...)).Delete()
	if err != nil {
		return nil, err
	}
	return result, nil
}
