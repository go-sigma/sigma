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
	"strings"

	"github.com/google/go-github/v53/github"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

const (
	perPage = 100
)

func (cr codeRepository) github(ctx context.Context, user3rdPartyObj *models.User3rdParty) error {
	client := github.NewTokenClient(ctx, ptr.To(user3rdPartyObj.Token))

	userObj, _, err := client.Users.Get(ctx, "")
	if err != nil {
		log.Error().Err(err).Msg("Get user info failed")
		return fmt.Errorf("Get user info failed: %v", err)
	}

	var repos []*github.Repository

	page := 1
	for {
		rs, _, err := client.Repositories.List(ctx, "", &github.RepositoryListOptions{ListOptions: github.ListOptions{Page: page, PerPage: perPage}})
		if err != nil {
			log.Error().Err(err).Msg("List repositories failed")
			return fmt.Errorf("List repositories failed: %v", err)
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
			rs, _, err := client.Repositories.ListByOrg(ctx, ptr.To(o.Login),
				&github.RepositoryListByOrgOptions{ListOptions: github.ListOptions{Page: page, PerPage: perPage}})
			if err != nil {
				log.Error().Err(err).Msg("List repositories for orgs failed")
				return fmt.Errorf("List repositories for orgs failed: %v", err)
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
			Owner:          ptr.To(ptr.To(r.Owner).Login),
			Name:           r.GetName(),
			SshUrl:         r.GetSSHURL(),
			CloneUrl:       r.GetCloneURL(),
		})
	}

	return cr.diff(ctx, user3rdPartyObj, newRepos)
}
