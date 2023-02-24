package tag

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/ximager/ximager/pkg/consts"
	"github.com/ximager/ximager/pkg/dal/dao"
	"github.com/ximager/ximager/pkg/types"
	"github.com/ximager/ximager/pkg/xerrors"
)

// ListTag handles the list tag request
func (h *handlers) ListTag(c echo.Context) error {
	ctx := c.Request().Context()

	var req types.ListTagRequest
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
	tags, err := tagService.ListTag(ctx, req)

	var resp []any
	for _, tag := range tags {
		resp = append(resp, types.TagItem{
			ID:        tag.ID,
			Name:      tag.Name,
			Digest:    tag.Digest,
			Size:      tag.Size,
			CreatedAt: tag.CreatedAt.Format(consts.DefaultTimePattern),
			UpdatedAt: tag.UpdatedAt.Format(consts.DefaultTimePattern),
		})
	}
	total, err := tagService.CountTag(ctx, req)
	if err != nil {
		log.Error().Err(err).Msg("Count tag from db failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	return c.JSON(http.StatusOK, types.CommonList{Total: total, Items: resp})
}
