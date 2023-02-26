package dao

import (
	"context"
	"fmt"

	"github.com/docker/distribution/reference"
	"github.com/opencontainers/go-digest"
	"github.com/ximager/ximager/pkg/dal/models"
	"github.com/ximager/ximager/pkg/dal/query"
)

// ReferenceService defines the operations related to reference.
type ReferenceService interface {
	Get(ctx context.Context, repository, ref string) (*models.Reference, error)
}

type referenceService struct {
	tx *query.Query
}

// NewReferenceService creates a new reference service.
func NewReferenceService(txs ...*query.Query) ReferenceService {
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
		tagService := NewTagService(s.tx)
		tag, err := tagService.GetByName(ctx, repository, ref)
		if err != nil {
			return nil, err
		}
		return &models.Reference{
			Tag:      []*models.Tag{tag},
			Artifact: tag.Artifact,
		}, nil
	}
	artifactService := NewArtifactService(s.tx)
	artifact, err := artifactService.GetByDigest(ctx, repository, dgest.String())
	if err != nil {
		return nil, err
	}
	return &models.Reference{
		Artifact: artifact,
		Tag:      artifact.Tags,
	}, nil
}
