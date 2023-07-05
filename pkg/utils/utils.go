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
	"fmt"
	"net/http"
	"strconv"

	"github.com/jinzhu/copier"
	"github.com/labstack/echo/v4"
	"github.com/opencontainers/go-digest"

	"github.com/ximager/ximager/pkg/types"
	"github.com/ximager/ximager/pkg/utils/ptr"
)

// GetContentLength returns the content length of the request.
func GetContentLength(req *http.Request) (int64, error) {
	if req == nil {
		return 0, fmt.Errorf("request is nil")
	}
	str := req.Header.Get("Content-Length")
	if str == "" {
		return 0, nil
	}
	length, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("content length is not a number")
	}
	return length, nil
}

// GenPathByDigest generates the path by the digest.
func GenPathByDigest(digest digest.Digest) string {
	hex := digest.Hex()
	return fmt.Sprintf("%s/%s/%s/%s", digest.Algorithm(), hex[0:2], hex[2:4], hex[4:])
}

// BindValidate binds and validates the request body
func BindValidate(c echo.Context, data any) error {
	err := c.Bind(data)
	if err != nil {
		return err
	}
	err = c.Validate(data)
	if err != nil {
		return err
	}
	return nil
}

// PanicIf panics if err is not nil
func PanicIf(err error) {
	if err != nil {
		panic(err)
	}
}

// Inject injects source to target if source is not nil
func Inject(target any, source any) error {
	if source == nil {
		return nil
	}
	return copier.Copy(target, source)
}

// NormalizePagination normalizes the pagination
func NormalizePagination(in types.Pagination) types.Pagination {
	if in.Last == nil || ptr.To(in.Last) < 0 {
		in.Last = ptr.Of(int64(0))
	}
	if in.Limit == nil || ptr.To(in.Limit) > 100 || ptr.To(in.Limit) <= 0 {
		in.Limit = ptr.Of(int(10))
	}
	return in
}
