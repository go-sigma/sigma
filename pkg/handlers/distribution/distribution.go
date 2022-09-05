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
