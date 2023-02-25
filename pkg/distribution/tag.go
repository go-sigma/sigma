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
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/distribution/distribution/v3"
	"github.com/distribution/distribution/v3/reference"
	rapiv2 "github.com/distribution/distribution/v3/registry/api/v2"
	dtspecv1 "github.com/opencontainers/distribution-spec/specs-go/v1"
)

// TagService provides access to information about tagged objects.
type TagService interface {
	// Get retrieves the descriptor identified by the tag. Some
	// implementations may differentiate between "trusted" tags and
	// "untrusted" tags. If a tag is "untrusted", the mapping will be returned
	// as an ErrTagUntrusted error, with the target descriptor.
	Get(ctx context.Context, tag string) (distribution.Manifest, *distribution.Descriptor, error)

	// Tag associates the tag with the provided descriptor, updating the
	// current association, if needed.
	Tag(ctx context.Context, tag string, desc distribution.Descriptor) error

	// Untag removes the given tag association
	Untag(ctx context.Context, tag string) error

	// All returns the set of tags managed by this tag service
	All(ctx context.Context) ([]string, error)

	// Manifests returns a reference to this repository's manifest service.
	// with the supplied options applied.
	Manifests() ManifestService
}

var _ TagService = &tag{}

type tag struct {
	client *http.Client
	ub     *rapiv2.URLBuilder
	name   reference.Named
}

func (t *tag) Manifests() ManifestService {
	return &manifest{
		client: t.client,
		ub:     t.ub,
		name:   t.name,
	}
}

func (t *tag) Tag(ctx context.Context, tag string, desc distribution.Descriptor) error {
	return nil
}

func (t *tag) Get(ctx context.Context, tag string) (distribution.Manifest, *distribution.Descriptor, error) {
	return t.Manifests().Get(ctx, Reference{
		repo: t.name,
		tag:  tag,
	})
}

func (t *tag) Untag(ctx context.Context, tag string) error {
	return t.Manifests().Delete(ctx, Reference{
		repo: t.name,
		tag:  tag,
	})
}

func (t *tag) All(ctx context.Context) ([]string, error) {
	var result []string

	values := url.Values{}
	values.Set("n", strconv.Itoa(paginationLimit))

	for {
		u, err := t.ub.BuildTagsURL(t.name, values)
		if err != nil {
			return nil, err
		}
		req, err := http.NewRequest(http.MethodGet, u, nil)
		if err != nil {
			return nil, err
		}
		req = req.WithContext(ctx)
		resp, err := t.client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
		}
		var tagList dtspecv1.TagList
		err = json.NewDecoder(resp.Body).Decode(&tagList)
		if err != nil {
			return nil, err
		}
		if len(tagList.Tags) == 0 {
			break
		}
		result = append(result, tagList.Tags...)
		if len(tagList.Tags) < paginationLimit {
			break
		}
		values.Set("last", tagList.Tags[len(tagList.Tags)-1])
	}
	return result, nil
}
