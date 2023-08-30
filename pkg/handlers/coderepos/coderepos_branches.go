package coderepos

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// ListBranches list all of the branches
// @Summary List code repository branches
// @security BasicAuth
// @Tags CodeRepository
// @Accept json
// @Produce json
// @Router /coderepos/{id}/branches [get]
// @Param id path string true "code repository id"
// @Success 200	{object} types.CommonList{items=[]types.CodeRepositoryBranchItem}
// @Failure 500 {object} xerrors.ErrCode
func (h *handlers) ListBranches(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.ListCodeRepositoryBranchesRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}

	codeRepositoryService := h.codeRepositoryServiceFactory.New()
	branchObjs, total, err := codeRepositoryService.ListBranchesWithoutPagination(ctx, req.ID)
	if err != nil {
		log.Error().Err(err).Msg("List branches failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("List branches failed: %v", err))
	}
	resp := make([]any, 0, len(branchObjs))
	for _, branchObj := range branchObjs {
		resp = append(resp, types.CodeRepositoryBranchItem{
			ID:        branchObj.ID,
			Name:      branchObj.Name,
			CreatedAt: branchObj.CreatedAt.Format(consts.DefaultTimePattern),
			UpdatedAt: branchObj.UpdatedAt.Format(consts.DefaultTimePattern),
		})
	}

	return c.JSON(http.StatusOK, types.CommonList{Total: total, Items: resp})
}
