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
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/go-resty/resty/v2"
	"github.com/labstack/echo/v4"
	"github.com/mholt/archiver/v3"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/utils"
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
func (a api) CreateCache(ctx context.Context, builderID int64, p string) error {
	file, err := os.Open(p)
	if err != nil {
		return err
	}
	code, _, err := a.DoRequest(ctx, http.MethodPost, fmt.Sprintf("/api/v1/caches/%d", builderID), nil, file)
	if err != nil {
		return err
	}
	if code != http.StatusCreated {
		return fmt.Errorf("Create cache response status code(%d) is not 201", code)
	}
	return nil
}

// GetCache ...
func (a api) GetCache(ctx context.Context, builderID int64) (io.ReadCloser, error) {
	code, reader, err := a.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/api/v1/caches/%d", builderID), nil)
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

func (b Builder) initCache() error {
	reader, err := b.api.GetCache(context.Background(), b.BuilderID)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if reader != nil {
		file, err := os.OpenFile(path.Join(cache, compressedCache), os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		_, err = io.Copy(file, reader)
		if err != nil {
			return err
		}
		err = file.Close()
		if err != nil {
			log.Error().Err(err).Msg("Cache file close failed")
		}
	}
	if utils.IsFile(path.Join(cache, compressedCache)) {
		log.Info().Msg("Start to decompress cache")
		err := archiver.Unarchive(path.Join(cache, compressedCache), home)
		if err != nil {
			return fmt.Errorf("Decompress cache failed: %v", err)
		}
		fileInfo, err := os.Stat(path.Join(cache, compressedCache))
		if err != nil {
			return fmt.Errorf("Read compressed file failed: %v", err)
		}
		err = os.Rename(cacheOut, cacheIn)
		if err != nil {
			return fmt.Errorf("Rename cache_out to cache_in failed: %v", err)
		}
		log.Info().Str("size", humanize.BigIBytes(big.NewInt(fileInfo.Size()))).Msg("Decompress cache success")
	}
	var dirs = []string{cacheOut, cacheIn}
	for _, dir := range dirs {
		if !utils.IsDir(dir) {
			err := os.MkdirAll(dir, 0755)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (b Builder) exportCache() error {
	log.Info().Msg("Start to compress cache")
	tgz := archiver.NewTarGz()
	err := tgz.Archive([]string{path.Join(cacheOut)}, path.Join("/tmp", compressedCache))
	if err != nil {
		return fmt.Errorf("Compress cache failed: %v", err)
	}
	err = os.Rename(path.Join("/tmp", compressedCache), path.Join(cache, compressedCache))
	if err != nil {
		return fmt.Errorf("Move compressed file to dir failed")
	}
	fileInfo, err := os.Stat(path.Join(cache, compressedCache))
	if err != nil {
		return fmt.Errorf("Read compressed file failed: %v", err)
	}
	err = b.api.CreateCache(context.Background(), b.BuilderID, path.Join(cache, compressedCache))
	if err != nil {
		return fmt.Errorf("Export cache to server failed: %v", err)
	}
	log.Info().Str("size", humanize.BigIBytes(big.NewInt(fileInfo.Size()))).Msg("Export cache success")
	return nil
}
