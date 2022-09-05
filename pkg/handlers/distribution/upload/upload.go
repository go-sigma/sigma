package upload

import "github.com/labstack/echo/v4"

// Handlers is the interface for the distribution blob handlers
type Handlers interface {
	DeleteUpload(ctx echo.Context) error
	GetUpload(ctx echo.Context) error
	PatchUpload(ctx echo.Context) error
	PostUpload(ctx echo.Context) error
	PutUpload(ctx echo.Context) error
}

var _ Handlers = &handler{}

type handler struct{}

// New creates a new instance of the distribution blob handlers
func New() Handlers {
	return &handler{}
}
