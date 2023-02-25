// The MIT License (MIT)
//
// Copyright © 2023 Tosone <i@tosone.cn>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

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
