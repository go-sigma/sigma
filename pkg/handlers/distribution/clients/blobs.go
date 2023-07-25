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
	"net/url"
	"path"
	"strconv"

	"github.com/distribution/distribution/v3"
	"github.com/labstack/echo/v4"
	"github.com/opencontainers/go-digest"

	"github.com/go-sigma/sigma/pkg/consts"
)

// GetBlob ...
func (c *clients) GetBlob(ctx context.Context, repository string, digest digest.Digest) (distribution.Descriptor, io.ReadCloser, error) {
	statusCode, header, reader, err := c.DoRequest(ctx, http.MethodGet, path.Join("/v2/", repository, "blobs", digest.String()), nil)
	if err != nil {
		return distribution.Descriptor{}, nil, err
	}
	if statusCode != 200 {
		return distribution.Descriptor{}, nil, fmt.Errorf("response status code: %d", statusCode)
	}
	descriptor, err := c.parseBlobRespHeader(header)
	if err != nil {
		return distribution.Descriptor{}, nil, err
	}
	return descriptor, reader, nil
}

// HeadBlob ...
func (c *clients) HeadBlob(ctx context.Context, repository string, digest digest.Digest) (distribution.Descriptor, error) {
	statusCode, header, _, err := c.DoRequest(ctx, http.MethodHead, path.Join("/v2/", repository, "blobs", digest.String()), nil)
	if err != nil {
		return distribution.Descriptor{}, err
	}
	if statusCode != 200 {
		return distribution.Descriptor{}, fmt.Errorf("response status code: %d", statusCode)
	}
	descriptor, err := c.parseBlobRespHeader(header)
	if err != nil {
		return distribution.Descriptor{}, err
	}
	return descriptor, nil
}

func (c *clients) parseBlobRespHeader(header http.Header) (distribution.Descriptor, error) {
	size, err := strconv.ParseInt(header.Get(echo.HeaderContentLength), 10, 64)
	if err != nil {
		return distribution.Descriptor{}, fmt.Errorf("blob content length (%s) is invalid", header.Get(echo.HeaderContentLength))
	}
	digest, err := digest.Parse(header.Get(consts.ContentDigest))
	if err != nil {
		return distribution.Descriptor{}, fmt.Errorf("blob content digest (%s) is invalid", header.Get(consts.ContentDigest))
	}
	return distribution.Descriptor{
		MediaType: header.Get(echo.HeaderContentType),
		Size:      size,
		Digest:    digest,
	}, nil
}

func (c *clients) initUpload(ctx context.Context, repository string) (*url.URL, error) {
	var header = http.Header{}
	header.Add(echo.HeaderContentType, echo.MIMEOctetStream)
	statusCode, respHeader, _, err := c.DoRequest(ctx, http.MethodPost, path.Join("/v2", repository, "/blobs/uploads/"), header)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("response status code: %d", statusCode)
	}
	location := respHeader.Get("Location")
	locationURL, err := url.Parse(location)
	if err != nil {
		return nil, fmt.Errorf("init upload parse response header location(%s) failed: %v", location, err)
	}
	return locationURL, nil
}

// PutBlob ...
func (c *clients) PutBlob(ctx context.Context, repository string, digest digest.Digest, content io.Reader) error {
	location, err := c.initUpload(ctx, repository)
	if err != nil {
		return err
	}
	q := location.Query()
	q.Set("digest", digest.String())
	location.RawQuery = q.Encode()

	var header = http.Header{}
	header.Add(echo.HeaderContentType, echo.MIMEOctetStream)
	statusCode, _, _, err := c.DoRequest(ctx, http.MethodPut, location.String(), header, content)
	if err != nil {
		return err
	}
	if statusCode != http.StatusCreated {
		return fmt.Errorf("response status code: %d", statusCode)
	}
	return nil
}
