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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/distribution/distribution/v3"
	dtspecv1 "github.com/opencontainers/distribution-spec/specs-go/v1"
)

// Client is the interface for the distribution client
type Client interface {
	// distribution.ManifestService
	// distribution.TagService
	// distribution.BlobService
	// distribution.Repository

	// basic

	// Ping checks if the registry is available
	Ping() error
	// Catalog returns the list of repositories
	Catalog(n, last int) ([]string, error)
	// Tags returns the list of tags for the given repository
	Tags(repo string) ([]string, error)

	// manifest

	// Manifest returns the manifest for the given repository and reference.
	HeadManifest(repo, ref string) (bool, error)
	// Manifest returns the manifest for the given repository and reference.
	GetManifest(repo, ref string) (distribution.Manifest, error)
	// PutManifest uploads the manifest for the given repository and reference.
	PutManifest(repo, ref, mediaType string, payload []byte) error
	// DeleteManifest deletes the manifest for the given repository and reference.
	DeleteManifest(repo, ref string) error

	// blob

	// HeadBlob returns the blob for the given repository and digest.
	HeadBlob(repo, digest string) (string, error)
	// GetBlob returns the blob for the given repository and digest.
	GetBlob(repo, digest string) (string, []byte, error)
	// DeleteBlob deletes the blob for the given repository and digest.
	DeleteBlob(repo, digest string) error

	// upload blob

	// InitiateBlobUpload initiate a blob upload and returns a uuid
	InitiateBlobUpload(repo string) (string, error)
	// GetBlobUploadStatus returns the status of the upload
	GetBlobUploadStatus(repo, uuid string) (int64, error)
	// PatchBlobUpload patches the upload
	PatchBlobUpload(repo, uuid string, payload []byte) error
	// PutBlobUpload uploads the blob
	PutBlobUpload(repo, uuid string, payload []byte) error
	// DeleteBlobUpload deletes the upload
	DeleteBlobUpload(repo, uuid string) error
}

var urlBuilder = map[string]string{
	"ping":     "%s/v2/",
	"catalog":  "%s/v2/_catalog?n=" + strconv.Itoa(paginationLimit),
	"tags":     "%s/v2/%s/tags/list",
	"manifest": "%s/v2/%s/manifests/%s",
}

type basicAuth struct {
	username string
	password string
}

type client struct {
	addr string
	auth *basicAuth
}

// NewClient creates a new instance of the distribution client
func NewClient(addr, username, password string) Client {
	return &client{
		addr: addr,
		auth: &basicAuth{
			username: username,
			password: password,
		},
	}
}

func (c *client) requester(method, url string, body ...io.Reader) (*http.Request, error) {
	b := io.Reader(nil)
	if len(body) > 0 {
		b = body[0]
	}
	req, err := http.NewRequest(method, url, b)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.auth.username, c.auth.password)
	return req, nil
}

// Ping checks if the registry is available
func (c *client) Ping() error {
	req, err := c.requester(http.MethodGet, fmt.Sprintf(urlBuilder["ping"], c.addr))
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ping failed: %s", resp.Status)
	}
	return nil
}

// Catalog returns the list of repositories
func (c *client) Catalog(n, last int) ([]string, error) {
	var url = fmt.Sprintf(urlBuilder["catalog"], c.addr)
	var result []string
	for {
		req, err := c.requester(http.MethodGet, url)
		if err != nil {
			return nil, err
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("catalog failed: %s", resp.Status)
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
		url += "&last=" + repositories.Repositories[len(repositories.Repositories)-1]
	}
	return result, nil
}

// Tags returns the list of tags for the given repository
func (c *client) Tags(repo string) ([]string, error) {
	req, err := c.requester(http.MethodGet, fmt.Sprintf(urlBuilder["tags"], c.addr, repo))
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list tags failed: %s", resp.Status)
	}
	var tags dtspecv1.TagList
	err = json.NewDecoder(resp.Body).Decode(&tags)
	if err != nil {
		return nil, err
	}
	return tags.Tags, nil
}

// HeadManifest returns the manifest for the given repository and reference.
func (c *client) HeadManifest(repo, ref string) (bool, error) {
	var url = fmt.Sprintf(urlBuilder["manifest"], c.addr, repo, ref)
	req, err := c.requester(http.MethodHead, url)
	if err != nil {
		return false, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("head manifest failed: %s", resp.Status)
	}

	return true, nil
}

// GetManifest retrieves the manifest specified by the given digest
func (c *client) GetManifest(repo, ref string) (distribution.Manifest, error) {
	var url = fmt.Sprintf(urlBuilder["manifest"], c.addr, repo, ref)
	req, err := c.requester(http.MethodGet, url)
	if err != nil {
		return nil, err
	}
	for _, t := range distribution.ManifestMediaTypes() {
		req.Header.Add(http.CanonicalHeaderKey("Accept"), t)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get manifest failed: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	mt := resp.Header.Get(http.CanonicalHeaderKey("Content-Type"))
	manifest, _, err := distribution.UnmarshalManifest(mt, body)
	if err != nil {
		return nil, err
	}
	return manifest, nil
}

// PutManifest creates or updates the given manifest returning the manifest digest
func (c *client) PutManifest(repo, ref, mediaType string, payload []byte) error {
	return nil
}

// DeleteManifest deletes the manifest for the given repository and reference.
func (c *client) DeleteManifest(repo, ref string) error {
	return nil
}

// HeadBlob returns the blob for the given repository and digest.
func (c *client) HeadBlob(repo, digest string) (string, error) {
	return "", nil
}

// GetBlob returns the blob for the given repository and digest.
func (c *client) GetBlob(repo, digest string) (string, []byte, error) {
	return "", nil, nil
}

// DeleteBlob deletes the blob for the given repository and digest.
func (c *client) DeleteBlob(repo, digest string) error {
	return nil
}

// InitiateBlobUpload initiate a blob upload and returns a uuid
func (c *client) InitiateBlobUpload(repo string) (string, error) {
	return "", nil
}

// GetBlobUploadStatus returns the status of the upload
func (c *client) GetBlobUploadStatus(repo, uuid string) (int64, error) {
	return 0, nil
}

// PatchBlobUpload patches the upload
func (c *client) PatchBlobUpload(repo, uuid string, payload []byte) error {
	return nil
}

// PutBlobUpload uploads the blob
func (c *client) PutBlobUpload(repo, uuid string, payload []byte) error {
	return nil
}

// DeleteBlobUpload deletes the upload
func (c *client) DeleteBlobUpload(repo, uuid string) error {
	return nil
}
