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

	gitlab "gitlab.com/gitlab-org/api/client-go"
	"golang.org/x/oauth2"

	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

func (cr codeRepository) gitlab(ctx context.Context, user3rdPartyObj *models.User3rdParty) error {
	as := gitlab.OAuthTokenSource{
		TokenSource: oauth2.StaticTokenSource(&oauth2.Token{
			AccessToken: ptr.To(user3rdPartyObj.Token),
		}),
	}

	client, err := gitlab.NewAuthSourceClient(as)
	if err != nil {
		return err
	}

	userObj, _, err := client.Users.CurrentUser()
	if err != nil {
		return err
	}

	var repos []*gitlab.Project

	page := 1
	for {
		rs, _, err := client.Projects.ListProjects(&gitlab.ListProjectsOptions{
			Owned:       new(true),
			ListOptions: gitlab.ListOptions{Page: page, PerPage: perPage}})
		if err != nil {
			slog.Error("list projects from gitlab failed", "err", err)
			return fmt.Errorf("list projects from gitlab failed: %w", err)
		}
		for _, r := range rs {
			if r.Namespace.Path == userObj.Username {
				repos = append(repos, r)
			}
		}
		if len(rs) < perPage {
			break
		}
		page++
	}

	var groups []*gitlab.Group

	page = 1
	for {
		minAccessLevel := gitlab.ReporterPermissions
		gs, _, err := client.Groups.ListGroups(&gitlab.ListGroupsOptions{
			AllAvailable:   new(true),
			MinAccessLevel: &minAccessLevel,
			ListOptions:    gitlab.ListOptions{Page: page, PerPage: perPage}})
		if err != nil {
			slog.Error("list groups from gitlab failed", "err", err)
			return fmt.Errorf("list groups from gitlab failed: %w", err)
		}
		groups = append(groups, gs...)
		if len(gs) < perPage {
			break
		}
		page++
	}

	for _, g := range groups {
		page = 1
		for {
			minAccessLevel := gitlab.ReporterPermissions
			rs, _, err := client.Groups.ListGroupProjects(g.ID, &gitlab.ListGroupProjectsOptions{
				MinAccessLevel: &minAccessLevel,
				ListOptions:    gitlab.ListOptions{Page: page, PerPage: perPage}})
			if err != nil {
				slog.Error("list projects from gitlab failed", "err", err)
				return fmt.Errorf("list projects from gitlab failed: %w", err)
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
			RepositoryID:   strconv.Itoa(r.ID),
			OwnerID:        strconv.Itoa(r.Namespace.ID),
			Owner:          r.Namespace.Path,
			Name:           r.Name,
			SshUrl:         r.SSHURLToRepo,
			CloneUrl:       r.HTTPURLToRepo,
		}
		if r.Namespace.Path != userObj.Username {
			repo.IsOrg = true
		}
		newRepos = append(newRepos, repo)
	}

	var branchMap = make(map[string][]*models.CodeRepositoryBranch)
	for _, r := range newRepos {
		var branches []*models.CodeRepositoryBranch
		page = 1
		for {
			bs, _, err := client.Branches.ListBranches(r.RepositoryID, &gitlab.ListBranchesOptions{ListOptions: gitlab.ListOptions{Page: page, PerPage: perPage}})
			if err != nil {
				slog.Error("list branches failed", "err", err, "owner", r.Owner, "repo", r.Name)
				return fmt.Errorf("list branches for repo(%s/%s) failed: %v", r.Owner, r.Name, err)
			}
			var bsObj = make([]*models.CodeRepositoryBranch, 0, len(bs))
			for _, b := range bs {
				bsObj = append(bsObj, &models.CodeRepositoryBranch{
					ID:   uuid.NewV7String(),
					Name: b.Name,
				})
			}
			branches = append(branches, bsObj...)
			if len(bs) < perPage {
				break
			}
			page++
		}
		branchMap[r.RepositoryID] = branches
	}

	return cr.diff(ctx, user3rdPartyObj, newRepos, branchMap)
}
