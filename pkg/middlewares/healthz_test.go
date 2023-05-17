package middlewares

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

type testOKResource struct{}

func (t testOKResource) HealthCheck() error {
	return nil
}

type testFailedResource struct{}

func (t testFailedResource) HealthCheck() error {
	return fmt.Errorf("failed")
}

func TestHealthzOK(t *testing.T) {
	var ok testOKResource
	mr := Healthz(ok)(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(req, rec)
	err := mr(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestHealthzFailed(t *testing.T) {
	var ok testFailedResource
	mr := Healthz(ok)(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(req, rec)
	err := mr(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestHealthzNext(t *testing.T) {
	var ok testOKResource
	mr := Healthz(ok)(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
	req := httptest.NewRequest(http.MethodGet, "/healthz-test", nil)
	rec := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(req, rec)
	err := mr(c)
	assert.NoError(t, err)
}
