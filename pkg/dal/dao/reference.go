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
	"fmt"

	"github.com/docker/distribution/reference"
	"github.com/opencontainers/go-digest"

	"github.com/ximager/ximager/pkg/dal/models"
	"github.com/ximager/ximager/pkg/dal/query"
)

//go:generate mockgen -destination=mocks/reference.go -package=mocks github.com/ximager/ximager/pkg/dal/dao ReferenceService
//go:generate mockgen -destination=mocks/reference_factory.go -package=mocks github.com/ximager/ximager/pkg/dal/dao ReferenceServiceFactory

// ReferenceService defines the operations related to reference.
type ReferenceService interface {
	Get(ctx context.Context, repository, ref string) (*models.Reference, error)
}

type referenceService struct {
	tx *query.Query
}

// ReferenceServiceFactory is the interface that provides the reference service factory methods.
type ReferenceServiceFactory interface {
	New(txs ...*query.Query) ReferenceService
}

type referenceServiceFactory struct{}

// NewReferenceServiceFactory creates a new reference service factory.
func NewReferenceServiceFactory() ReferenceServiceFactory {
	return &referenceServiceFactory{}
}

func (f *referenceServiceFactory) New(txs ...*query.Query) ReferenceService {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &referenceService{
		tx: tx,
	}
}

// Get gets the reference with the specified repository name and reference.
func (s *referenceService) Get(ctx context.Context, repository, ref string) (*models.Reference, error) {
	dgest, err := digest.Parse(ref)
	if err != nil {
		if !reference.TagRegexp.MatchString(ref) {
			return nil, fmt.Errorf("not valid digest or tag")
		}
		tagServiceFactory := NewTagServiceFactory()
		tagService := tagServiceFactory.New(s.tx)
		tag, err := tagService.GetByName(ctx, repository, ref)
		if err != nil {
			return nil, err
		}
		return &models.Reference{
			Tag:      []*models.Tag{tag},
			Artifact: tag.Artifact,
		}, nil
	}
	artifactServiceFactory := NewArtifactServiceFactory()
	artifactService := artifactServiceFactory.New(s.tx)
	artifact, err := artifactService.GetByDigest(ctx, repository, dgest.String())
	if err != nil {
		return nil, err
	}
	return &models.Reference{
		Artifact: artifact,
		Tag:      artifact.Tags,
	}, nil
}
