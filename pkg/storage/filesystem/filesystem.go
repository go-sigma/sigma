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

package filesystem

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/storage"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

// fs is the filesystem storage driver
type fs struct {
	rootDirectory string
}

func init() {
	utils.PanicIf(storage.Register(enums.StorageTypeFilesystem, &factory{}))
}

type factory struct{}

var _ storage.Factory = factory{}

// New returns a new filesystem storage driver
func (f factory) New(config *config.Configuration) (storage.StorageDriver, error) {
	driver := &fs{rootDirectory: path.Join(config.Storage.Filesystem.Path, config.Storage.RootDirectory)}
	if !utils.IsExist(driver.rootDirectory) {
		err := os.MkdirAll(driver.rootDirectory, 0755)
		if err != nil {
			return nil, err
		}
	}
	if !utils.IsExist(path.Join(driver.rootDirectory, consts.BlobUploads)) {
		err := os.MkdirAll(path.Join(driver.rootDirectory, consts.BlobUploads), 0755)
		if err != nil {
			return nil, err
		}
	}
	if !utils.IsExist(path.Join(driver.rootDirectory, consts.BlobUploadParts)) {
		err := os.MkdirAll(path.Join(driver.rootDirectory, consts.BlobUploadParts), 0755)
		if err != nil {
			return nil, err
		}
	}
	if !utils.IsExist(path.Join(driver.rootDirectory, consts.Blobs)) {
		err := os.MkdirAll(path.Join(driver.rootDirectory, consts.Blobs), 0755)
		if err != nil {
			return nil, err
		}
	}
	if !utils.IsExist(path.Join(driver.rootDirectory, consts.Manifests)) {
		err := os.MkdirAll(path.Join(driver.rootDirectory, consts.Manifests), 0755)
		if err != nil {
			return nil, err
		}
	}
	if !utils.IsExist(path.Join(driver.rootDirectory, consts.DirCache)) {
		err := os.MkdirAll(path.Join(driver.rootDirectory, consts.DirCache), 0755)
		if err != nil {
			return nil, err
		}
	}
	return driver, nil
}

// Move moves a file from sourcePath to destPath
func (f *fs) Move(ctx context.Context, sourcePath string, destPath string) error {
	if !utils.IsExist(filepath.Dir(storage.SanitizePath(f.rootDirectory, destPath))) {
		err := os.MkdirAll(filepath.Dir(storage.SanitizePath(f.rootDirectory, destPath)), 0755)
		if err != nil {
			return err
		}
	}
	return os.Rename(storage.SanitizePath(f.rootDirectory, sourcePath), storage.SanitizePath(f.rootDirectory, destPath))
}

// Delete deletes a file at the given path
func (f *fs) Delete(ctx context.Context, path string) error {
	return os.RemoveAll(storage.SanitizePath(f.rootDirectory, path))
}

// Reader returns a reader for the file at the given path
func (f *fs) Reader(ctx context.Context, path string) (io.ReadCloser, error) {
	fp, err := os.OpenFile(storage.SanitizePath(f.rootDirectory, path), os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	return fp, nil
}

// CreateUploadID creates a new multipart upload and returns an
// opaque upload ID.
func (f *fs) CreateUploadID(ctx context.Context, _ string) (string, error) {
	uploadID := uuid.NewV7String()
	return uploadID, os.MkdirAll(storage.SanitizePath(f.rootDirectory, path.Join(consts.BlobUploadParts, uploadID)), 0755)
}

// WritePart writes a part of a multipart upload.
func (f *fs) UploadPart(ctx context.Context, _, uploadID string, _ int32, body io.Reader) (string, error) {
	eTag := uuid.NewV7String()
	fp, err := os.OpenFile(storage.SanitizePath(f.rootDirectory, path.Join(consts.BlobUploadParts, uploadID, eTag)), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return "", err
	}
	_, err = io.Copy(fp, body)
	if err != nil {
		return "", err
	}
	return eTag, fp.Close()
}

// CommitUpload commits a multipart upload.
func (f *fs) CommitUpload(ctx context.Context, rPath, uploadID string, parts []string) error {
	rPath = storage.SanitizePath(f.rootDirectory, rPath)
	fake := path.Join(rPath + ".fake")
	fp, err := os.OpenFile(fake, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	for _, part := range parts {
		partPath := storage.SanitizePath(f.rootDirectory, path.Join(consts.BlobUploadParts, uploadID, part))
		partFP, err := os.Open(partPath)
		if err != nil {
			fp.Close() // nolint: errcheck
			return err
		}
		_, err = io.Copy(fp, partFP)
		if err != nil {
			fp.Close() // nolint: errcheck
			return err
		}
		partFP.Close() // nolint: errcheck
	}
	err = fp.Close()
	if err != nil {
		return err
	}
	err = os.RemoveAll(storage.SanitizePath(f.rootDirectory, path.Join(consts.BlobUploadParts, uploadID)))
	if err != nil {
		return err
	}
	return os.Rename(fake, rPath)
}

// AbortUpload aborts a multipart upload.
func (f *fs) AbortUpload(ctx context.Context, _ string, uploadID string) error {
	return os.RemoveAll(storage.SanitizePath(f.rootDirectory, path.Join(consts.BlobUploadParts, uploadID)))
}

// Upload upload a file to the given path.
func (f *fs) Upload(ctx context.Context, p string, body io.Reader) error {
	if body == nil {
		return fmt.Errorf("body is nil")
	}

	temp, err := os.CreateTemp("", consts.AppName)
	if err != nil {
		return err
	}
	defer func() {
		e := os.Remove(temp.Name())
		if e != nil {
			slog.Error("remove temp file failed", "err", e)
		}
	}()
	_, err = io.Copy(temp, body)
	if err != nil {
		return err
	}

	err = temp.Close()
	if err != nil {
		return err
	}

	temp, err = os.Open(temp.Name())
	if err != nil {
		return err
	}
	defer func() {
		err = temp.Close()
		if err != nil {
			slog.Error("close temp file failed", "err", err)
		}
	}()

	p = storage.SanitizePath(f.rootDirectory, p)
	if utils.IsExist(p) {
		e := os.RemoveAll(p)
		if e != nil {
			return e
		}
	}
	if !utils.IsDir(path.Dir(p)) {
		e := os.MkdirAll(path.Dir(p), 0755)
		if e != nil {
			return e
		}
	}
	fp, err := os.Create(p)
	if err != nil {
		return err
	}
	defer fp.Close() // nolint: errcheck
	_, err = io.Copy(fp, temp)
	if err != nil {
		return err
	}
	return nil
}

// Redirect get a temporary link
func (f *fs) Redirect(_ context.Context, _ string) (string, error) {
	panic("Never implement")
}

// Ping verifies the filesystem backend is reachable by writing and removing a
// small probe object. A random suffix avoids collisions under multi-replica
// setups. Delete failures are intentionally ignored.
func (f *fs) Ping(ctx context.Context) error {
	probePath := path.Join(".sigma-healthz", "probe-"+uuid.NewV7String())
	err := f.Upload(ctx, probePath, strings.NewReader("ping"))
	if err != nil {
		return fmt.Errorf("filesystem ping upload: %w", err)
	}
	_ = f.Delete(ctx, probePath)
	return nil
}
