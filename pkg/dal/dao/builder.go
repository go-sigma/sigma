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

	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
)

// BuilderService is the interface that provides methods to operate on Builder model
type BuilderService interface {
	// Create creates a new builder record in the database
	Create(ctx context.Context, audit *models.Builder) error
	// CreateLog creates a new BuilderLog record in the database
	CreateLog(ctx context.Context, log *models.BuilderLog) error
	// GetLog get log from object storage or database
	GetLog(ctx context.Context, id int64) (*models.BuilderLog, error)
}

type builderService struct {
	tx *query.Query
}

// BuilderServiceFactory is the interface that provides the builder service factory methods.
type BuilderServiceFactory interface {
	New(txs ...*query.Query) BuilderService
}

type builderServiceFactory struct{}

// NewBuilderServiceFactory creates a new builder service factory.
func NewBuilderServiceFactory() BuilderServiceFactory {
	return &builderServiceFactory{}
}

func (f *builderServiceFactory) New(txs ...*query.Query) BuilderService {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &builderService{
		tx: tx,
	}
}

// Create creates a new builder record in the database
func (s builderService) Create(ctx context.Context, builder *models.Builder) error {
	return s.tx.WithContext(ctx).Builder.Create(builder)
}

// CreateLog creates a new BuilderLog record in the database
func (s builderService) CreateLog(ctx context.Context, log *models.BuilderLog) error {
	return s.tx.WithContext(ctx).BuilderLog.Create(log)
}

// GetLog get log from object storage or database
func (s builderService) GetLog(ctx context.Context, id int64) (*models.BuilderLog, error) {
	return s.tx.WithContext(ctx).BuilderLog.Where(s.tx.BuilderLog.BuilderID.Eq(id)).First()
}
