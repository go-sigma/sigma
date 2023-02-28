// Copyright 2023 XImager
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

package distribution

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"time"

	rapiv2 "github.com/distribution/distribution/v3/registry/api/v2"
	dtspecv1 "github.com/opencontainers/distribution-spec/specs-go/v1"
)

const (
	paginationLimit = 1000
)

// Registry provides an interface for calling Repositories, which returns a catalog of repositories.
type Registry interface {
	Repositories(ctx context.Context) ([]string, error)
}

type registry struct {
	client *http.Client
	ub     *rapiv2.URLBuilder
}

// NewRegistry creates a new instance of the registry client
func NewRegistry(baseURL string, transport http.RoundTripper) (Registry, error) {
	ub, err := rapiv2.NewURLBuilderFromString(baseURL, false)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   1 * time.Minute,
	}

	return &registry{
		client: client,
		ub:     ub,
	}, nil
}

// Repositories returns the list of repositories
func (r *registry) Repositories(ctx context.Context) ([]string, error) {
	var result []string

	values := url.Values{}
	values.Set("n", strconv.Itoa(paginationLimit))

	for {
		u, err := r.ub.BuildCatalogURL(values)
		if err != nil {
			return nil, err
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			return nil, err
		}
		resp, err := r.client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, err
		}
		var repositories dtspecv1.RepositoryList
		err = json.NewDecoder(resp.Body).Decode(&repositories)
		if err != nil {
			return nil, err
		}
		if len(repositories.Repositories) == 0 {
			break
		}
		result = append(result, repositories.Repositories...)
		if len(repositories.Repositories) < paginationLimit {
			break
		}
		values.Set("last", repositories.Repositories[len(repositories.Repositories)-1])
	}
	return result, nil
}
