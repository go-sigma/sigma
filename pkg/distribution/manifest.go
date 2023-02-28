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
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/distribution/distribution/v3"
	"github.com/distribution/distribution/v3/reference"
	rapiv2 "github.com/distribution/distribution/v3/registry/api/v2"
	"github.com/opencontainers/go-digest"
)

// ManifestService describes operations on image manifests.
type ManifestService interface {
	// Exists returns true if the manifest exists.
	// reference can be a tag or a digest.
	Exists(ctx context.Context, ref Reference) (bool, error)

	// Get retrieves the manifest specified by the given reference.
	// reference can be a tag or a digest.
	Get(ctx context.Context, ref Reference) (distribution.Manifest, *distribution.Descriptor, error)

	// Put creates or updates the given manifest returning the manifest digest
	// reference can be a tag or a digest.
	Put(ctx context.Context, ref Reference, manifest distribution.Manifest) (digest.Digest, error)

	// Delete removes the manifest specified by the given digest. Deleting
	// a manifest that doesn't exist will return ErrManifestNotFound
	// reference can be a tag or a digest.
	Delete(ctx context.Context, ref Reference) error
}

var _ ManifestService = &manifest{}

type manifest struct {
	client *http.Client
	ub     *rapiv2.URLBuilder
	name   reference.Named
}

// Reference is a reference to a manifest.
type Reference struct {
	repo   reference.Named
	tag    string
	digest digest.Digest
}

// Sanitize returns a reference that can be used to access the manifest.
func (r *Reference) Sanitize() (reference.Named, error) {
	if r.tag == "" && r.digest == "" {
		return nil, fmt.Errorf("tag or digest is required")
	}
	if r.tag != "" {
		return reference.WithTag(r.repo, r.tag)
	} else {
		return reference.WithDigest(r.repo, r.digest)
	}
}

// Exists returns true if the manifest exists.
// reference can be a tag or a digest.
func (m *manifest) Exists(ctx context.Context, ref Reference) (bool, error) {
	r, err := ref.Sanitize()
	if err != nil {
		return false, err
	}
	u, err := m.ub.BuildManifestURL(r)
	if err != nil {
		return false, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, u, nil)
	if err != nil {
		return false, err
	}
	resp, err := m.client.Do(req)
	if err != nil {
		return false, err
	}

	if SuccessStatus(resp.StatusCode) {
		return true, nil
	} else if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
}

func (m *manifest) Get(ctx context.Context, ref Reference) (distribution.Manifest, *distribution.Descriptor, error) {
	r, err := ref.Sanitize()
	if err != nil {
		return nil, nil, err
	}
	u, err := m.ub.BuildManifestURL(r)
	if err != nil {
		return nil, nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}
	for _, t := range distribution.ManifestMediaTypes() {
		req.Header.Add(http.CanonicalHeaderKey("Accept"), t)
	}
	resp, err := m.client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	mt := resp.Header.Get(http.CanonicalHeaderKey("Content-Type"))
	manifest, desc, err := distribution.UnmarshalManifest(mt, body)
	if err != nil {
		return nil, nil, err
	}
	return manifest, &desc, nil
}

func (m *manifest) Put(ctx context.Context, ref Reference, manifest distribution.Manifest) (digest.Digest, error) {
	r, err := ref.Sanitize()
	if err != nil {
		return "", err
	}
	u, err := m.ub.BuildManifestURL(r)
	if err != nil {
		return "", err
	}
	mt, payload, err := manifest.Payload()
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", mt)

	resp, err := m.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if SuccessStatus(resp.StatusCode) {
		dgstHeader := resp.Header.Get("Docker-Content-Digest")
		dgst, err := digest.Parse(dgstHeader)
		if err != nil {
			return "", err
		}
		return dgst, nil
	}
	return "", nil
}

func (m *manifest) Delete(ctx context.Context, ref Reference) error {
	r, err := ref.Sanitize()
	if err != nil {
		return err
	}
	u, err := m.ub.BuildManifestURL(r)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return err
	}

	resp, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if SuccessStatus(resp.StatusCode) {
		return nil
	}
	return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
}
