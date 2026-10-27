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
	"strconv"
	"strings"

	"github.com/google/go-github/v74/github"

	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

func (cr codeRepository) github(ctx context.Context, user3rdPartyObj *models.User3rdParty, credential scmCredential) error {
	client := github.NewTokenClient(ctx, credential.Token)

	userObj, _, err := client.Users.Get(ctx, "")
	if err != nil {
		slog.Error("get user info failed", "err", err)
		return fmt.Errorf("get user info failed: %v", err)
	}

	var repos []*github.Repository

	page := 1
	for {
		rs, _, err := client.Repositories.ListByAuthenticatedUser(ctx,
			&github.RepositoryListByAuthenticatedUserOptions{Page: page, PerPage: perPage})
		if err != nil {
			slog.Error("list repositories failed", "err", err)
			return fmt.Errorf("list repositories failed: %v", err)
		}
		for _, r := range rs {
			if strings.HasPrefix(ptr.To(r.FullName), fmt.Sprintf("%s/", ptr.To(userObj.Login))) {
				repos = append(repos, r)
			}
		}
		if len(rs) < perPage {
			break
		}
		page++
	}

	var orgs []*github.Organization

	page = 1
	for {
		os, _, err := client.Organizations.List(ctx, "", &github.ListOptions{Page: page, PerPage: perPage})
		if err != nil {
			slog.Error("list organizations failed", "err", err)
			return fmt.Errorf("list organizations failed: %v", err)
		}
		orgs = append(orgs, os...)
		if len(os) < perPage {
			break
		}
		page++
	}

	for _, o := range orgs {
		page = 1
		for {
			rs, _, err := client.Repositories.ListByOrg(ctx, ptr.To(o.Login),
				&github.RepositoryListByOrgOptions{Page: page, PerPage: perPage})
			if err != nil {
				slog.Error("list repositories for orgs failed", "err", err)
				return fmt.Errorf("list repositories for orgs failed: %v", err)
			}
			repos = append(repos, rs...)
			if len(rs) < perPage {
				break
			}
			page++
		}
	}

	var newRepos = make([]*models.CodeRepository, 0, len(repos))
	for _, r := range repos {
		repo := &models.CodeRepository{
			ID:             uuid.NewV7String(),
			User3rdPartyID: user3rdPartyObj.ID,
			RepositoryID:   strconv.FormatInt(r.GetID(), 10),
			OwnerID:        strconv.FormatInt(ptr.To(ptr.To(r.Owner).ID), 10),
			Owner:          ptr.To(ptr.To(r.Owner).Login),
			Name:           r.GetName(),
			SshUrl:         r.GetSSHURL(),
			CloneUrl:       r.GetCloneURL(),
			PushedAt:       r.GetPushedAt().UnixMilli(),
		}
		if ptr.To(ptr.To(r.Owner).Login) != ptr.To(userObj.Login) {
			repo.IsOrg = true
		}
		newRepos = append(newRepos, repo)
	}

	return cr.diff(ctx, user3rdPartyObj, newRepos)
}
