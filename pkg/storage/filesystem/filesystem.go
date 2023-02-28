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

package filesystem

import (
	"context"
	"io"
	"os"
	"path"
	"strings"

	gonanoid "github.com/matoous/go-nanoid"
	"github.com/spf13/viper"
	"github.com/ximager/ximager/pkg/storage"
)

const (
	// name is the name of the filesystem storage driver
	name = "filesystem"
)

// fs is the filesystem storage driver
type fs struct {
	rootDirectory string
}

// New returns a new filesystem storage driver
func New() storage.StorageDriver {
	return &fs{rootDirectory: viper.GetString("storage.rootDirectory")}
}

func (f *fs) sanitizePath(p string) string {
	return strings.Trim(strings.TrimPrefix(path.Join(f.rootDirectory, p), "."), "/")
}

// Name returns the name of the filesystem storage driver
func (f *fs) Name() string {
	return name
}

// Stat returns the file info for the given path
func (f *fs) Stat(ctx context.Context, path string) (storage.FileInfo, error) {
	return os.Stat(path)
}

// Move moves a file from sourcePath to destPath
func (f *fs) Move(ctx context.Context, sourcePath string, destPath string) error {
	return os.Rename(sourcePath, destPath)
}

// Delete deletes a file at the given path
func (f *fs) Delete(ctx context.Context, path string) error {
	return os.RemoveAll(path)
}

// Reader returns a reader for the file at the given path
func (f *fs) Reader(ctx context.Context, path string, offset int64) (io.ReadCloser, error) {
	fp, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	seekPos, err := fp.Seek(offset, io.SeekStart)
	if err != nil {
		fp.Close() // nolint: errcheck
		return nil, err
	} else if seekPos < offset {
		fp.Close() // nolint: errcheck
		return nil, err
	}
	return fp, nil
}

// CreateUploadID creates a new multipart upload and returns an
// opaque upload ID.
func (f *fs) CreateUploadID(ctx context.Context, rPath string) (string, error) {
	rPath = f.sanitizePath(rPath)
	uploadID := gonanoid.MustGenerate("abcdefghijklmnopqrstuvwxyz0123456789", 32)
	err := os.MkdirAll(path.Join(rPath, uploadID), 0755)
	if err != nil {
		return "", err
	}
	return uploadID, nil
}

// WritePart writes a part of a multipart upload.
func (f *fs) UploadPart(ctx context.Context, rPath, uploadID string, partNumber int64, body io.Reader) (string, error) {
	rPath = f.sanitizePath(rPath)
	eTag := gonanoid.MustGenerate("abcdefghijklmnopqrstuvwxyz0123456789", 32)
	fp, err := os.OpenFile(path.Join(rPath, uploadID, eTag), os.O_CREATE|os.O_WRONLY, 0644)
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
func (f *fs) CommitUpload(ctx context.Context, rPath string, uploadID string, parts []string) error {
	rPath = f.sanitizePath(rPath)
	fake := path.Join(rPath + uploadID + "fake")
	fp, err := os.OpenFile(fake, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	for _, part := range parts {
		partPath := path.Join(rPath, uploadID, part)
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
		return nil
	}
	err = os.RemoveAll(rPath)
	if err != nil {
		return nil
	}
	return os.Rename(fake, rPath)
}

// AbortUpload aborts a multipart upload.
func (f *fs) AbortUpload(ctx context.Context, rPath string, uploadID string) error {
	return os.RemoveAll(path.Join(rPath, uploadID))
}

// Upload upload a file to the given path.
func (f *fs) Upload(ctx context.Context, path string, body io.Reader) error {
	path = f.sanitizePath(path)
	fp, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	_, err = io.Copy(fp, body)
	if err != nil {
		fp.Close() // nolint: errcheck
		return err
	}
	return fp.Close()
}
