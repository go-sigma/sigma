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

package coderepo

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
)

func (cr codeRepository) diff(ctx context.Context, user3rdPartyObj *models.User3rdParty, newRepos []*models.CodeRepository) error {
	codeRepositoryService := cr.codeRepositoryServiceFactory.New()
	oldRepos, err := codeRepositoryService.ListAll(ctx, user3rdPartyObj.ID)
	if err != nil {
		log.Error().Err(err).Msg("List all old repositories failed")
		return fmt.Errorf("List all old repositories failed: %v", err)
	}

	needUpdateRepos := make([]*models.CodeRepository, 0, len(newRepos))
	needDelRepos := make([]*models.CodeRepository, 0, len(oldRepos))
	for _, oldRepo := range oldRepos {
		found := false
		for _, newRepo := range newRepos {
			if oldRepo.RepositoryID == newRepo.RepositoryID {
				needUpdateRepos = append(needUpdateRepos, newRepo)
				found = true
				break
			}
		}
		if !found {
			needDelRepos = append(needDelRepos, oldRepo)
		}
	}

	needInsertRepos := make([]*models.CodeRepository, 0, len(newRepos))
	for _, newRepo := range newRepos {
		found := false
		for _, oldRepo := range oldRepos {
			if oldRepo.RepositoryID == newRepo.RepositoryID {
				found = true
				break
			}
		}
		if !found {
			needInsertRepos = append(needInsertRepos, newRepo)
		}
	}

	oldOwners, err := codeRepositoryService.ListOwnersAll(ctx, user3rdPartyObj.ID)
	if err != nil {
		log.Error().Err(err).Msg("List all old repository owners failed")
		return fmt.Errorf("List all old repository owners failed: %v", err)
	}

	needUpdateOwners := make([]*models.CodeRepositoryOwner, 0, len(newRepos))
	needDelOwners := make([]*models.CodeRepositoryOwner, 0, len(oldOwners))
	for _, oldOwner := range oldOwners {
		found := false
		for _, newRepo := range newRepos {
			if oldOwner.Owner == newRepo.Owner {
				needUpdateOwners = append(needUpdateOwners, &models.CodeRepositoryOwner{
					User3rdPartyID: newRepo.User3rdPartyID,
					OwnerID:        newRepo.OwnerID,
					Owner:          newRepo.Owner,
					IsOrg:          newRepo.IsOrg,
				})
				found = true
				break
			}
		}
		if !found {
			needDelOwners = append(needDelOwners, oldOwner)
		}
	}
	uniqueOwner := sets.New[string]()
	needInsertOwners := make([]*models.CodeRepositoryOwner, 0, len(oldOwners))
	for _, newRepo := range newRepos {
		found := false
		for _, oldOwner := range oldOwners {
			if oldOwner.Owner == newRepo.Owner {
				found = true
				break
			}
		}
		if !found {
			if uniqueOwner.Has(newRepo.Owner) {
				continue
			}
			needInsertOwners = append(needInsertOwners, &models.CodeRepositoryOwner{
				User3rdPartyID: user3rdPartyObj.ID,
				OwnerID:        newRepo.OwnerID,
				Owner:          newRepo.Owner,
				IsOrg:          newRepo.IsOrg,
			})
			uniqueOwner = uniqueOwner.Insert(newRepo.Owner)
		}
	}

	err = query.Q.Transaction(func(tx *query.Query) error {
		codeRepositoryService := cr.codeRepositoryServiceFactory.New(tx)
		if len(needInsertRepos) > 0 {
			err := codeRepositoryService.CreateInBatches(ctx, needInsertRepos)
			if err != nil {
				log.Error().Err(err).Msg("Create new repositories failed")
				return fmt.Errorf("Create new repositories failed: %v", err)
			}
		}
		if len(needUpdateRepos) > 0 {
			err := codeRepositoryService.UpdateInBatches(ctx, needUpdateRepos)
			if err != nil {
				log.Error().Err(err).Msg("Update new repositories failed")
				return fmt.Errorf("Update new repositories failed: %v", err)
			}
		}
		if len(needDelRepos) > 0 {
			var needDelRepoIDs = make([]int64, 0, len(needDelRepos))
			for _, r := range needDelRepos {
				needDelRepoIDs = append(needDelRepoIDs, r.ID)
			}
			err := codeRepositoryService.DeleteInBatches(ctx, needDelRepoIDs)
			if err != nil {
				log.Error().Err(err).Msg("Delete old repositories failed")
				return fmt.Errorf("Delete old repositories failed: %v", err)
			}
		}
		if len(needInsertOwners) > 0 {
			err := codeRepositoryService.CreateOwnersInBatches(ctx, needInsertOwners)
			if err != nil {
				log.Error().Err(err).Msg("Create new code repository owners failed")
				return fmt.Errorf("Create new code repository owner failed: %v", err)
			}
		}
		if len(needUpdateOwners) > 0 {
			err := codeRepositoryService.UpdateOwnersInBatches(ctx, needUpdateOwners)
			if err != nil {
				log.Error().Err(err).Msg("Update new code repository owners failed")
				return fmt.Errorf("Update new code repository owner failed: %v", err)
			}
		}
		if len(needDelOwners) > 0 {
			var needDelRepoOwnerIDs = make([]int64, 0, len(needDelOwners))
			for _, r := range needDelOwners {
				needDelRepoOwnerIDs = append(needDelRepoOwnerIDs, r.ID)
			}
			err := codeRepositoryService.DeleteOwnerInBatches(ctx, needDelRepoOwnerIDs)
			if err != nil {
				log.Error().Err(err).Msg("Delete old repositories failed")
				return fmt.Errorf("Delete old repositories failed: %v", err)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
