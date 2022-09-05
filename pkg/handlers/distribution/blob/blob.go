package blob

import (
	"fmt"
	"regexp"

	"github.com/distribution/distribution/v3/reference"
	"github.com/labstack/echo/v4"
	"github.com/opencontainers/go-digest"
)

// Handlers is the interface for the distribution blob handlers
type Handlers interface {
	DeleteBlob(ctx echo.Context) error
	HeadBlob(ctx echo.Context) error
	GetBlob(ctx echo.Context) error
}

var _ Handlers = &handler{}

var blobRouteReg = regexp.MustCompile(fmt.Sprintf(`^/v2/%s/blobs/%s$`, reference.NameRegexp.String(), digest.DigestRegexp.String()))

type handler struct{}

// New creates a new instance of the distribution blob handlers
func New() Handlers {
	return &handler{}
}
