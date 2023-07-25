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

package clients

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path"

	"github.com/distribution/distribution/v3"
	"github.com/labstack/echo/v4"
)

// GetManifest ...
func (c *clients) GetManifest(ctx context.Context, repository, reference string) (distribution.Manifest, distribution.Descriptor, error) {
	var header = http.Header{}
	header.Add(echo.HeaderAccept, "application/vnd.docker.distribution.manifest.v2+json")
	header.Add(echo.HeaderAccept, "application/vnd.docker.distribution.manifest.list.v2+json")
	header.Add(echo.HeaderAccept, "application/vnd.oci.image.index.v1+json")
	header.Add(echo.HeaderAccept, "application/vnd.oci.image.manifest.v1+json")
	statusCode, header, reader, err := c.DoRequest(ctx, http.MethodGet, path.Join("/v2/", repository, "manifests", reference), nil)
	if err != nil {
		return nil, distribution.Descriptor{}, err
	}
	if statusCode != 200 {
		return nil, distribution.Descriptor{}, fmt.Errorf("response status code: %d", statusCode)
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, distribution.Descriptor{}, fmt.Errorf("get manifest body failed: %v", err)
	}
	contentType := header.Get(echo.HeaderContentType)
	manifest, descriptor, err := distribution.UnmarshalManifest(contentType, data)
	if err != nil {
		return nil, distribution.Descriptor{}, err
	}
	return manifest, descriptor, nil
}

// HeadManifest ...
func (c *clients) HeadManifest(ctx context.Context, repository, reference string) (bool, error) {
	var header = http.Header{}
	header.Add(echo.HeaderAccept, "application/vnd.docker.distribution.manifest.v2+json")
	header.Add(echo.HeaderAccept, "application/vnd.docker.distribution.manifest.list.v2+json")
	header.Add(echo.HeaderAccept, "application/vnd.oci.image.index.v1+json")
	header.Add(echo.HeaderAccept, "application/vnd.oci.image.manifest.v1+json")
	statusCode, _, _, err := c.DoRequest(ctx, http.MethodHead, path.Join("/v2/", repository, "manifests", reference), nil)
	if err != nil {
		return false, err
	}
	if statusCode == http.StatusNotFound {
		return false, nil
	}
	if statusCode != 200 {
		return false, fmt.Errorf("response status code: %d", statusCode)
	}
	return true, nil
}
