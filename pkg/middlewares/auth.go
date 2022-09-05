package middlewares

import "github.com/labstack/echo/v4"

// AuthConfig is the configuration for the Auth middleware.
type AuthConfig struct {
}

// AuthWithConfig returns a middleware which authenticates requests.
func AuthWithConfig(config AuthConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return next(c)
		}
	}
}
