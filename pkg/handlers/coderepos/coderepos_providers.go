package coderepos

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// Providers list providers
// @Summary List code repository providers
// @security BasicAuth
// @Tags CodeRepository
// @Accept json
// @Produce json
// @Router /coderepos/providers/ [get]
// @Success 200	{object} types.CommonList{items=[]types.ListCodeRepositoryProvidersResponse}
// @Failure 401 {object} xerrors.ErrCode
// @Failure 500 {object} xerrors.ErrCode
func (h *handlers) Providers(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	iuser := c.Get(consts.ContextUser)
	if iuser == nil {
		log.Error().Msg("Get user from header failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized)
	}
	user, ok := iuser.(*models.User)
	if !ok {
		log.Error().Msg("Convert user from header failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized)
	}

	userService := h.userServiceFactory.New()
	user3rdPartyObjs, err := userService.ListUser3rdParty(ctx, user.ID)
	if err != nil {
		log.Error().Err(err).Msg("List providers failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("List providers failed: %v", err))
	}
	resp := make([]any, 0, len(user3rdPartyObjs))
	for _, user3rdPartyObj := range user3rdPartyObjs {
		resp = append(resp, types.ListCodeRepositoryProvidersResponse{
			Provider: user3rdPartyObj.Provider,
		})
	}
	return c.JSON(http.StatusOK, types.CommonList{Total: int64(len(user3rdPartyObjs)), Items: resp})
}
