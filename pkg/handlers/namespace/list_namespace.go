package namespace

import (
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/ximager/ximager/pkg/consts"
	"github.com/ximager/ximager/pkg/services/namespaces"
	"github.com/ximager/ximager/pkg/types"
	"github.com/ximager/ximager/pkg/xerrors"
)

type ListNamespaceResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func (h *handlers) ListNamespace(c echo.Context) error {
	ctx := c.Request().Context()

	namespaceService := namespaces.NewNamespaceService()
	namespaces, err := namespaceService.ListNamespace(ctx)
	if err != nil {
		log.Error().Err(err).Msg("List namespace from db failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	var resp []any
	for _, ns := range namespaces {
		resp = append(resp, types.Namespace{
			ID:          ns.ID,
			Name:        ns.Name,
			Description: ns.Description,
			CreatedAt:   ns.CreatedAt.Format(consts.DefaultTimePattern),
			UpdatedAt:   ns.UpdatedAt.Format(consts.DefaultTimePattern),
		})
	}

	return c.JSON(200, types.CommonList{Total: 1, Items: resp})
}
