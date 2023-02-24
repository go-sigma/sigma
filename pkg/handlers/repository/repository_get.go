package repository

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/ximager/ximager/pkg/consts"
	"github.com/ximager/ximager/pkg/dal/dao"
	"github.com/ximager/ximager/pkg/types"
	"github.com/ximager/ximager/pkg/xerrors"
)

// GetRepository handles the get repository request
func (h *handlers) GetRepository(c echo.Context) error {
	ctx := c.Request().Context()

	var req types.GetRepositoryRequest
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

	repositoryService := dao.NewRepositoryService()
	repository, err := repositoryService.Get(ctx, req.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Error().Err(err).Msg("Get repository from db failed")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, err.Error())
		}
		log.Error().Err(err).Msg("Get repository from db failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	artifactService := dao.NewArtifactService()
	artifactCountRef, err := artifactService.CountByRepository(ctx, []uint{repository.ID})
	if err != nil {
		log.Error().Err(err).Msg("Count artifact from db failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	return c.JSON(http.StatusOK, types.RepositoryItem{
		ID:            repository.ID,
		Name:          repository.Name,
		ArtifactCount: artifactCountRef[repository.ID],
		CreatedAt:     repository.CreatedAt.Format(consts.DefaultTimePattern),
		UpdatedAt:     repository.UpdatedAt.Format(consts.DefaultTimePattern),
	})
}
