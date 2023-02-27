package user

import "github.com/labstack/echo/v4"

// Handlers is the interface for the tag handlers
type Handlers interface {
	// Login handles the login request
	Login(c echo.Context) error
}

var _ Handlers = &handlers{}

type handlers struct{}

// New creates a new instance of the distribution handlers
func New() Handlers {
	return &handlers{}
}
