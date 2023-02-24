package artifact

import (
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/ximager/ximager/pkg/consts"
	"github.com/ximager/ximager/pkg/dal/dao"
	"github.com/ximager/ximager/pkg/types"
	"github.com/ximager/ximager/pkg/xerrors"
)

// GetArtifact handles the get artifact request
func (h *handlers) GetArtifact(c echo.Context) error {
	ctx := c.Request().Context()

	var req types.GetArtifactRequest
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
	tag, err := artifactService.GetByDigest(ctx, req.Repository, req.Digest)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Error().Err(err).Msg("Artifact not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, err.Error())
		}
		log.Error().Err(err).Msg("Get artifact failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	return c.JSON(200, types.ArtifactItem{
		ID:        tag.ID,
		Digest:    tag.Digest,
		Size:      tag.Size,
		CreatedAt: tag.CreatedAt.Format(consts.DefaultTimePattern),
		UpdatedAt: tag.UpdatedAt.Format(consts.DefaultTimePattern),
	})
}
