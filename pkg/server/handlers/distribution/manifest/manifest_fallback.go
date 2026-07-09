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

package manifest

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/distribution/clients"
)

// fallbackProxy cannot found the manifest, proxy to the origin registry
func (h *handler) fallbackProxy(c *gin.Context) (int, http.Header, []byte, error) {
	var headers = make(http.Header)
	headers.Add(consts.HeaderAccept, "application/vnd.docker.distribution.manifest.v2+json")
	headers.Add(consts.HeaderAccept, "application/vnd.oci.image.manifest.v1+json")
	headers.Add(consts.HeaderAccept, "application/vnd.docker.distribution.manifest.list.v2+json")
	headers.Add(consts.HeaderAccept, "application/vnd.oci.image.index.v1+json")

	f := clients.NewClientsFactory()
	cli, err := f.New(h.Config) // TODO: config param
	if err != nil {
		return 0, nil, nil, err
	}
	statusCode, header, reader, err := cli.DoRequest(c.Request.Context(), c.Request.Method, c.Request.URL.Path, headers)
	if err != nil {
		return 0, nil, nil, err
	}
	bodyBytes, err := io.ReadAll(reader)
	if err != nil {
		return 0, nil, nil, err
	}
	slog.Info("", "manifest", string(bodyBytes), "method", c.Request.Method, "path", c.Request.URL.Path, "headers", headers)
	return statusCode, header, bodyBytes, nil
}
