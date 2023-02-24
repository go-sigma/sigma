package repository

import "github.com/labstack/echo/v4"

// Handlers is the interface for the repository handlers
type Handlers interface {
	// ListNamespace handles the list repository request
	ListRepository(c echo.Context) error
	// GetRepository handles the get repository request
	GetRepository(c echo.Context) error
	// DeleteRepository handles the delete repository request
	DeleteRepository(c echo.Context) error
}

var _ Handlers = &handlers{}

type handlers struct{}

// New creates a new instance of the distribution handlers
func New() Handlers {
	return &handlers{}
}
