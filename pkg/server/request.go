// Copyright 2026 sigma
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

package server

import (
	"encoding/json/v2"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"github.com/go-sigma/sigma/pkg/server/errcode"
)

// RequestHandlerFunc is the handler signature for endpoints with a typed,
// validated request object.
type RequestHandlerFunc[T any] func(c *gin.Context, req *T)

// WrapRequest binds and validates a typed request before invoking h.
func WrapRequest[T any](h RequestHandlerFunc[T]) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := new(T)
		if err := BindRequest(c, req); err != nil {
			slog.Error("bind and validate request failed", "err", err, "path", c.Request.URL.Path)
			errcode.NewHTTPError(c, errcode.HTTPErrCodeBadRequest, err.Error())
			return
		}
		h(c, req)
	}
}

// BindRequest binds request body, query parameters, and path parameters into
// dst, then validates the fully populated request once.
func BindRequest(c *gin.Context, dst any) error {
	if err := bindBody(c, dst); err != nil {
		return err
	}
	if err := binding.MapFormWithTag(dst, c.Request.URL.Query(), "query"); err != nil {
		return err
	}
	if err := binding.MapFormWithTag(dst, paramsMap(c.Params), "param"); err != nil {
		return err
	}
	if binding.Validator == nil {
		return nil
	}
	return binding.Validator.ValidateStruct(dst)
}

func bindBody(c *gin.Context, dst any) error {
	if c.Request == nil || c.Request.Body == nil || c.Request.Body == http.NoBody || c.Request.ContentLength == 0 {
		return nil
	}

	if c.ContentType() != binding.MIMEJSON {
		return nil
	}
	return json.UnmarshalRead(c.Request.Body, dst)
}

func paramsMap(params gin.Params) map[string][]string {
	values := make(map[string][]string, len(params))
	for _, param := range params {
		values[param.Key] = []string{param.Value}
	}
	return values
}
