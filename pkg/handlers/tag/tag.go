package tag

import "github.com/labstack/echo/v4"

// Handlers is the interface for the tag handlers
type Handlers interface {
	// ListTag handles the list tag request
	ListTag(c echo.Context) error
	// GetTag handles the get tag request
	GetTag(c echo.Context) error
	// DeleteTag handles the delete tag request
	DeleteTag(c echo.Context) error
}

var _ Handlers = &handlers{}

type handlers struct{}

// New creates a new instance of the distribution handlers
func New() Handlers {
	return &handlers{}
}
