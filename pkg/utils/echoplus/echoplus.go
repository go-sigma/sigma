package echoplus

import (
	"github.com/labstack/echo/v4"
)

type Plus interface {
}

type plus struct {
	*echo.Echo
}

// New ...
func New(echo *echo.Echo) Plus {
	return &plus{
		Echo: echo,
	}
}

// AuthzConfig ...
type AuthzConfig struct {
	Skip   bool
	Source []AuthzConfigSource
}

// AuthzConfigSource ...
type AuthzConfigSource struct {
	Name     string `json:"name"`
	Position string `json:"position"`
}

// AuthnConfig ...
type AuthnConfig struct {
	Skip bool
}

// AuthConfig ...
type AuthConfig struct {
	AuthnConfig *AuthnConfig
	AuthzConfig *AuthzConfig
}

// AuthMapper ...
var AuthMapper = make(map[string]AuthConfig)

// Get ...
func (p *plus) Get(path string, authnConfig *AuthnConfig, authzConfig *AuthzConfig, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	AuthMapper[path] = AuthConfig{
		AuthnConfig: authnConfig,
		AuthzConfig: authzConfig,
	}
	return p.Echo.GET(path, h, m...)
}

// Post ...
func (p *plus) Post(path string, authnConfig *AuthnConfig, authzConfig *AuthzConfig, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	AuthMapper[path] = AuthConfig{
		AuthnConfig: authnConfig,
		AuthzConfig: authzConfig,
	}
	return p.Echo.POST(path, h, m...)
}

// Put ...
func (p *plus) Put(path string, authnConfig *AuthnConfig, authzConfig *AuthzConfig, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	AuthMapper[path] = AuthConfig{
		AuthnConfig: authnConfig,
		AuthzConfig: authzConfig,
	}
	return p.Echo.PUT(path, h, m...)
}

// Delete ...
func (p *plus) Delete(path string, authnConfig *AuthnConfig, authzConfig *AuthzConfig, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	AuthMapper[path] = AuthConfig{
		AuthnConfig: authnConfig,
		AuthzConfig: authzConfig,
	}
	return p.Echo.DELETE(path, h, m...)
}

// Patch ...
func (p *plus) Patch(path string, authnConfig *AuthnConfig, authzConfig *AuthzConfig, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	AuthMapper[path] = AuthConfig{
		AuthnConfig: authnConfig,
		AuthzConfig: authzConfig,
	}
	return p.Echo.PATCH(path, h, m...)
}
