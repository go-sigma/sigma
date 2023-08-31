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

func (cr codeRepository) diff(ctx context.Context, user3rdPartyObj *models.User3rdParty, newRepos []*models.CodeRepository, branchMap map[string][]*models.CodeRepositoryBranch) error {
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
	return cr.diffBranch(ctx, user3rdPartyObj, branchMap)
}

func (cr codeRepository) diffBranch(ctx context.Context, user3rdPartyObj *models.User3rdParty, branchMap map[string][]*models.CodeRepositoryBranch) error {
	if len(branchMap) == 0 {
		return nil
	}
	codeRepositoryService := cr.codeRepositoryServiceFactory.New()
	repositoryObjs, err := codeRepositoryService.ListAll(ctx, user3rdPartyObj.ID)
	if err != nil {
		log.Error().Err(err).Msg("List all repositories failed")
		return fmt.Errorf("List all repositories failed: %v", err)
	}

	var needInsertBranches []*models.CodeRepositoryBranch
	var needDelBranches []int64
	for _, repo := range repositoryObjs {
		oldBranches, _, err := codeRepositoryService.ListBranchesWithoutPagination(ctx, repo.ID)
		if err != nil {
			log.Error().Err(err).Int64("id", repo.ID).Msg("List repo branches failed")
			return fmt.Errorf("List repo branches failed: %v", err)
		}
		if len(branchMap[repo.RepositoryID]) == 0 {
			var bs []*models.CodeRepositoryBranch
			for _, b := range branchMap[repo.RepositoryID] {
				bs = append(bs, &models.CodeRepositoryBranch{CodeRepositoryID: repo.ID, Name: b.Name})
			}
			needInsertBranches = append(needInsertBranches, bs...)
			continue
		}

		for _, oldB := range oldBranches {
			found := false
			for _, newB := range branchMap[repo.RepositoryID] {
				if newB.Name == oldB.Name {
					found = true
				}
			}
			if !found {
				needDelBranches = append(needDelBranches, oldB.ID)
			}
		}

		for _, newB := range branchMap[repo.RepositoryID] {
			found := false
			for _, oldB := range oldBranches {
				if newB.Name == oldB.Name {
					found = true
				}
			}
			if !found {
				needInsertBranches = append(needInsertBranches, &models.CodeRepositoryBranch{
					CodeRepositoryID: repo.ID,
					Name:             newB.Name,
				})
			}
		}
	}
	err = query.Q.Transaction(func(tx *query.Query) error {
		codeRepositoryService := cr.codeRepositoryServiceFactory.New(tx)
		if len(needInsertBranches) > 0 {
			err := codeRepositoryService.CreateBranchesInBatches(ctx, needInsertBranches)
			if err != nil {
				log.Error().Err(err).Msg("Create new branches failed")
				return fmt.Errorf("Create new branches failed: %v", err)
			}
		}
		if len(needDelBranches) > 0 {
			err := codeRepositoryService.DeleteBranchesInBatches(ctx, needDelBranches)
			if err != nil {
				log.Error().Err(err).Msg("Delete branches failed")
				return fmt.Errorf("Delete branches failed: %v", err)
			}
		}
		return nil
	})
	return err
}
