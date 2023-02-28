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

package xerrors

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/ximager/ximager/pkg/utils/ptr"
)

// HTTPError is an error that can be returned by an HTTP handler.
type HTTPError struct {
	// Code is a machine-readable error code.
	Code string `json:"code"`
	// Title is a human-readable title for the error.
	Title string `json:"title"`
	// Message is a human-readable error message. This field is optional.
	Message *string `json:"message,omitempty"`
}

// NewHTTPError creates a new HTTPError instance.
func NewHTTPError(c echo.Context, code HTTPErrCode, message ...string) error {
	err := HTTPError{
		Code:  code.Code,
		Title: code.Title,
	}
	if len(message) > 0 {
		err.Message = ptr.Of(message[0])
	}
	return c.JSON(code.StatusCode, err)
}

// HTTPErrCode is a struct that contains the error code, title and status code.
type HTTPErrCode struct {
	// StatusCode is the HTTP status code.
	StatusCode int
	// Code is a machine-readable error code.
	Code string
	// Title is a human-readable title for the error.
	Title string
}

var (
	// HTTPErrCodeOK is an OK error.
	HTTPErrCodeOK = HTTPErrCode{http.StatusOK, "OK", "OK"}
	// HTTPErrCodeCreated is a created error.
	HTTPErrCodeCreated = HTTPErrCode{http.StatusCreated, "CREATED", "Created"}
	// HTTPErrCodeBadRequest is a bad request error.
	HTTPErrCodeBadRequest = HTTPErrCode{http.StatusBadRequest, "BAD_REQUEST", "Bad Request"}
	// HTTPErrCodeUnauthorized is an unauthorized error.
	HTTPErrCodeUnauthorized = HTTPErrCode{http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized"}
	// HTTPErrCodeForbidden is a forbidden error.
	HTTPErrCodeForbidden = HTTPErrCode{http.StatusForbidden, "FORBIDDEN", "Forbidden"}
	// HTTPErrCodeNotFound is a not found error.
	HTTPErrCodeNotFound = HTTPErrCode{http.StatusNotFound, "NOT_FOUND", "Not Found"}
	// HTTPErrCodeMethodNotAllowed is a method not allowed error.
	HTTPErrCodeMethodNotAllowed = HTTPErrCode{http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method Not Allowed"}
	// HTTPErrCodeNotAcceptable is a not acceptable error.
	HTTPErrCodeNotAcceptable = HTTPErrCode{http.StatusNotAcceptable, "NOT_ACCEPTABLE", "Not Acceptable"}
	// HTTPErrCodeConflict is a conflict error.
	HTTPErrCodeConflict = HTTPErrCode{http.StatusConflict, "CONFLICT", "Conflict"}
	// HTTPErrCodePaginationInvalid is a pagination invalid error.
	HTTPErrCodePaginationInvalid = HTTPErrCode{http.StatusBadRequest, "PAGINATION_INVALID", "Pagination Invalid"}
	// HTTPErrCodeInternalError is an internal error.
	HTTPErrCodeInternalError = HTTPErrCode{http.StatusInternalServerError, "INTERNAL_ERROR", "Internal Error"}
)
