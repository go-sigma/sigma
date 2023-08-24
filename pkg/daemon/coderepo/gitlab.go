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
	"github.com/xanzy/go-gitlab"

	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

func (cr codeRepository) gitlab(ctx context.Context, user3rdPartyObj *models.User3rdParty) error {
	client, err := gitlab.NewClient(ptr.To(user3rdPartyObj.Token))
	if err != nil {
		return err
	}

	var repos []*gitlab.Project

	page := 1
	for {
		rs, _, err := client.Projects.ListProjects(&gitlab.ListProjectsOptions{ListOptions: gitlab.ListOptions{Page: page, PerPage: perPage}})
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

	var groups []*gitlab.Group

	page = 1
	for {
		gs, _, err := client.Groups.ListGroups(&gitlab.ListGroupsOptions{ListOptions: gitlab.ListOptions{Page: page, PerPage: perPage}})
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
			rs, _, err := client.Groups.ListGroupProjects(g.ID, &gitlab.ListGroupProjectsOptions{ListOptions: gitlab.ListOptions{Page: page, PerPage: perPage}})
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
		newRepos = append(newRepos, &models.CodeRepository{
			User3rdPartyID: user3rdPartyObj.ID,
			RepositoryID:   strconv.Itoa(r.ID),
			Owner:          ptr.To(r.Owner).Name,
			Name:           r.Name,
			SshUrl:         r.SSHURLToRepo,
			CloneUrl:       r.HTTPURLToRepo,
		})
	}

	return cr.diff(ctx, user3rdPartyObj, newRepos)
}
