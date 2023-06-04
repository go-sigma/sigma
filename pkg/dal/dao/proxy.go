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
)

//go:generate mockgen -destination=mocks/proxy.go -package=mocks github.com/ximager/ximager/pkg/dal/dao ProxyTaskService
//go:generate mockgen -destination=mocks/proxy_factory.go -package=mocks github.com/ximager/ximager/pkg/dal/dao ProxyTaskServiceFactory

// ProxyTaskService defines the interface to access proxy task.
type ProxyTaskService interface {
	// SaveProxyTaskArtifact save a new artifact proxy task if conflict update.
	SaveProxyTaskArtifact(ctx context.Context, task *models.ProxyTaskArtifact) error
	// FindProxyTaskArtifactByBlob find the artifact proxy task by blob.
	FindProxyTaskArtifactByBlob(ctx context.Context, blob string) ([]*models.ProxyTaskArtifact, error)
	// SaveProxyTaskTag save a new proxy task tag if conflict update.
	SaveProxyTaskTag(ctx context.Context, task *models.ProxyTaskTag) error
	// FindProxyTaskTagByManifest find the proxy task tags by manifest.
	FindProxyTaskTagByManifest(ctx context.Context, blob string) ([]*models.ProxyTaskTag, error)
}

type proxyTaskService struct {
	tx *query.Query
}

// ProxyTaskServiceFactory is the interface that provides the proxy task service factory methods.
type ProxyTaskServiceFactory interface {
	New(txs ...*query.Query) ProxyTaskService
}

type proxyTaskServiceFactory struct{}

// NewProxyTaskServiceFactory creates a new proxy task service factory.
func NewProxyTaskServiceFactory() ProxyTaskServiceFactory {
	return &proxyTaskServiceFactory{}
}

func (f *proxyTaskServiceFactory) New(txs ...*query.Query) ProxyTaskService {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &proxyTaskService{
		tx: tx,
	}
}

// SaveProxyArtifact save a new artifact proxy task if conflict update.
func (s *proxyTaskService) SaveProxyTaskArtifact(ctx context.Context, task *models.ProxyTaskArtifact) error {
	return s.tx.ProxyTaskArtifact.WithContext(ctx).Save(task)
}

// FindByBlob find the artifact proxy task by blob.
func (s *proxyTaskService) FindProxyTaskArtifactByBlob(ctx context.Context, blob string) ([]*models.ProxyTaskArtifact, error) {
	return s.tx.ProxyTaskArtifact.WithContext(ctx).
		LeftJoin(s.tx.ProxyTaskArtifactBlob, s.tx.ProxyTaskArtifactBlob.ProxyTaskArtifactID.EqCol(s.tx.ProxyTaskArtifact.ID), s.tx.ProxyTaskArtifactBlob.Blob.Eq(blob)).
		Where(s.tx.ProxyTaskArtifactBlob.Blob.IsNotNull()). // implement the inner join
		Preload(s.tx.ProxyTaskArtifact.Blobs).
		Find()
}

// SaveProxyTaskTag save a new proxy task tag if conflict update.
func (s *proxyTaskService) SaveProxyTaskTag(ctx context.Context, task *models.ProxyTaskTag) error {
	return s.tx.ProxyTaskTag.WithContext(ctx).Save(task)
}

// FindProxyTaskTagByManifest find the proxy task tags by manifest.
func (s *proxyTaskService) FindProxyTaskTagByManifest(ctx context.Context, manifest string) ([]*models.ProxyTaskTag, error) {
	return s.tx.ProxyTaskTag.WithContext(ctx).
		LeftJoin(s.tx.ProxyTaskTagManifest, s.tx.ProxyTaskTagManifest.ProxyTaskTagID.EqCol(s.tx.ProxyTaskTag.ID), s.tx.ProxyTaskTagManifest.Digest.Eq(manifest)).
		Where(s.tx.ProxyTaskTagManifest.Digest.IsNotNull()). // implement the inner join
		Preload(s.tx.ProxyTaskTag.Manifests).
		Find()
}
