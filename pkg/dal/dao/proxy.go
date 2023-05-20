// Copyright 2023 XImager
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

	"github.com/ximager/ximager/pkg/dal/models"
	"github.com/ximager/ximager/pkg/dal/query"
	"github.com/ximager/ximager/pkg/types/enums"
)

//go:generate mockgen -destination=mocks/proxy.go -package=mocks github.com/ximager/ximager/pkg/dal/dao ProxyService

// ProxyService defines the interface to access proxy task.
type ProxyService interface {
	// SaveProxyArtifact save a new artifact proxy task if conflict update.
	SaveProxyArtifact(ctx context.Context, task *models.ProxyArtifactTask) error
	// UpdateProxyArtifactStatus update the artifact proxy task status.
	UpdateProxyArtifactStatus(ctx context.Context, id uint64, status enums.TaskCommonStatus) error
	// FindByBlob find the artifact proxy task by blob.
	FindByBlob(ctx context.Context, blob string) ([]*models.ProxyArtifactTask, error)
}

type proxyService struct {
	tx *query.Query
}

// NewProxyService creates a new proxy service.
func NewProxyService(txs ...*query.Query) ProxyService {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &proxyService{
		tx: tx,
	}
}

// SaveProxyArtifact save a new artifact proxy task if conflict update.
func (s *proxyService) SaveProxyArtifact(ctx context.Context, task *models.ProxyArtifactTask) error {
	return s.tx.ProxyArtifactTask.WithContext(ctx).Save(task)
}

// UpdateProxyArtifactStatus update the artifact proxy task status.
func (s *proxyService) UpdateProxyArtifactStatus(ctx context.Context, id uint64, status enums.TaskCommonStatus) error {
	_, err := s.tx.ProxyArtifactTask.WithContext(ctx).Where(s.tx.ProxyArtifactTask.ID.Eq(id)).
		UpdateColumns(map[string]interface{}{
			"status": status,
		})
	return err
}

// FindByBlob find the artifact proxy task by blob.
func (s *proxyService) FindByBlob(ctx context.Context, blob string) ([]*models.ProxyArtifactTask, error) {
	return s.tx.ProxyArtifactTask.WithContext(ctx).
		LeftJoin(s.tx.ProxyArtifactBlob, s.tx.ProxyArtifactBlob.ProxyArtifactTaskID.EqCol(s.tx.ProxyArtifactTask.ID), s.tx.ProxyArtifactBlob.Blob.Eq(blob)).
		Where(s.tx.ProxyArtifactBlob.Blob.IsNotNull()). // implement the inner join
		Preload(s.tx.ProxyArtifactTask.Blobs).
		Find()
}
