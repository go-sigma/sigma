package manifest

import "github.com/labstack/echo/v4"

// Handlers is the interface for the distribution manifest handlers
type Handlers interface {
	GetManifest(ctx echo.Context) error
	HeadManifest(ctx echo.Context) error
	PutManifest(ctx echo.Context) error
	DeleteManifest(ctx echo.Context) error
}

var _ Handlers = &handler{}

type handler struct{}

// New creates a new instance of the distribution manifest handlers
func New() Handlers {
	return &handler{}
}
