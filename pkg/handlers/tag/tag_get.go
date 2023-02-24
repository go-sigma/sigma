package tag

import (
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/ximager/ximager/pkg/consts"
	"github.com/ximager/ximager/pkg/dal/dao"
	"github.com/ximager/ximager/pkg/types"
	"github.com/ximager/ximager/pkg/xerrors"
)

// GetTag handles the get tag request
func (h *handlers) GetTag(c echo.Context) error {
	ctx := c.Request().Context()

	var req types.GetTagRequest
	err := c.Bind(&req)
	if err != nil {
		log.Error().Err(err).Msg("Bind request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}
	err = c.Validate(&req)
	if err != nil {
		log.Error().Err(err).Msg("Validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}

	tagService := dao.NewTagService()
	tag, err := tagService.GetByID(ctx, req.ID)

	return c.JSON(200, types.TagItem{
		ID:        tag.ID,
		Name:      tag.Name,
		Digest:    tag.Digest,
		Size:      tag.Size,
		CreatedAt: tag.CreatedAt.Format(consts.DefaultTimePattern),
		UpdatedAt: tag.UpdatedAt.Format(consts.DefaultTimePattern),
	})
}
