package distribution

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// GetHealthy handles the get healthy request
func (h *handlers) GetHealthy(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}
