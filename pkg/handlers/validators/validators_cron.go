package validators

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// ValidateCron handles the validate cron request
//
//	@Summary	Validate cron
//	@Tags		Validator
//	@security	BasicAuth
//	@Accept		json
//	@Produce	json
//	@Router		/validators/cron [post]
//	@Param		message	body	types.ValidateCronRequest	true	"Validate cron object"
//	@Success	204
//	@Failure	400	{object}	xerrors.ErrCode
func (h *handlers) ValidateCron(c echo.Context) error {
	var req types.ValidateCronRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}

	_, err = cron.ParseStandard(req.Cron)
	if err != nil {
		log.Error().Err(err).Msg("Parse cron rule failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Parse cron rule failed: %v", err))
	}

	return c.NoContent(http.StatusNoContent)
}
