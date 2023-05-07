package clients

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/ximager/ximager/pkg/types"
)

func TestBasicAuthToken(t *testing.T) {
	cUsername := "ximager"
	cPassword := "ximager"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path { // nolint: gocritic
		case "/v2/":
			username, password, ok := r.BasicAuth()
			if ok {
				if username == cUsername && password == cPassword {
					w.WriteHeader(http.StatusOK)
					return
				}
			}
			w.Header().Set("Www-Authenticate", `Basic realm="basic-realm"`)
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	defer srv.Close()

	viper.SetDefault("log.level", "debug")
	viper.SetDefault("proxy.endpoint", srv.URL)
	viper.SetDefault("proxy.tlsVerify", true)
	viper.SetDefault("proxy.username", cUsername)
	viper.SetDefault("proxy.password", cPassword)

	_, err := New()
	assert.NoError(t, err)
}

func TestTLSBasicAuthToken(t *testing.T) {
	cUsername := "ximager"
	cPassword := "ximager"
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path { // nolint: gocritic
		case "/v2/":
			username, password, ok := r.BasicAuth()
			if ok {
				if username == cUsername && password == cPassword {
					w.WriteHeader(http.StatusOK)
					return
				}
			}
			w.Header().Set("Www-Authenticate", `Basic realm="basic-realm"`)
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	defer srv.Close()

	viper.SetDefault("log.level", "info")
	viper.SetDefault("proxy.endpoint", srv.URL)
	viper.SetDefault("proxy.tlsVerify", false)
	viper.SetDefault("proxy.username", cUsername)
	viper.SetDefault("proxy.password", cPassword)

	_, err := New()
	assert.NoError(t, err)
}

func TestBearerAuthToken(t *testing.T) {
	var wwwAuthenticate string

	cUsername := "ximager"
	cPassword := "ximager"
	token := "ximager"
	service := "registry.docker.io"
	scope := "repository:library/alpine:pull"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v2/":
			if r.Header.Get("Authorization") == "Bearer "+token {
				w.WriteHeader(http.StatusOK)
				return
			}
			w.Header().Set("Www-Authenticate", wwwAuthenticate)
			w.WriteHeader(http.StatusUnauthorized)
		case "/user/token":
			username, password, ok := r.BasicAuth()
			if ok {
				if username == cUsername && password == cPassword {
					query := r.URL.Query()
					svc := query.Get("service")
					so := query.Get("scope")
					if svc != service || so != scope {
						t.Error("service or scope not match")
					}
					w.WriteHeader(http.StatusOK)
					err := json.NewEncoder(w).Encode(types.PostUserTokenResponse{Token: token})
					assert.NoError(t, err)
					return
				}
			}
		}
	}))
	defer srv.Close()

	viper.SetDefault("log.level", "info")
	viper.SetDefault("proxy.endpoint", srv.URL)
	viper.SetDefault("proxy.tlsVerify", true)
	viper.SetDefault("proxy.username", cUsername)
	viper.SetDefault("proxy.password", cPassword)

	wwwAuthenticate = fmt.Sprintf(`Bearer realm="%s",service="%s",scope="%s"`, srv.URL+"/user/token", "registry.docker.io", "repository:library/alpine:pull")

	_, err := New()
	assert.NoError(t, err)
}
