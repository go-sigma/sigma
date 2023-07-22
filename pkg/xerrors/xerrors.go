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
)

var (
	// HTTPErrCodeOK is an OK error.
	HTTPErrCodeOK = ErrCode{HTTPStatusCode: http.StatusOK, Code: "OK", Title: "OK"}
	// HTTPErrCodeCreated is a created error.
	HTTPErrCodeCreated = ErrCode{HTTPStatusCode: http.StatusCreated, Code: "CREATED", Title: "Created"}
	// HTTPErrCodeBadRequest is a bad request error.
	HTTPErrCodeBadRequest = ErrCode{HTTPStatusCode: http.StatusBadRequest, Code: "BAD_REQUEST", Title: "Bad Request"}
	// HTTPErrCodeUnauthorized is an unauthorized error.
	HTTPErrCodeUnauthorized = ErrCode{HTTPStatusCode: http.StatusUnauthorized, Code: "UNAUTHORIZED", Title: "Unauthorized"}
	// HTTPErrCodeForbidden is a forbidden error.
	HTTPErrCodeForbidden = ErrCode{HTTPStatusCode: http.StatusForbidden, Code: "FORBIDDEN", Title: "Forbidden"}
	// HTTPErrCodeNotFound is a not found error.
	HTTPErrCodeNotFound = ErrCode{HTTPStatusCode: http.StatusNotFound, Code: "NOT_FOUND", Title: "Not Found"}
	// HTTPErrCodeMethodNotAllowed is a method not allowed error.
	HTTPErrCodeMethodNotAllowed = ErrCode{HTTPStatusCode: http.StatusMethodNotAllowed, Code: "METHOD_NOT_ALLOWED", Title: "Method Not Allowed"}
	// HTTPErrCodeNotAcceptable is a not acceptable error.
	HTTPErrCodeNotAcceptable = ErrCode{HTTPStatusCode: http.StatusNotAcceptable, Code: "NOT_ACCEPTABLE", Title: "Not Acceptable"}
	// HTTPErrCodeConflict is a conflict error.
	HTTPErrCodeConflict = ErrCode{HTTPStatusCode: http.StatusConflict, Code: "CONFLICT", Title: "Conflict"}
	// HTTPErrCodePaginationInvalid is a pagination invalid error.
	HTTPErrCodePaginationInvalid = ErrCode{HTTPStatusCode: http.StatusBadRequest, Code: "PAGINATION_INVALID", Title: "Pagination Invalid"}
	// HTTPErrCodeInternalError is an internal error.
	HTTPErrCodeInternalError = ErrCode{HTTPStatusCode: http.StatusInternalServerError, Code: "INTERNAL_ERROR", Title: "Internal Error"}
	// HTTPErrCodeVerificationCodeInvalid is a verification code error.
	HTTPErrCodeVerificationCodeInvalid = ErrCode{HTTPStatusCode: http.StatusBadRequest, Code: "VERIFICATION_CODE_INVALID", Title: "Verification Code Invalid"}
)
