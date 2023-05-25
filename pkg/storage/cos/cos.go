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

package cos

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"reflect"
	"time"

	"github.com/spf13/viper"
	cos "github.com/tencentyun/cos-go-sdk-v5"

	"github.com/ximager/ximager/pkg/storage"
	"github.com/ximager/ximager/pkg/utils"
)

func init() {
	utils.PanicIf(storage.RegisterDriverFactory(path.Base(reflect.TypeOf(factory{}).PkgPath()), &factory{}))
}

type factory struct{}

var _ storage.Factory = factory{}

func (f factory) New() (storage.StorageDriver, error) {
	endpoint := viper.GetString("storage.cos.endpoint")
	ak := viper.GetString("storage.cos.ak")
	sk := viper.GetString("storage.cos.sk")

	u, _ := url.Parse(endpoint)
	b := &cos.BaseURL{BucketURL: u}
	c := cos.NewClient(b, &http.Client{
		Timeout: 100 * time.Second,
		Transport: &cos.AuthorizationTransport{
			SecretID:  ak,
			SecretKey: sk,
		},
	})

	return &tencentcos{
		client: c,
	}, nil
}

type tencentcos struct {
	client *cos.Client
}

// Stat retrieves the FileInfo for the given path, including the current
// size in bytes and the creation time.
func (t *tencentcos) Stat(ctx context.Context, path string) (storage.FileInfo, error) {
	return os.Stat(path)
}

// Move moves an object stored at sourcePath to destPath, removing the
// original object.
// Note: This may be no more efficient than a copy followed by a delete for
// many implementations.
func (t *tencentcos) Move(ctx context.Context, sourcePath string, destPath string) error {
	return nil
}

// Delete recursively deletes all objects stored at "path" and its subpaths.
func (t *tencentcos) Delete(ctx context.Context, path string) error {
	return nil
}

// Reader retrieves an io.ReadCloser for the content stored at "path"
// with a given byte offset.
// May be used to resume reading a stream by providing a nonzero offset.
func (t *tencentcos) Reader(ctx context.Context, path string, offset int64) (io.ReadCloser, error) {
	f, _ := os.Create("test")
	return f, nil
}

// CreateUploadID creates a new multipart upload and returns an
// opaque upload ID.
func (t *tencentcos) CreateUploadID(ctx context.Context, path string) (string, error) {
	return "", nil
}

// WritePart writes a part of a multipart upload.
func (t *tencentcos) UploadPart(ctx context.Context, path, uploadID string, partNumber int64, body io.Reader) (string, error) {
	return "", nil
}

// CommitUpload commits a multipart upload.
func (t *tencentcos) CommitUpload(ctx context.Context, path string, uploadID string, parts []string) error {
	return nil
}

// AbortUpload aborts a multipart upload.
func (t *tencentcos) AbortUpload(ctx context.Context, path string, uploadID string) error {
	return nil
}

// Upload upload a file to the given path.
func (t *tencentcos) Upload(ctx context.Context, path string, body io.Reader) error {
	return nil
}
