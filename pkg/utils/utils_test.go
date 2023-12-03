// Copyright 2023 sigma
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
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
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

func TestTrimHTTP(t *testing.T) {
	type args struct {
		in string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "common",
			args: args{
				in: "http://localhost:8080",
			},
			want: "localhost:8080",
		},
		{
			name: "common-1",
			args: args{
				in: "https://localhost:8080",
			},
			want: "localhost:8080",
		},
		{
			name: "common-2",
			args: args{
				in: "localhost:8080",
			},
			want: "localhost:8080",
		},
		{
			name: "common-3",
			args: args{
				in: "localhost:8080/",
			},
			want: "localhost:8080",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TrimHTTP(tt.args.in); got != tt.want {
				t.Errorf("TrimHTTP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsDir(t *testing.T) {
	convey.Convey("Check if given path is a directory", t, func() {
		convey.Convey("Pass a file name", func() {
			convey.So(IsDir("file.go"), convey.ShouldEqual, false)
		})
		convey.Convey("Pass a directory name", func() {
			convey.So(IsDir("ptr"), convey.ShouldEqual, true)
		})
		convey.Convey("Pass a invalid path", func() {
			convey.So(IsDir("foo"), convey.ShouldEqual, false)
		})
	})
}

func TestIsFile(t *testing.T) {
	if !IsFile("utils.go") {
		t.Errorf("IsExist:\n Expect => %v\n Got => %v\n", true, false)
	}

	if IsFile("ptr") {
		t.Errorf("IsExist:\n Expect => %v\n Got => %v\n", false, true)
	}

	if IsFile("files.go") {
		t.Errorf("IsExist:\n Expect => %v\n Got => %v\n", false, true)
	}
}

func TestIsExist(t *testing.T) {
	convey.Convey("Check if file or directory exists", t, func() {
		convey.Convey("Pass a file name that exists", func() {
			convey.So(IsExist("utils.go"), convey.ShouldEqual, true)
		})
		convey.Convey("Pass a directory name that exists", func() {
			convey.So(IsExist("ptr"), convey.ShouldEqual, true)
		})
		convey.Convey("Pass a directory name that does not exist", func() {
			convey.So(IsExist(".hg"), convey.ShouldEqual, false)
		})
	})
}

func TestDirWithSlash(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "common",
			args: args{
				id: "",
			},
			want: "",
		},
		{
			name: "common",
			args: args{
				id: "1",
			},
			want: "1",
		},
		{
			name: "common",
			args: args{
				id: "12",
			},
			want: "12",
		},
		{
			name: "common",
			args: args{
				id: "123",
			},
			want: "12/3",
		},
		{
			name: "common",
			args: args{
				id: "1234",
			},
			want: "12/34",
		},
		{
			name: "common",
			args: args{
				id: "12345",
			},
			want: "12/34/5",
		},
		{
			name: "common",
			args: args{
				id: "123456789",
			},
			want: "12/34/56789",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DirWithSlash(tt.args.id); got != tt.want {
				t.Errorf("DirWithSlash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringsJoin(t *testing.T) {
	type args struct {
		strs []enums.OciPlatform
		sep  string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "common",
			args: args{
				strs: []enums.OciPlatform{},
				sep:  ",",
			},
			want: "",
		},
		{
			name: "common",
			args: args{
				strs: []enums.OciPlatform{
					enums.OciPlatformLinuxAmd64,
				},
				sep: ",",
			},
			want: "linux/amd64",
		},
		{
			name: "common",
			args: args{
				strs: []enums.OciPlatform{
					enums.OciPlatformLinux386,
					enums.OciPlatformLinuxAmd64,
				},
				sep: ",",
			},
			want: "linux/386,linux/amd64",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StringsJoin(tt.args.strs, tt.args.sep); got != tt.want {
				t.Errorf("StringsJoin() = %v, want %v", got, tt.want)
			}
		})
	}
}
