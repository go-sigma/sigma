package artifact

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/ximager/ximager/pkg/consts"
	"github.com/ximager/ximager/pkg/dal/dao"
	"github.com/ximager/ximager/pkg/types"
	"github.com/ximager/ximager/pkg/xerrors"
)

// ListArtifact handles the list artifact request
func (h *handlers) ListArtifact(c echo.Context) error {
	ctx := c.Request().Context()

	var req types.ListArtifactRequest
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

	artifactService := dao.NewArtifactService()
	artifacts, err := artifactService.ListArtifact(ctx, req)

	var resp []any
	for _, artifact := range artifacts {
		resp = append(resp, types.ArtifactItem{
			ID:        artifact.ID,
			Digest:    artifact.Digest,
			Size:      artifact.Size,
			CreatedAt: artifact.CreatedAt.Format(consts.DefaultTimePattern),
			UpdatedAt: artifact.UpdatedAt.Format(consts.DefaultTimePattern),
		})
	}
	total, err := artifactService.CountArtifact(ctx, req)
	if err != nil {
		log.Error().Err(err).Msg("Count artifact from db failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	return c.JSON(http.StatusOK, types.CommonList{Total: total, Items: resp})
}
