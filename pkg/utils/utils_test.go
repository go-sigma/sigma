// Copyright 2023 XImager
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

package utils

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/assert"

	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/validators"
)

func TestPanicIf(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	PanicIf(fmt.Errorf("test panic"))
}

func TestGetContentLength(t *testing.T) {
	_, err := GetContentLength(nil)
	assert.Error(t, err)
	req, err := http.NewRequest(http.MethodGet, "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}
	length, err := GetContentLength(req)
	if err != nil {
		t.Fatal(err)
	}
	if length != 0 {
		t.Errorf("expected 0, got %d", length)
	}
	req.Header.Set("Content-Length", "123")
	length, err = GetContentLength(req)
	if err != nil {
		t.Fatal(err)
	}
	if length != 123 {
		t.Errorf("expected 123, got %d", length)
	}
	req.Header.Set("Content-Length", "abc")
	_, err = GetContentLength(req)
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestGenPathByDigest(t *testing.T) {
	dgest, err := digest.Parse("sha256:08e7660f72aaa312f2ad1e13bc35afd988fa476052fd83296e0702e31ea00141")
	assert.NoError(t, err)
	path := GenPathByDigest(dgest)
	assert.Equal(t, "sha256/08/e7/660f72aaa312f2ad1e13bc35afd988fa476052fd83296e0702e31ea00141", path)
}

func TestBindValidate(t *testing.T) {
	e := echo.New()
	validators.Initialize(e)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"username":"test","password":"123498712311Aa!","email":"test@xx.com"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	type User struct {
		Username string `json:"username" validate:"required,alphanum,min=2,max=20"`
		Password string `json:"password" validate:"required,min=6,max=20"`
		Email    string `json:"email" validate:"required,email"`
	}
	var user User
	err := BindValidate(c, &user)
	assert.NoError(t, err)

	req = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"username":"","password":"123498712311Aa!","email":"test@xx.com"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	err = BindValidate(c, &user)
	assert.Error(t, err)
}

func TestInject(t *testing.T) {
	var a = 1
	var b = 2
	err := Inject(&a, nil)
	assert.Equal(t, 1, a)
	assert.NoError(t, err)
	err = Inject(&a, &b)
	assert.Equal(t, 2, a)
	assert.NoError(t, err)
}

func TestNormalizePagination(t *testing.T) {
	type args struct {
		in types.Pagination
	}
	tests := []struct {
		name string
		args args
		want types.Pagination
	}{
		{
			name: "test 1",
			args: args{
				in: types.Pagination{
					Page:  ptr.Of(int(0)),
					Limit: ptr.Of(int(0)),
				},
			},
			want: types.Pagination{
				Page:  ptr.Of(int(1)),
				Limit: ptr.Of(int(10)),
			},
		},
		{
			name: "test 2",
			args: args{
				in: types.Pagination{
					Page:  ptr.Of(int(-1)),
					Limit: ptr.Of(int(0)),
				},
			},
			want: types.Pagination{
				Page:  ptr.Of(int(1)),
				Limit: ptr.Of(int(10)),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizePagination(tt.args.in); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NormalizePagination() = %v, want %v", got, tt.want)
			}
		})
	}
}
