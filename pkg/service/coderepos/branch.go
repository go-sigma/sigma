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
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"

	"code.gitea.io/sdk/gitea"
	"github.com/google/go-github/v74/github"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

// branchesPerPage is the page size used when paging through the provider branch API.
const branchesPerPage = 100

// ListCodeRepositoryBranches lists the branches of a code repository.
//
// Branches are not synced into the database: they change far more often than repositories and
// fetching them up front cost one API call per repository on every sync. They are read from
// the provider when a build form needs them.
func (s *service) ListCodeRepositoryBranches(ctx context.Context, userID string, provider enums.Provider, codeRepositoryID string) ([]string, error) {
	codeRepositoryObj, token, baseURL, err := s.scmTarget(ctx, userID, provider, codeRepositoryID)
	if err != nil {
		return nil, err
	}
	branches, err := listScmBranches(ctx, provider, baseURL, token, codeRepositoryObj.Owner, codeRepositoryObj.Name)
	if err != nil {
		return nil, err
	}
	return branches, nil
}

// GetCodeRepositoryBranch verifies that a branch exists and returns its name.
func (s *service) GetCodeRepositoryBranch(ctx context.Context, userID string, provider enums.Provider, codeRepositoryID string, name string) (string, error) {
	branches, err := s.ListCodeRepositoryBranches(ctx, userID, provider, codeRepositoryID)
	if err != nil {
		return "", err
	}
	if slices.Contains(branches, name) {
		return name, nil
	}
	return "", errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Branch %s not found", name))
}

// scmTarget resolves everything needed to call a provider API for one code repository: the
// repository itself, the decrypted access token and the provider base URL.
func (s *service) scmTarget(ctx context.Context, userID string, provider enums.Provider, codeRepositoryID string) (*models.CodeRepository, string, string, error) {
	scmProvider, err := s.mustScmProvider(provider)
	if err != nil {
		return nil, "", "", err
	}

	user3rdPartyObj, err := s.RepoUser.GetUser3rdPartyByProvider(ctx, userID, provider)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", "", errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("No %s account is linked", provider.String()))
		}
		slog.Error("get user 3rdparty by provider failed", "err", err, "provider", provider.String())
		return nil, "", "", errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get user 3rdparty by provider failed: %v", err))
	}

	codeRepositoryObj, err := s.RepoCode.Get(ctx, codeRepositoryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", "", errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Code repository %s not found", codeRepositoryID))
		}
		slog.Error("get code repository failed", "err", err, "id", codeRepositoryID)
		return nil, "", "", errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get code repository failed: %v", err))
	}
	// A repository id is user input: never talk to a provider with another account's binding.
	if codeRepositoryObj.User3rdPartyID != user3rdPartyObj.ID {
		return nil, "", "", errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Code repository %s not found", codeRepositoryID))
	}

	credentialObj, err := s.RepoCode.GetCloneCredential(ctx, user3rdPartyObj.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", "", errcode.HTTPErrCodeBadRequest.Detail(fmt.Sprintf("No credential is configured for %s", provider.String()))
		}
		slog.Error("get code repository credential failed", "err", err, "provider", provider.String())
		return nil, "", "", errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get code repository credential failed: %v", err))
	}
	if credentialObj.Type != enums.ScmCredentialTypeToken || credentialObj.Token == nil || *credentialObj.Token == "" {
		return nil, "", "", errcode.HTTPErrCodeBadRequest.Detail(fmt.Sprintf(
			"Listing %s branches requires a token credential", provider.String()))
	}
	token, err := s.SvcCredential.Open(ptr.To(credentialObj.Token))
	if err != nil {
		slog.Error("decrypt code repository credential failed", "err", err, "provider", provider.String())
		return nil, "", "", errcode.HTTPErrCodeInternalError.Detail("Decrypt code repository credential failed")
	}

	return codeRepositoryObj, token, scmProvider.BaseURL, nil
}

// listScmBranches pages through the provider branch API of one repository.
func listScmBranches(ctx context.Context, provider enums.Provider, baseURL, token, owner, name string) ([]string, error) {
	failed := func(err error) ([]string, error) {
		slog.Error("list scm branches failed", "err", err, "provider", provider.String(), "owner", owner, "repo", name)
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("List branches failed: %v", err))
	}

	var branches []string
	page := 1
	for {
		var names []string
		var err error
		switch provider {
		case enums.ProviderGithub:
			var rs []*github.Branch
			rs, _, err = github.NewTokenClient(ctx, token).Repositories.ListBranches(ctx, owner, name,
				&github.BranchListOptions{Page: page, PerPage: branchesPerPage})
			for _, r := range rs {
				names = append(names, r.GetName())
			}
		case enums.ProviderGitlab:
			client, cerr := gitlab.NewClient(token, gitlab.WithBaseURL(baseURL))
			if cerr != nil {
				return failed(cerr)
			}
			var rs []*gitlab.Branch
			rs, _, err = client.Branches.ListBranches(owner+"/"+name,
				&gitlab.ListBranchesOptions{Page: int64(page), PerPage: branchesPerPage})
			for _, r := range rs {
				names = append(names, r.Name)
			}
		case enums.ProviderGitea:
			client, cerr := gitea.NewClient(baseURL, gitea.SetToken(token), gitea.SetContext(ctx))
			if cerr != nil {
				return failed(cerr)
			}
			var rs []*gitea.Branch
			rs, _, err = client.ListRepoBranches(owner, name,
				gitea.ListRepoBranchesOptions{Page: page, PageSize: branchesPerPage})
			for _, r := range rs {
				names = append(names, r.Name)
			}
		default:
			return nil, errcode.HTTPErrCodeBadRequest.Detail(fmt.Sprintf("unsupported scm provider %s", provider.String()))
		}
		if err != nil {
			return failed(err)
		}
		branches = append(branches, names...)
		if len(names) < branchesPerPage {
			return branches, nil
		}
		page++
	}
}
