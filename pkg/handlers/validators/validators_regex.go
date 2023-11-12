package validators

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// ValidateRegexp handles the validate regexp request
//
//	@Summary	Validate regexp
//	@Tags		Validator
//	@security	BasicAuth
//	@Accept		json
//	@Produce	json
//	@Router		/validators/regexp [post]
//	@Param		message	body	types.ValidateCronRequest	true	"Validate regexp object"
//	@Success	204
//	@Failure	400	{object}	xerrors.ErrCode
func (h *handlers) ValidateRegexp(c echo.Context) error {
	var req types.ValidateRegexpRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}

	_, err = regexp.Compile(req.Regex)
	if err != nil {
		log.Error().Err(err).Msg("Parse regex failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Parse regex failed: %v", err))
	}

	return c.NoContent(http.StatusNoContent)
}
