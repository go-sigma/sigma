// Copyright 2024 sigma
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

package coderepo_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	repocoderepo "github.com/go-sigma/sigma/pkg/dal/repository/coderepo"
	"github.com/go-sigma/sigma/pkg/testkit"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

func TestNewCodeRepositoryRepository(t *testing.T) {
	require.NotNil(t, repocoderepo.NewCodeRepositoryRepository())
	require.NotNil(t, repocoderepo.NewCodeRepositoryRepository(query.Q))
}

func TestCodeRepositoryRepository(t *testing.T) {
	testkit.InitRepository(t)

	ctx := t.Context()
	codeRepositoryRepository := repocoderepo.NewCodeRepositoryRepository()
	password := "password"
	accountID := "account-1"
	userObj := &models.User{
		ID:       uuid.NewV7String(),
		Username: "coderepo-user",
		Password: &password,
	}
	user3rdPartyObj := &models.User3rdParty{
		ID:        uuid.NewV7String(),
		UserID:    userObj.ID,
		Provider:  enums.ProviderGithub,
		AccountID: &accountID,
	}
	require.NoError(t, query.Q.User.WithContext(ctx).Create(userObj))
	require.NoError(t, query.Q.User3rdParty.WithContext(ctx).Create(user3rdPartyObj))

	repoOne := &models.CodeRepository{
		ID:             uuid.NewV7String(),
		RepositoryID:   "1001",
		User3rdPartyID: user3rdPartyObj.ID,
		OwnerID:        "owner-1",
		Owner:          "sigma",
		Name:           "sigma/api",
		SshUrl:         "git@github.com:go-sigma/sigma-api.git",
		CloneUrl:       "https://github.com/go-sigma/sigma-api.git",
	}
	repoTwo := &models.CodeRepository{
		ID:             uuid.NewV7String(),
		RepositoryID:   "1002",
		User3rdPartyID: user3rdPartyObj.ID,
		OwnerID:        "owner-1",
		Owner:          "sigma",
		Name:           "sigma/worker",
		SshUrl:         "git@github.com:go-sigma/sigma-worker.git",
		CloneUrl:       "https://github.com/go-sigma/sigma-worker.git",
	}
	require.NoError(t, codeRepositoryRepository.CreateInBatches(ctx, []*models.CodeRepository{repoOne, repoTwo}))

	ownerObj := &models.CodeRepositoryOwner{
		ID:             uuid.NewV7String(),
		User3rdPartyID: user3rdPartyObj.ID,
		OwnerID:        "owner-1",
		Owner:          "sigma",
		IsOrg:          true,
	}
	require.NoError(t, codeRepositoryRepository.CreateOwnersInBatches(ctx, []*models.CodeRepositoryOwner{ownerObj}))

	branchOne := &models.CodeRepositoryBranch{
		ID:               uuid.NewV7String(),
		CodeRepositoryID: repoOne.ID,
		Name:             "main",
	}
	branchTwo := &models.CodeRepositoryBranch{
		ID:               uuid.NewV7String(),
		CodeRepositoryID: repoOne.ID,
		Name:             "release",
	}
	require.NoError(t, codeRepositoryRepository.CreateBranchesInBatches(ctx, []*models.CodeRepositoryBranch{branchOne, branchTwo}))

	repositories, err := codeRepositoryRepository.ListAll(ctx, user3rdPartyObj.ID)
	require.NoError(t, err)
	require.Len(t, repositories, 2)

	got, err := codeRepositoryRepository.Get(ctx, repoOne.ID)
	require.NoError(t, err)
	require.Equal(t, user3rdPartyObj.ID, got.User3rdParty.ID)

	owners, err := codeRepositoryRepository.ListOwnersAll(ctx, user3rdPartyObj.ID)
	require.NoError(t, err)
	require.Len(t, owners, 1)
	require.Equal(t, "sigma", owners[0].Owner)

	method := enums.SortMethodAsc
	name := "api"
	owner := "sigma"
	page := 1
	limit := 10
	sortBy := "name"
	repositories, total, err := codeRepositoryRepository.ListWithPagination(ctx, userObj.ID, enums.ProviderGithub, &owner, &name, api.Pagination{
		Page:  &page,
		Limit: &limit,
	}, api.Sortable{Sort: &sortBy, Method: &method})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, repositories, 1)
	require.Equal(t, repoOne.ID, repositories[0].ID)

	ownerFilter := "sig"
	owners, total, err = codeRepositoryRepository.ListOwnerWithoutPagination(ctx, userObj.ID, enums.ProviderGithub, &ownerFilter)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, owners, 1)

	branches, total, err := codeRepositoryRepository.ListBranchesWithoutPagination(ctx, repoOne.ID)
	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, branches, 2)

	branch, err := codeRepositoryRepository.GetBranchByName(ctx, repoOne.ID, "release")
	require.NoError(t, err)
	require.Equal(t, branchTwo.ID, branch.ID)

	repoOne.Owner = "go-sigma"
	repoOne.Name = "go-sigma/api"
	repoOne.SshUrl = "git@github.com:go-sigma/api.git"
	repoOne.CloneUrl = "https://github.com/go-sigma/api.git"
	require.NoError(t, codeRepositoryRepository.UpdateInBatches(ctx, []*models.CodeRepository{repoOne}))

	got, err = codeRepositoryRepository.Get(ctx, repoOne.ID)
	require.NoError(t, err)
	require.Equal(t, "go-sigma/api", got.Name)

	ownerObj.Owner = "go-sigma"
	ownerObj.IsOrg = false
	require.NoError(t, codeRepositoryRepository.UpdateOwnersInBatches(ctx, []*models.CodeRepositoryOwner{ownerObj}))

	owners, err = codeRepositoryRepository.ListOwnersAll(ctx, user3rdPartyObj.ID)
	require.NoError(t, err)
	require.Equal(t, "go-sigma", owners[0].Owner)
	require.False(t, owners[0].IsOrg)

	require.NoError(t, codeRepositoryRepository.DeleteBranchesInBatches(ctx, []string{branchOne.ID, branchTwo.ID}))
	branches, total, err = codeRepositoryRepository.ListBranchesWithoutPagination(ctx, repoOne.ID)
	require.NoError(t, err)
	require.Zero(t, total)
	require.Empty(t, branches)

	require.NoError(t, codeRepositoryRepository.DeleteOwnerInBatches(ctx, []string{ownerObj.ID}))
	owners, err = codeRepositoryRepository.ListOwnersAll(ctx, user3rdPartyObj.ID)
	require.NoError(t, err)
	require.Empty(t, owners)

	require.NoError(t, codeRepositoryRepository.DeleteInBatches(ctx, []string{repoOne.ID, repoTwo.ID}))
	repositories, err = codeRepositoryRepository.ListAll(ctx, user3rdPartyObj.ID)
	require.NoError(t, err)
	require.Empty(t, repositories)
}

func TestCodeRepositoryCloneCredential(t *testing.T) {
	testkit.InitRepository(t)

	ctx := t.Context()
	codeRepositoryRepository := repocoderepo.NewCodeRepositoryRepository()
	user3rdPartyID := uuid.NewV7String()
	token := "token"
	credential := &models.CodeRepositoryCloneCredential{
		ID:             uuid.NewV7String(),
		User3rdPartyID: user3rdPartyID,
		Type:           enums.ScmCredentialTypeToken,
		Token:          &token,
	}
	require.NoError(t, query.Q.CodeRepositoryCloneCredential.WithContext(ctx).Create(credential))

	got, err := codeRepositoryRepository.GetCloneCredential(ctx, user3rdPartyID)
	require.NoError(t, err)
	require.Equal(t, credential.ID, got.ID)
	require.Equal(t, "token", ptr.To(got.Token))
}
