package namespace

import (
	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/ximager/ximager/pkg/dal/models"
	"github.com/ximager/ximager/pkg/services/namespaces"
)

type CreateNamespaceRequest struct {
	Name string `json:"name" validate:"required"`
}

// PostNamespace handles the post namespace request
func (h *handlers) PostNamespace(c echo.Context) error {
	var req CreateNamespaceRequest
	err := c.Bind(&req)
	if err != nil {
		return err
	}
	log.Info().Interface("req", req).Msg("PostNamespace")
	vr := validator.New()
	err = vr.Struct(&req)
	if err != nil {
		return err
	}
	namespaceService := namespaces.NewNamespaceService()
	_, err = namespaceService.Create(c.Request().Context(), &models.Namespace{Name: req.Name})
	if err != nil {
		return err
	}
	return nil
}
