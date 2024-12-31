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
	"strconv"

	"github.com/rs/zerolog/log"
	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

func (cr codeRepository) gitlab(ctx context.Context, user3rdPartyObj *models.User3rdParty) error {
	client, err := gitlab.NewOAuthClient(ptr.To(user3rdPartyObj.Token))
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
			Owned:       ptr.Of(true),
			ListOptions: gitlab.ListOptions{Page: page, PerPage: perPage}})
		if err != nil {
			log.Error().Err(err).Msg("List projects from gitlab failed")
			return fmt.Errorf("List projects from gitlab failed: %w", err)
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
		gs, _, err := client.Groups.ListGroups(&gitlab.ListGroupsOptions{
			AllAvailable:   ptr.Of(true),
			MinAccessLevel: ptr.Of(gitlab.ReporterPermissions),
			ListOptions:    gitlab.ListOptions{Page: page, PerPage: perPage}})
		if err != nil {
			log.Error().Err(err).Msg("List groups from gitlab failed")
			return fmt.Errorf("List groups from gitlab failed: %w", err)
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
			rs, _, err := client.Groups.ListGroupProjects(g.ID, &gitlab.ListGroupProjectsOptions{
				MinAccessLevel: ptr.Of(gitlab.ReporterPermissions),
				ListOptions:    gitlab.ListOptions{Page: page, PerPage: perPage}})
			if err != nil {
				log.Error().Err(err).Msg("List projects from gitlab failed")
				return fmt.Errorf("List projects from gitlab failed: %w", err)
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
				log.Error().Err(err).Str("owner", r.Owner).Str("repo", r.Name).Msg("List branches failed")
				return fmt.Errorf("List branches for repo(%s/%s) failed: %v", r.Owner, r.Name, err)
			}
			var bsObj = make([]*models.CodeRepositoryBranch, 0, len(bs))
			for _, b := range bs {
				bsObj = append(bsObj, &models.CodeRepositoryBranch{
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
