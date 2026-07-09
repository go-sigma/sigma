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

package mcpserver

import (
	"errors"

	"github.com/go-sigma/sigma/pkg/server/errcode"
)

var (
	errUnauthorized = errors.New("unauthorized")
	errForbidden    = errors.New("permission denied")
)

func toolErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, errUnauthorized) {
		return errUnauthorized.Error()
	}
	if errors.Is(err, errForbidden) {
		return errForbidden.Error()
	}
	if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
		if e.Description != "" {
			return e.Description
		}
		if e.Title != "" {
			return e.Title
		}
		return e.Code
	}
	return "internal error"
}

func statusCodeFromError(err error) int {
	if err == nil {
		return 200
	}
	if errors.Is(err, errUnauthorized) || errors.Is(err, errForbidden) {
		return 401
	}
	if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
		return e.HTTPStatusCode
	}
	return 500
}
