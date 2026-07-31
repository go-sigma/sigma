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

package builder

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"resty.dev/v3"

	"github.com/go-sigma/sigma/pkg/consts"
)

// apiClient 缓存 API 客户端
type apiClient struct {
	authorization string
	endpoint      string
	cli           *resty.Client
}

// newAPIClient 创建缓存 API 客户端
func newAPIClient(authorization, endpoint string) apiClient {
	client := resty.New()
	if strings.HasPrefix(endpoint, "https") {
		client = resty.NewWithClient(&http.Client{
			Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}, // nolint: gosec
		})
	}
	return apiClient{
		authorization: authorization,
		endpoint:      endpoint,
		cli:           client,
	}
}

// createCache 上传构建缓存
func (a apiClient) createCache(ctx context.Context, builderID string, p string) error {
	file, err := osOpen(p)
	if err != nil {
		return err
	}
	code, _, err := a.doRequest(ctx, http.MethodPost, fmt.Sprintf("/api/v1/caches/%s", builderID), nil, file)
	if err != nil {
		return err
	}
	if code != http.StatusCreated {
		return fmt.Errorf("create cache response status code(%d) is not 201", code)
	}
	return nil
}

// getCache 下载构建缓存
func (a apiClient) getCache(ctx context.Context, builderID string) (io.ReadCloser, error) {
	code, reader, err := a.doRequest(ctx, http.MethodGet, fmt.Sprintf("/api/v1/caches/%s", builderID), nil)
	if err != nil {
		return nil, err
	}
	if code == http.StatusNotFound {
		return nil, osErrNotExist
	} else if code != http.StatusOK {
		return nil, fmt.Errorf("get cache response status code(%d) is not 200", code)
	}
	return reader, nil
}

func (a apiClient) doRequest(ctx context.Context, method, path string, headers http.Header, bodyReaders ...io.Reader) (int, io.ReadCloser, error) {
	req := a.cli.R()
	for k, vals := range headers {
		for _, val := range vals {
			req.Header.Add(k, val)
		}
	}
	req.SetHeader(consts.HeaderAuthorization, fmt.Sprintf("Bearer %s", a.authorization))
	req.SetHeader(consts.HeaderContentType, "application/json")
	req.SetResponseDoNotParse(true)
	if len(bodyReaders) != 0 {
		req.SetBody(bodyReaders[0])
	}
	req.SetContext(ctx)
	url := fmt.Sprintf("%s/%s", a.endpoint, strings.TrimPrefix(path, "/"))
	slog.Info("client do request", "url", url, "method", method)
	resp, err := req.Execute(method, url)
	if err != nil {
		return 0, nil, err
	}

	return resp.StatusCode(), resp.Body, nil
}
