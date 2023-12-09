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

package xerrors

import (
	"fmt"

	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

// ErrCode provides relevant information about a given error code.
type ErrCode struct {
	// Code provides a unique, string key, often capitalized with
	// underscores, to identify the error code. This value is used as the
	// keyed value when serializing api errors.
	Code string `json:"code" example:"UNAUTHORIZED"`

	// Title is a short, human readable description of the error condition
	// included in API responses.
	Title string `json:"title" example:"Authentication required"`

	// Description provides a complete account of the errors purpose, suitable
	// for use in documentation.
	Description string `json:"description" example:"The access controller was unable to authenticate the client. Often this will be accompanied by a Www-Authenticate HTTP response header indicating how to authenticate."`

	// HTTPStatusCode provides the http status code that is associated with
	// this error condition.
	HTTPStatusCode int `json:"http_status_code" example:"401"`
}

// Error ...
func (e ErrCode) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Title)
}

// Detail ...
func (e *ErrCode) Detail(desc string) ErrCode {
	e.Description = desc
	return ptr.To(e)
}

var _ error = ErrCode{}
