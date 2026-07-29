// Copyright 2026 sigma
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

package coderepos

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/dal/models"
	repocoderepo "github.com/go-sigma/sigma/pkg/dal/repository/coderepo"
	repouser "github.com/go-sigma/sigma/pkg/dal/repository/user"
)

func TestCodeRepositoryQueries(t *testing.T) {
	ctrl := gomock.NewController(t)
	codeRepository := repocoderepo.NewMockCodeRepositoryRepository(ctrl)
	service := &codeRepositoryService{codeRepositoryRepository: codeRepository}

	codeRepository.EXPECT().
		ListBranchesWithoutPagination(gomock.Any(), "repository-1").
		Return([]*models.CodeRepositoryBranch{{ID: "branch-1"}}, int64(1), nil)
	branches, total, err := service.ListCodeRepositoryBranches(t.Context(), "repository-1")
	require.NoError(t, err)
	require.Len(t, branches, 1)
	require.Equal(t, int64(1), total)

	codeRepository.EXPECT().
		GetBranchByName(gomock.Any(), "repository-1", "main").
		Return(&models.CodeRepositoryBranch{ID: "branch-1", Name: "main"}, nil)
	branch, err := service.GetCodeRepositoryBranch(t.Context(), "repository-1", "main")
	require.NoError(t, err)
	require.Equal(t, "branch-1", branch.ID)
}

func TestGetCodeRepositoryProviderError(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepository := repouser.NewMockUserRepository(ctrl)
	expectedErr := errors.New("lookup failed")
	userRepository.EXPECT().
		GetUser3rdPartyByProvider(gomock.Any(), "user-1", enums.ProviderGithub).
		Return(nil, expectedErr)

	_, err := (&codeRepositoryService{userRepository: userRepository}).
		GetCodeRepositoryUser3rdParty(t.Context(), "user-1", enums.ProviderGithub)
	require.Error(t, err)
}
