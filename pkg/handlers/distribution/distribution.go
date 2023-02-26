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
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/ximager/ximager/pkg/consts"
	"github.com/ximager/ximager/pkg/handlers/distribution/blob"
	"github.com/ximager/ximager/pkg/handlers/distribution/manifest"
	"github.com/ximager/ximager/pkg/handlers/distribution/upload"
)

// Handlers is the interface for the distribution handlers
type Handlers interface {
	// GetHealthy handles the get healthy request
	GetHealthy(ctx echo.Context) error
	// ListTags handles the list tags request
	ListTags(ctx echo.Context) error
	// ListRepositories handles the list repositories request
	ListRepositories(ctx echo.Context) error
}

var _ Handlers = &handlers{}

type handlers struct{}

// New creates a new instance of the distribution handlers
func New() Handlers {
	return &handlers{}
}

// NewBlob creates a new instance of the distribution blob handlers
func NewBlob() blob.Handlers {
	return blob.New()
}

// NewUpload creates a new instance of the distribution upload handlers
func NewUpload() upload.Handlers {
	return upload.New()
}

// NewManifest creates a new instance of the distribution manifest handlers
func NewManifest() manifest.Handlers {
	return manifest.New()
}

// All handles the all request
func All(c echo.Context) error {
	c.Response().Header().Set(consts.APIVersionKey, consts.APIVersionValue)

	method := c.Request().Method
	uri := c.Request().RequestURI

	baseHandler := New()
	if method == http.MethodGet && uri == "/v2/" {
		return baseHandler.GetHealthy(c)
	}
	if method == http.MethodGet && uri == "/v2/_catalog" {
		return baseHandler.ListRepositories(c)
	}
	if method == http.MethodGet && strings.HasSuffix(uri, "/tags/list") {
		return baseHandler.ListTags(c)
	}

	uploadHandler := NewUpload()
	if method == http.MethodPost && strings.HasSuffix(uri, "blobs/uploads/") {
		return uploadHandler.PostUpload(c)
	}

	urix := uri[:strings.LastIndex(uri, "/")]

	if strings.HasSuffix(urix, "/blobs/uploads") {
		if method == http.MethodGet {
			return uploadHandler.GetUpload(c)
		} else if method == http.MethodPatch {
			return uploadHandler.PatchUpload(c)
		} else if method == http.MethodPut {
			return uploadHandler.PutUpload(c)
		} else if method == http.MethodDelete {
			return uploadHandler.DeleteUpload(c)
		} else {
			return c.String(405, "Method Not Allowed")
		}
	}

	blobHandler := NewBlob()
	if strings.HasSuffix(urix, "/blobs") {
		if method == http.MethodGet {
			return blobHandler.GetBlob(c)
		} else if method == http.MethodHead {
			return blobHandler.HeadBlob(c)
		} else {
			return c.String(405, "Method Not Allowed")
		}
	}

	manifestHandler := NewManifest()
	if strings.HasSuffix(urix, "/manifests") {
		if method == http.MethodGet {
			return manifestHandler.GetManifest(c)
		} else if method == http.MethodHead {
			return manifestHandler.HeadManifest(c)
		} else if method == http.MethodPut {
			return manifestHandler.PutManifest(c)
		} else if method == http.MethodDelete {
			return manifestHandler.DeleteManifest(c)
		} else {
			return c.String(405, "Method Not Allowed")
		}
	}

	return c.String(200, "OK: "+method+" "+uri)
}
