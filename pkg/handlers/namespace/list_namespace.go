package namespace

import (
	"github.com/labstack/echo/v4"
	"github.com/ximager/ximager/pkg/consts"
	"github.com/ximager/ximager/pkg/services/namespaces"
)

type ListNamespaceResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func (h *handlers) ListNamespace(c echo.Context) error {
	namespaceService := namespaces.NewNamespaceService()
	namespaces, err := namespaceService.ListNamespace(c.Request().Context())
	if err != nil {
		return err
	}

	var resp []ListNamespaceResponse
	for _, ns := range namespaces {
		resp = append(resp, ListNamespaceResponse{
			ID:          ns.ID,
			Name:        ns.Name,
			Description: "eyJpc3MiOiJhdXRoLmRvY2tlci5jb20iLCJzdWIiOiJqbGhhd24iLCJhdWQiOiJyZWdpc3RyeS5kb2NrZXIuY29tIiwiZXhwIjoxNDE1Mzg3MzE1LCJuYmYiOjE0MTUzODcwMTUsImlhdCI6MTQxNTM4NzAxNSwianRpIjoidFlKQ08xYzZjbnl5N2tBbjBjN3JLUGdiVjFIMWJGd3MiLCJhY2Nlc3MiOlt7InR5cGUiOiJyZXBvc2l0b3J5IiwibmFtZSI6InNhbWFsYmEvbXktYXBwIiwiYWN0aW9ucyI6WyJwdXNoIl19XX0",
			CreatedAt:   ns.CreatedAt.Format(consts.DefaultTimePattern),
			UpdatedAt:   ns.UpdatedAt.Format(consts.DefaultTimePattern),
		})
	}

	return c.JSON(200, resp)
}
