package routers

import (
	"net/http"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"

	"github.com/ximager/ximager/pkg/handlers/distribution"
	"github.com/ximager/ximager/pkg/handlers/namespace"
	"github.com/ximager/ximager/pkg/middlewares"
	"github.com/ximager/ximager/web"
)

// CustomValidator is a custom validator for echo
type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		// Optionally, you could return the error to give each route more control over the status code
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func Initialize(e *echo.Echo) error {
	web.RegisterHandlers(e)

	e.Validator = &CustomValidator{validator: validator.New()}

	e.GET("/health", func(c echo.Context) error {
		return c.String(200, "OK")
	})

	e.GET("/service/token", func(c echo.Context) error {
		str := `{"token": "eyJ0eXAiOiJKV1QiLCJhbGciOiJFUzI1NiIsImtpZCI6IlBZWU86VEVXVTpWN0pIOjI2SlY6QVFUWjpMSkMzOlNYVko6WEdIQTozNEYyOjJMQVE6WlJNSzpaN1E2In0.eyJpc3MiOiJhdXRoLmRvY2tlci5jb20iLCJzdWIiOiJqbGhhd24iLCJhdWQiOiJyZWdpc3RyeS5kb2NrZXIuY29tIiwiZXhwIjoxNDE1Mzg3MzE1LCJuYmYiOjE0MTUzODcwMTUsImlhdCI6MTQxNTM4NzAxNSwianRpIjoidFlKQ08xYzZjbnl5N2tBbjBjN3JLUGdiVjFIMWJGd3MiLCJhY2Nlc3MiOlt7InR5cGUiOiJyZXBvc2l0b3J5IiwibmFtZSI6InNhbWFsYmEvbXktYXBwIiwiYWN0aW9ucyI6WyJwdXNoIl19XX0.QhflHPfbd6eVF4lM9bwYpFZIV0PfikbyXuLx959ykRTBpe3CYnzs6YBK8FToVb5R47920PVLrh8zuLzdCr9t3w", "expires_in": 3600,"issued_at": "2009-11-10T23:00:00Z"}`
		return c.JSONBlob(200, []byte(str))
	})

	namespaceGroup := e.Group("/namespace", middlewares.AuthWithConfig(middlewares.AuthConfig{}))
	namespaceHandler := namespace.New()
	namespaceGroup.OPTIONS("/", func(c echo.Context) error {
		return c.String(200, "OK")
	})
	namespaceGroup.POST("/", namespaceHandler.PostNamespace)
	namespaceGroup.PUT("/:id", func(c echo.Context) error {
		return c.String(200, "OK")
	})
	namespaceGroup.DELETE("/:id", func(c echo.Context) error {
		return c.String(200, "OK")
	})
	namespaceGroup.GET("/:id", func(c echo.Context) error {
		return c.String(200, "OK")
	})
	namespaceGroup.GET("/", namespaceHandler.ListNamespace)

	e.Any("/v2/*", distribution.All)

	return nil
}
