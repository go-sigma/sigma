package echoplus

import (
	"regexp"

	"github.com/labstack/echo/v4"

	"github.com/go-sigma/sigma/pkg/middlewares/authn"
	"github.com/go-sigma/sigma/pkg/middlewares/authz"
)

// Plus ...
type Plus interface {
	// Get ...
	Get(path string, authnConfig *authn.AuthnConfig, authzConfig *authz.AuthzConfig, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	// Post ...
	Post(path string, authnConfig *authn.AuthnConfig, authzConfig *authz.AuthzConfig, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	// Put ...
	Put(path string, authnConfig *authn.AuthnConfig, authzConfig *authz.AuthzConfig, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	// Delete ...
	Delete(path string, authnConfig *authn.AuthnConfig, authzConfig *authz.AuthzConfig, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	// Patch ...
	Patch(path string, authnConfig *authn.AuthnConfig, authzConfig *authz.AuthzConfig, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
}

type engine interface {
	// Get ...
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	// POST ...
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	// PUT ...
	PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	// DELETE ...
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	// PATCH ...
	PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
}

type plus struct {
	engine engine
}

// New ...
func New(e engine) Plus {
	return &plus{
		engine: e,
	}
}

// Get ...
func (p *plus) Get(path string, authnConfig *authn.AuthnConfig, authzConfig *authz.AuthzConfig, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	authn.AuthMapper[regexp.MustCompile(path)] = authnConfig
	authz.AuthMapper[regexp.MustCompile(path)] = authzConfig
	return p.engine.GET(path, h, m...)
}

// Post ...
func (p *plus) Post(path string, authnConfig *authn.AuthnConfig, authzConfig *authz.AuthzConfig, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	authn.AuthMapper[regexp.MustCompile(path)] = authnConfig
	authz.AuthMapper[regexp.MustCompile(path)] = authzConfig
	return p.engine.POST(path, h, m...)
}

// Put ...
func (p *plus) Put(path string, authnConfig *authn.AuthnConfig, authzConfig *authz.AuthzConfig, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	authn.AuthMapper[regexp.MustCompile(path)] = authnConfig
	authz.AuthMapper[regexp.MustCompile(path)] = authzConfig
	return p.engine.PUT(path, h, m...)
}

// Delete ...
func (p *plus) Delete(path string, authnConfig *authn.AuthnConfig, authzConfig *authz.AuthzConfig, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	authn.AuthMapper[regexp.MustCompile(path)] = authnConfig
	authz.AuthMapper[regexp.MustCompile(path)] = authzConfig
	return p.engine.DELETE(path, h, m...)
}

// Patch ...
func (p *plus) Patch(path string, authnConfig *authn.AuthnConfig, authzConfig *authz.AuthzConfig, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	authn.AuthMapper[regexp.MustCompile(path)] = authnConfig
	authz.AuthMapper[regexp.MustCompile(path)] = authzConfig
	return p.engine.PATCH(path, h, m...)
}
