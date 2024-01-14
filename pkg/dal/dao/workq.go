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
	"github.com/go-sigma/sigma/pkg/types/enums"
)

//go:generate mockgen -destination=mocks/workq.go -package=mocks github.com/go-sigma/sigma/pkg/dal/dao WorkQueueService
//go:generate mockgen -destination=mocks/workq_factory.go -package=mocks github.com/go-sigma/sigma/pkg/dal/dao WorkQueueServiceFactory

// WorkQueueService is the interface that provides methods to operate on work queue model
type WorkQueueService interface {
	// Create creates a new work queue record in the database
	Create(ctx context.Context, workqObj *models.WorkQueue) error
	// Get get a work queue record
	Get(ctx context.Context, topic string) (*models.WorkQueue, error)
	// UpdateStatus update a work queue record status
	UpdateStatus(ctx context.Context, id int64, version, newVersion string, times int, status enums.TaskCommonStatus) error
}

type workQueueService struct {
	tx *query.Query
}

// WorkQueueServiceFactory is the interface that provides the work queue service factory methods.
type WorkQueueServiceFactory interface {
	New(txs ...*query.Query) WorkQueueService
}

type workQueueServiceFactory struct{}

// NewWorkQueueServiceFactory creates a new work queue service factory.
func NewWorkQueueServiceFactory() WorkQueueServiceFactory {
	return &workQueueServiceFactory{}
}

func (s *workQueueServiceFactory) New(txs ...*query.Query) WorkQueueService {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &workQueueService{
		tx: tx,
	}
}

// Create creates a new work queue record in the database
func (s workQueueService) Create(ctx context.Context, workqObj *models.WorkQueue) error {
	return s.tx.WorkQueue.WithContext(ctx).Create(workqObj)
}

// Get get a work queue record
func (s workQueueService) Get(ctx context.Context, topic string) (*models.WorkQueue, error) {
	return s.tx.WorkQueue.WithContext(ctx).Where(
		s.tx.WorkQueue.Status.Eq(enums.TaskCommonStatusPending),
		s.tx.WorkQueue.Topic.Eq(topic),
	).Order(s.tx.WorkQueue.UpdatedAt).First()
}

// UpdateStatus update a work queue record
func (s workQueueService) UpdateStatus(ctx context.Context, id int64, version, newVersion string, times int, status enums.TaskCommonStatus) error {
	value := map[string]any{
		query.WorkQueue.Status.ColumnName().String():  status,
		query.WorkQueue.Version.ColumnName().String(): newVersion,
		query.WorkQueue.Times.ColumnName().String():   times,
	}
	result, err := s.tx.WorkQueue.WithContext(ctx).Where(
		s.tx.WorkQueue.ID.Eq(id),
		s.tx.WorkQueue.Version.Eq(version),
	).UpdateColumns(value)
	if err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
