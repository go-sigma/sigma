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

package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

// api ...
type api struct {
	authorization string
	endpoint      string
	cli           *resty.Client
}

// NewAPI ...
func NewAPI(authorization, endpoint string) api {
	client := resty.New()
	if strings.HasPrefix(endpoint, "https") {
		client = resty.NewWithClient(&http.Client{
			Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}, // nolint: gosec
		})
	}
	return api{
		authorization: authorization,
		endpoint:      endpoint,
		cli:           client,
	}
}

// CreateCache ...
func (a api) CreateCache(ctx context.Context, builderID, runnerID int64, p string) error {
	file, err := os.Open(p)
	if err != nil {
		return err
	}
	code, _, err := a.DoRequest(ctx, http.MethodPost, fmt.Sprintf("/api/v1/caches/?builder_id=%d&runner_id=%d", builderID, runnerID), nil, file)
	if err != nil {
		return err
	}
	if code != http.StatusCreated {
		return fmt.Errorf("Create cache response status code(%d) is not 201", code)
	}
	return nil
}

// GetCache ...
func (a api) GetCache(ctx context.Context, builderID, runnerID int64) (io.ReadCloser, error) {
	code, reader, err := a.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/api/v1/caches/?builder_id=%d&runner_id=%d", builderID, runnerID), nil)
	if err != nil {
		return nil, err
	}
	if code == http.StatusNotFound {
		return nil, os.ErrNotExist
	} else if code != http.StatusOK {
		return nil, fmt.Errorf("Get cache response status code(%d) is not 200", code)
	}
	return reader, nil
}

func (a api) DoRequest(ctx context.Context, method, path string, headers http.Header, bodyReaders ...io.Reader) (int, io.ReadCloser, error) {
	req := a.cli.R()
	for k, vals := range headers {
		for _, val := range vals {
			req.Header.Add(k, val)
		}
	}
	req.SetHeader(echo.HeaderAuthorization, fmt.Sprintf("Bearer %s", a.authorization))
	req.SetHeader(echo.HeaderContentType, "application/json")
	req.SetDoNotParseResponse(true)
	if len(bodyReaders) != 0 {
		req.SetBody(bodyReaders[0])
	}
	req.SetContext(ctx)
	url := fmt.Sprintf("%s/%s", a.endpoint, strings.TrimPrefix(path, "/"))
	log.Info().Str("url", url).Str("method", method).Msg("Client do request")
	resp, err := req.Execute(method, url)
	if err != nil {
		return 0, nil, err
	}

	return resp.StatusCode(), resp.RawBody(), nil
}
