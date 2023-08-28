package oauth2

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// RedirectCallback ...
func (h *handlers) RedirectCallback(c echo.Context) error {
	var req types.Oauth2CallbackRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}
	return c.Redirect(http.StatusMovedPermanently, fmt.Sprintf("%s/?code=%s#/login/callback/%s", req.Endpoint, req.Code, req.Provider))
}
