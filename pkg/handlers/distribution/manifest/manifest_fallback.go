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

package manifest

import (
	"io"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/ximager/ximager/pkg/handlers/distribution/clients"
)

// fallbackProxy cannot found the manifest, proxy to the origin registry
func fallbackProxy(c echo.Context) (int, http.Header, []byte, error) {
	cli, err := clients.New()
	if err != nil {
		return 0, nil, nil, err
	}
	statusCode, header, reader, err := cli.DoRequest(c.Request().Method, c.Request().URL.Path)
	if err != nil {
		return 0, nil, nil, err
	}
	bodyBytes, err := io.ReadAll(reader)
	if err != nil {
		return 0, nil, nil, err
	}
	return statusCode, header, bodyBytes, nil
}