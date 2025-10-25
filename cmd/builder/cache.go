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
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"resty.dev/v3"

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
		return fmt.Errorf("create cache response status code(%d) is not 201", code)
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
		return nil, fmt.Errorf("get cache response status code(%d) is not 200", code)
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

	return resp.StatusCode(), resp.Body, nil
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
			log.Error().Err(err).Msg("cache file close failed")
		}
	}
	if utils.IsFile(path.Join(cache, compressedCache)) {
		log.Info().Msg("Start to decompress cache")
		// 使用zstd命令解压缩
		cmd := exec.Command("zstd", "-d", path.Join(cache, compressedCache), "-o", path.Join(home, "cache.tar"))
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("decompress cache failed: %v", err)
		}
		
		// 解包tar文件
		cmd = exec.Command("tar", "-xf", path.Join(home, "cache.tar"), "-C", home)
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("untar cache failed: %v", err)
		}
		
		// 清理临时tar文件
		_ = os.Remove(path.Join(home, "cache.tar"))
		
		fileInfo, err := os.Stat(path.Join(cache, compressedCache))
		if err != nil {
			return fmt.Errorf("read compressed file failed: %v", err)
		}
		err = os.Rename(cacheOut, cacheIn)
		if err != nil {
			return fmt.Errorf("rename cache_out to cache_in failed: %v", err)
		}
		log.Info().Str("size", humanize.BigIBytes(big.NewInt(fileInfo.Size()))).Msg("decompress cache success")
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
	log.Info().Msg("start to compress cache")
	
	// 先打包成tar文件
	cmd := exec.Command("tar", "-cf", path.Join("/tmp", "cache.tar"), "-C", cacheOut, ".")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("tar cache failed: %v", err)
	}
	
	// 使用zstd命令压缩
	cmd = exec.Command("zstd", path.Join("/tmp", "cache.tar"), "-o", path.Join("/tmp", compressedCache))
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("compress cache failed: %v", err)
	}
	
	// 清理临时tar文件
	_ = os.Remove(path.Join("/tmp", "cache.tar"))
	
	err = os.Rename(path.Join("/tmp", compressedCache), path.Join(cache, compressedCache))
	if err != nil {
		return fmt.Errorf("move compressed file to dir failed: %v", err)
	}
	fileInfo, err := os.Stat(path.Join(cache, compressedCache))
	if err != nil {
		return fmt.Errorf("read compressed file failed: %v", err)
	}
	err = b.api.CreateCache(context.Background(), b.BuilderID, path.Join(cache, compressedCache))
	if err != nil {
		return fmt.Errorf("export cache to server failed: %v", err)
	}
	log.Info().Str("size", humanize.BigIBytes(big.NewInt(fileInfo.Size()))).Msg("export cache success")
	return nil
}
