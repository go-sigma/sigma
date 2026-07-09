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
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	"os"
	"os/exec"
	"path"

	"github.com/dustin/go-humanize"

	"github.com/go-sigma/sigma/pkg/utils"
)

const (
	cache           = "/opt/cache"
	cacheIn         = "/opt/cache_in"
	cacheOut        = "/opt/cache_out"
	compressedCache = "cache.tgz"
)

// osOpen wraps os.Open for testability
var osOpen = os.Open

// osOpenFile wraps os.OpenFile for testability
var osOpenFile = os.OpenFile

// osErrNotExist wraps os.ErrNotExist for testability
var osErrNotExist = os.ErrNotExist

// initCache 下载并解压构建缓存
func (f *BuildFlow) initCache() error {
	reader, err := f.api.getCache(context.Background(), f.BuilderID)
	if err != nil && !errors.Is(err, osErrNotExist) {
		return err
	}
	if reader != nil {
		file, err := osOpenFile(path.Join(cache, compressedCache), os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		_, err = io.Copy(file, reader)
		if err != nil {
			return err
		}
		err = file.Close()
		if err != nil {
			slog.Error("cache file close failed", "err", err)
		}
	}
	if utils.IsFile(path.Join(cache, compressedCache)) {
		slog.Info("start to decompress cache")
		cmd := exec.Command("zstd", "-d", path.Join(cache, compressedCache), "-o", path.Join(home, "cache.tar")) // #nosec G204 -- command is fixed and paths are constrained to builder cache directories
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("decompress cache failed: %v", err)
		}

		cmd = exec.Command("tar", "-xf", path.Join(home, "cache.tar"), "-C", home) // #nosec G204 -- command is fixed and paths are constrained to builder cache directories
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("untar cache failed: %v", err)
		}

		_ = os.Remove(path.Join(home, "cache.tar"))

		fileInfo, err := os.Stat(path.Join(cache, compressedCache))
		if err != nil {
			return fmt.Errorf("read compressed file failed: %v", err)
		}
		err = os.Rename(cacheOut, cacheIn)
		if err != nil {
			return fmt.Errorf("rename cache_out to cache_in failed: %v", err)
		}
		slog.Info("decompress cache success", "size", humanize.BigIBytes(big.NewInt(fileInfo.Size())))
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

// exportCache 压缩并上传构建缓存
func (f *BuildFlow) exportCache() error {
	slog.Info("start to compress cache")

	cmd := exec.Command("tar", "-cf", path.Join("/tmp", "cache.tar"), "-C", cacheOut, ".") // #nosec G204 -- command is fixed and paths are constrained to builder cache directories
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("tar cache failed: %v", err)
	}

	cmd = exec.Command("zstd", path.Join("/tmp", "cache.tar"), "-o", path.Join("/tmp", compressedCache)) // #nosec G204 -- command is fixed and paths are constrained to builder cache directories
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("compress cache failed: %v", err)
	}

	_ = os.Remove(path.Join("/tmp", "cache.tar"))

	err = os.Rename(path.Join("/tmp", compressedCache), path.Join(cache, compressedCache))
	if err != nil {
		return fmt.Errorf("move compressed file to dir failed: %v", err)
	}
	fileInfo, err := os.Stat(path.Join(cache, compressedCache))
	if err != nil {
		return fmt.Errorf("read compressed file failed: %v", err)
	}
	err = f.api.createCache(context.Background(), f.BuilderID, path.Join(cache, compressedCache))
	if err != nil {
		return fmt.Errorf("export cache to server failed: %v", err)
	}
	slog.Info("export cache success", "size", humanize.BigIBytes(big.NewInt(fileInfo.Size())))
	return nil
}
