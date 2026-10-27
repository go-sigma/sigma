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
	"log/slog"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	repocoderepo "github.com/go-sigma/sigma/pkg/dal/repository/coderepo"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

// diff reconciles the stored repositories and owners of a third-party account with the
// provider. Branches are not synced: they are fetched from the provider when a build form
// needs them, which keeps a sync from making one API call per repository.
func (cr codeRepository) diff(ctx context.Context, user3rdPartyObj *models.User3rdParty, newRepos []*models.CodeRepository) error {
	codeRepositoryRepository := cr.codeRepositoryRepository
	oldRepos, err := codeRepositoryRepository.ListAll(ctx, user3rdPartyObj.ID)
	if err != nil {
		slog.Error("list all old repositories failed", "err", err)
		return fmt.Errorf("list all old repositories failed: %v", err)
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

	oldOwners, err := codeRepositoryRepository.ListOwnersAll(ctx, user3rdPartyObj.ID)
	if err != nil {
		slog.Error("list all old repository owners failed", "err", err)
		return fmt.Errorf("list all old repository owners failed: %v", err)
	}

	needUpdateOwners := make([]*models.CodeRepositoryOwner, 0, len(newRepos))
	needDelOwners := make([]*models.CodeRepositoryOwner, 0, len(oldOwners))
	for _, oldOwner := range oldOwners {
		found := false
		for _, newRepo := range newRepos {
			if oldOwner.Owner == newRepo.Owner {
				needUpdateOwners = append(needUpdateOwners, &models.CodeRepositoryOwner{
					ID:             oldOwner.ID,
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
				ID:             uuid.NewV7String(),
				User3rdPartyID: user3rdPartyObj.ID,
				OwnerID:        newRepo.OwnerID,
				Owner:          newRepo.Owner,
				IsOrg:          newRepo.IsOrg,
			})
			uniqueOwner = uniqueOwner.Insert(newRepo.Owner)
		}
	}

	err = query.Q.Transaction(func(tx *query.Query) error {
		codeRepositoryRepository := repocoderepo.NewCodeRepositoryRepository(tx)
		if len(needInsertRepos) > 0 {
			err := codeRepositoryRepository.CreateInBatches(ctx, needInsertRepos)
			if err != nil {
				slog.Error("create new repositories failed", "err", err)
				return fmt.Errorf("create new repositories failed: %v", err)
			}
		}
		if len(needUpdateRepos) > 0 {
			err := codeRepositoryRepository.UpdateInBatches(ctx, needUpdateRepos)
			if err != nil {
				slog.Error("update new repositories failed", "err", err)
				return fmt.Errorf("update new repositories failed: %v", err)
			}
		}
		if len(needDelRepos) > 0 {
			var needDelRepoIDs = make([]string, 0, len(needDelRepos))
			for _, r := range needDelRepos {
				needDelRepoIDs = append(needDelRepoIDs, r.ID)
			}
			err := codeRepositoryRepository.DeleteInBatches(ctx, needDelRepoIDs)
			if err != nil {
				slog.Error("delete old repositories failed", "err", err)
				return fmt.Errorf("delete old repositories failed: %v", err)
			}
		}
		if len(needInsertOwners) > 0 {
			err := codeRepositoryRepository.CreateOwnersInBatches(ctx, needInsertOwners)
			if err != nil {
				slog.Error("create new code repository owners failed", "err", err)
				return fmt.Errorf("create new code repository owner failed: %v", err)
			}
		}
		if len(needUpdateOwners) > 0 {
			err := codeRepositoryRepository.UpdateOwnersInBatches(ctx, needUpdateOwners)
			if err != nil {
				slog.Error("update new code repository owners failed", "err", err)
				return fmt.Errorf("update new code repository owner failed: %v", err)
			}
		}
		if len(needDelOwners) > 0 {
			var needDelRepoOwnerIDs = make([]string, 0, len(needDelOwners))
			for _, r := range needDelOwners {
				needDelRepoOwnerIDs = append(needDelRepoOwnerIDs, r.ID)
			}
			err := codeRepositoryRepository.DeleteOwnerInBatches(ctx, needDelRepoOwnerIDs)
			if err != nil {
				slog.Error("delete old code repository owners failed", "err", err)
				return fmt.Errorf("delete old code repository owners failed: %v", err)
			}
		}
		return nil
	})
	return err
}
