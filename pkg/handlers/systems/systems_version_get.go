package systems

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/version"
)

// GetEndpoint handles the get version request
//
//	@Summary	Get version
//	@Tags		System
//	@Accept		json
//	@Produce	json
//	@Router		/systems/version [get]
//	@Success	200	{object}	types.GetSystemVersionResponse
func (h *handlers) GetVersion(c echo.Context) error {
	return c.JSON(http.StatusOK, types.GetSystemVersionResponse{
		Version:   version.Version,
		GitHash:   version.GitHash,
		BuildDate: version.BuildDate,
	})
}
