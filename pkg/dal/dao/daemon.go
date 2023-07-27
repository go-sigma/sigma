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

	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

//go:generate mockgen -destination=mocks/daemon.go -package=mocks github.com/go-sigma/sigma/pkg/dal/dao DaemonService
//go:generate mockgen -destination=mocks/daemon_factory.go -package=mocks github.com/go-sigma/sigma/pkg/dal/dao DaemonServiceFactory

// DaemonService is the interface that provides methods to operate on daemon model
type DaemonService interface {
	// Create creates a new daemon log record in the database
	Create(ctx context.Context, daemonLog *models.DaemonLog) error
	// CreateMany creates many new daemon log records in the database
	CreateMany(ctx context.Context, daemonLogs []*models.DaemonLog) error
	// Delete delete a daemon log record with specific id
	Delete(ctx context.Context, id int64) error
	// List lists all daemon log
	List(ctx context.Context, pagination types.Pagination, sort types.Sortable) ([]*models.DaemonLog, int64, error)
}

type daemonService struct {
	tx *query.Query
}

// DaemonServiceFactory is the interface that provides the daemon service factory methods.
type DaemonServiceFactory interface {
	New(txs ...*query.Query) DaemonService
}

type daemonServiceFactory struct{}

// NewDaemonServiceFactory creates a new audit service factory.
func NewDaemonServiceFactory() DaemonServiceFactory {
	return &daemonServiceFactory{}
}

func (f *daemonServiceFactory) New(txs ...*query.Query) DaemonService {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &daemonService{
		tx: tx,
	}
}

// Create creates a new daemon record in the database
func (s *daemonService) Create(ctx context.Context, daemonLog *models.DaemonLog) error {
	return s.tx.DaemonLog.WithContext(ctx).Create(daemonLog)
}

// CreateMany creates many new daemon log records in the database
func (s *daemonService) CreateMany(ctx context.Context, daemonLogs []*models.DaemonLog) error {
	return s.tx.DaemonLog.WithContext(ctx).Create(daemonLogs...)
}

// Delete delete a daemon log record with specific id
func (s *daemonService) Delete(ctx context.Context, id int64) error {
	matched, err := s.tx.DaemonLog.WithContext(ctx).Where(s.tx.DaemonLog.ID.Eq(id)).Delete()
	if err != nil {
		return err
	}
	if matched.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// List lists all daemon log
func (s *daemonService) List(ctx context.Context, pagination types.Pagination, sort types.Sortable) ([]*models.DaemonLog, int64, error) {
	pagination = utils.NormalizePagination(pagination)
	query := s.tx.DaemonLog.WithContext(ctx)
	field, ok := s.tx.DaemonLog.GetFieldByName(ptr.To(sort.Sort))
	if ok {
		switch ptr.To(sort.Method) {
		case enums.SortMethodDesc:
			query = query.Order(field.Desc())
		case enums.SortMethodAsc:
			query = query.Order(field)
		default:
			query = query.Order(s.tx.DaemonLog.UpdatedAt.Desc())
		}
	} else {
		query = query.Order(s.tx.DaemonLog.UpdatedAt.Desc())
	}
	return query.FindByPage(ptr.To(pagination.Limit)*(ptr.To(pagination.Page)-1), ptr.To(pagination.Limit))
}
