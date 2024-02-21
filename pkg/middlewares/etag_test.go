// Copyright 2024 sigma
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

var testWeakEtag = "W/11-8dcfee46"
var testStrEtag = "11-a591a6d40bf420404a011733cfb7b190d62c65bf0bcda32b57b277d9ad9f146e"
var e *echo.Echo

func init() {
	e = echo.New()
	e.GET("/etag", func(c echo.Context) error {
		return c.String(200, "Hello World")
	}, WithEtagConfig(EtagConfig{Weak: false}))

	e.GET("/etag/weak", func(c echo.Context) error {
		return c.String(200, "Hello World")
	}, Etag())
}

func TestStrongEtag(t *testing.T) {
	// Test strong Etag
	req := httptest.NewRequest(http.MethodGet, "/etag", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	if rec.Header().Get("Etag") != testStrEtag {
		t.Errorf("Expected Etag %s, got %s", testStrEtag, rec.Header().Get("Etag"))
	}

	if rec.Body.String() != "Hello World" {
		t.Errorf("Expected body %s, got %s", "Hello World", rec.Body.String())
	}

	// Test If-None-Match
	req = httptest.NewRequest(http.MethodGet, "/etag", nil)
	req.Header.Set("If-None-Match", testStrEtag)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotModified {
		t.Errorf("Expected status code %d, got %d", http.StatusNotModified, rec.Code)
	}

	if rec.Header().Get("Etag") != testStrEtag {
		t.Errorf("Expected Etag %s, got %s", testStrEtag, rec.Header().Get("Etag"))
	}

	if rec.Body.String() != "" {
		t.Errorf("Expected body %s, got %s", "", rec.Body.String())
	}

	// Test If-None-Match invalid
	req = httptest.NewRequest(http.MethodGet, "/etag", nil)
	req.Header.Set("If-None-Match", "invalid")
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	if rec.Header().Get("Etag") != testStrEtag {
		t.Errorf("Expected Etag %s, got %s", testStrEtag, rec.Header().Get("Etag"))
	}

	if rec.Body.String() != "Hello World" {
		t.Errorf("Expected body %s, got %s", "Hello World", rec.Body.String())
	}
}

func TestWeakEtag(t *testing.T) {
	// Test weak Etag
	req := httptest.NewRequest(http.MethodGet, "/etag/weak", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	if rec.Header().Get("Etag") != testWeakEtag {
		t.Errorf("Expected Etag %s, got %s", testWeakEtag, rec.Header().Get("Etag"))
	}

	if rec.Body.String() != "Hello World" {
		t.Errorf("Expected body %s, got %s", "Hello World", rec.Body.String())
	}

	// Test If-None-Match weak
	req = httptest.NewRequest(http.MethodGet, "/etag/weak", nil)
	req.Header.Set("If-None-Match", testWeakEtag)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotModified {
		t.Errorf("Expected status code %d, got %d", http.StatusNotModified, rec.Code)
	}

	if rec.Header().Get("Etag") != testWeakEtag {
		t.Errorf("Expected Etag %s, got %s", testWeakEtag, rec.Header().Get("Etag"))
	}

	if rec.Body.String() != "" {
		t.Errorf("Expected body %s, got %s", "", rec.Body.String())
	}

	// Test If-None-Match weak invalid
	req = httptest.NewRequest(http.MethodGet, "/etag/weak", nil)
	req.Header.Set("If-None-Match", "invalid")
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	if rec.Header().Get("Etag") != testWeakEtag {
		t.Errorf("Expected Etag %s, got %s", testWeakEtag, rec.Header().Get("Etag"))
	}

	if rec.Body.String() != "Hello World" {
		t.Errorf("Expected body %s, got %s", "Hello World", rec.Body.String())
	}
}
