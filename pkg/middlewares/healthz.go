package middlewares

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/ximager/ximager/pkg/xerrors"
)

// Resource something that need be health checked
type Resource interface {
	HealthCheck() error // returns error if health check no passed
}

// Healthz create a health check middleware
func Healthz(rs ...Resource) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().URL.Path == "/healthz" {
				for _, r := range rs {
					if err := r.HealthCheck(); err != nil {
						return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
					}
				}
				return c.String(http.StatusOK, "OK")
			}
			return next(c)
		}
	}
}
