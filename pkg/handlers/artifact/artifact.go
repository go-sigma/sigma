package artifact

import "github.com/labstack/echo/v4"

// Handlers is the interface for the artifact handlers
type Handlers interface {
	// ListArtifact handles the list artifact request
	ListArtifact(c echo.Context) error
	// GetArtifact handles the get artifact request
	GetArtifact(c echo.Context) error
	// DeleteArtifact handles the delete artifact request
	DeleteArtifact(c echo.Context) error
}

var _ Handlers = &handlers{}

type handlers struct{}

// New creates a new instance of the distribution handlers
func New() Handlers {
	return &handlers{}
}
