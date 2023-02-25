// The MIT License (MIT)
//
// Copyright © 2023 Tosone <i@tosone.cn>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

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
