package namespace

import "github.com/labstack/echo/v4"

// Handlers is the interface for the namespace handlers
type Handlers interface {
	// PostNamespace handles the post namespace request
	PostNamespace(c echo.Context) error
	// ListNamespace handles the list namespace request
	ListNamespace(c echo.Context) error
	// DeleteNamespace handles the delete namespace request
	DeleteNamespace(c echo.Context) error
}

var _ Handlers = &handlers{}

type handlers struct{}

// New creates a new instance of the distribution handlers
func New() Handlers {
	return &handlers{}
}
