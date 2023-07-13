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

package gc

import (
	"context"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/go-sigma/sigma/pkg/dal/models"
)

func (g gc) gcArtifact(ctx context.Context, scope string) error {
	namespaceService := g.namespaceServiceFactory.New()
	var namespaceObjs []*models.Namespace
	if scope != "" {
		namespaceObj, err := namespaceService.GetByName(ctx, scope)
		if err != nil {
			return err
		}
		namespaceObjs = []*models.Namespace{namespaceObj}
	} else {
		var err error
		namespaceObjs, err = namespaceService.FindAll(ctx)
		if err != nil {
			return err
		}
	}

	timeTarget := time.Now().Add(-1 * viper.GetDuration("daemon.gc.retention"))

	repositoryService := g.repositoryServiceFactory.New()
	artifactService := g.artifactServiceFactory.New()
	for _, namespaceObj := range namespaceObjs {
		var repositoryCurIndex int64
		for {
			repositoryObjs, err := repositoryService.FindAll(ctx, namespaceObj.ID, pagination, repositoryCurIndex)
			if err != nil {
				return err
			}
			for _, repositoryObj := range repositoryObjs {
				var artifactCurIndex int64
				for {
					artifactObjs, err := artifactService.FindWithLastPull(ctx, repositoryObj.ID, timeTarget, pagination, artifactCurIndex)
					if err != nil {
						return err
					}
					var artifactIDs = make([]int64, 0, pagination)
					for _, artifactObj := range artifactObjs {
						artifactIDs = append(artifactIDs, artifactObj.ID)
					}
					associateArtifactIDs, err := artifactService.FindAssociateWithArtifact(ctx, artifactIDs)
					if err != nil {
						return err
					}
					associateTagIDs, err := artifactService.FindAssociateWithTag(ctx, artifactIDs)
					if err != nil {
						return err
					}
					artifactSets := mapset.NewSet(artifactIDs...)
					artifactSets.RemoveAll(associateArtifactIDs...)
					artifactSets.RemoveAll(associateTagIDs...)

					artifactSlices := artifactSets.ToSlice()
					if len(artifactSlices) > 0 {
						err = artifactService.DeleteByIDs(ctx, artifactSlices)
						if err != nil {
							return err
						}
						var digests []string
						for _, a := range artifactObjs {
							digests = append(digests, a.Digest)
						}
						log.Info().Ints64("id", artifactSlices).Strs("digest", digests).Msg("Delete artifact success")
					}

					if len(artifactObjs) < pagination {
						break
					}
					artifactCurIndex = artifactObjs[len(artifactObjs)-1].ID
				}
			}
			if len(repositoryObjs) < pagination {
				break
			}
			repositoryCurIndex = repositoryObjs[len(repositoryObjs)-1].ID
		}
	}
	return nil
}
