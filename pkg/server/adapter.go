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
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/go-sigma/sigma/pkg/server/errcode"
)

// HandlerFunc is the handler signature.
// Handlers are responsible for writing their own error responses before returning.
type HandlerFunc func(c *gin.Context)

// RequestHandlerFunc is the handler signature for endpoints with a typed,
// validated request object.
type RequestHandlerFunc[T any] func(c *gin.Context, req *T)

// Wrap converts a HandlerFunc into a gin.HandlerFunc.
func Wrap(h HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		h(c)
	}
}

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

// WrapHandle registers a HandlerFunc onto a gin.IRoutes (router or group).
func WrapHandle(r gin.IRoutes, method, path string, h HandlerFunc) gin.IRoutes {
	return r.Handle(method, path, Wrap(h))
}
