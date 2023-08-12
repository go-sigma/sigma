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

	"code.gitea.io/sdk/gitea"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

func (cr codeRepository) gitea(ctx context.Context, user3rdPartyObj *models.User3rdParty) error {
	client, err := gitea.NewClient("", gitea.SetToken(ptr.To(user3rdPartyObj.Token)), gitea.SetContext(ctx))
	if err != nil {
		return err
	}

	var repos []*gitea.Repository

	page := 1
	for {
		rs, _, err := client.ListMyRepos(gitea.ListReposOptions{ListOptions: gitea.ListOptions{Page: page, PageSize: perPage}})
		if err != nil {
			log.Error().Err(err).Msg("List repositories failed")
			return fmt.Errorf("List repositories failed: %v", err)
		}
		repos = append(repos, rs...)
		if len(rs) < perPage {
			break
		}
		page++
	}

	var orgs []*gitea.Organization

	page = 1
	for {
		os, _, err := client.ListMyOrgs(gitea.ListOrgsOptions{ListOptions: gitea.ListOptions{Page: page, PageSize: perPage}})
		if err != nil {
			log.Error().Err(err).Msg("List organizations failed")
			return fmt.Errorf("List organizations failed: %v", err)
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
			rs, _, err := client.ListOrgRepos(o.UserName, gitea.ListOrgReposOptions{ListOptions: gitea.ListOptions{Page: page, PageSize: perPage}})
			if err != nil {
				log.Error().Err(err).Msg("List repositories failed")
				return fmt.Errorf("List repositories failed: %v", err)
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
			Owner:          ptr.To(r.Owner).UserName,
			Name:           r.Name,
			SshUrl:         r.SSHURL,
			CloneUrl:       r.CloneURL,
		})
	}

	return cr.diff(ctx, user3rdPartyObj, newRepos)
}
