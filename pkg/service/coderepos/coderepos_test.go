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
	repocoderepomocks "github.com/go-sigma/sigma/pkg/dal/repository/coderepo/mocks"
	repousermocks "github.com/go-sigma/sigma/pkg/dal/repository/user/mocks"
	"github.com/go-sigma/sigma/pkg/server/errcode"
)

func TestCodeRepositoryQueries(t *testing.T) {
	ctrl := gomock.NewController(t)
	codeRepository := repocoderepomocks.NewMockCodeRepositoryRepository(ctrl)
	userRepository := repousermocks.NewMockUserRepository(ctrl)
	svc := &service{Config: scmConfig(), RepoCode: codeRepository, RepoUser: userRepository}

	// A repository id is user input: a repository of another binding must not be reachable.
	userRepository.EXPECT().
		GetUser3rdPartyByProvider(gomock.Any(), "user-1", enums.ProviderGithub).
		Return(&models.User3rdParty{ID: "3rdparty-1"}, nil)
	codeRepository.EXPECT().
		Get(gomock.Any(), "repository-1").
		Return(&models.CodeRepository{ID: "repository-1", User3rdPartyID: "3rdparty-other"}, nil)

	_, err := svc.ListCodeRepositoryBranches(t.Context(), "user-1", enums.ProviderGithub, "repository-1")
	require.Error(t, err)
	require.ErrorIs(t, err, errcode.HTTPErrCodeNotFound)

	// Branches are not synced, so a repository without a credential cannot list them either.
	userRepository.EXPECT().
		GetUser3rdPartyByProvider(gomock.Any(), "user-1", enums.ProviderGithub).
		Return(&models.User3rdParty{ID: "3rdparty-1"}, nil)
	codeRepository.EXPECT().
		Get(gomock.Any(), "repository-1").
		Return(&models.CodeRepository{ID: "repository-1", User3rdPartyID: "3rdparty-1"}, nil)
	codeRepository.EXPECT().
		GetCloneCredential(gomock.Any(), "3rdparty-1").
		Return(&models.CodeRepositoryCloneCredential{Type: enums.ScmCredentialTypeSsh}, nil)

	_, err = svc.ListCodeRepositoryBranches(t.Context(), "user-1", enums.ProviderGithub, "repository-1")
	require.Error(t, err)

	// An unknown provider is rejected before any database access.
	_, err = svc.ListCodeRepositoryBranches(t.Context(), "user-1", enums.ProviderGitlab, "repository-1")
	require.Error(t, err)
}

func TestGetCodeRepositoryProviderError(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepository := repousermocks.NewMockUserRepository(ctrl)
	expectedErr := errors.New("lookup failed")
	userRepository.EXPECT().
		GetUser3rdPartyByProvider(gomock.Any(), "user-1", enums.ProviderGithub).
		Return(nil, expectedErr)

	_, err := (&service{RepoUser: userRepository}).
		GetCodeRepositoryUser3rdParty(t.Context(), "user-1", enums.ProviderGithub)
	require.Error(t, err)
}
