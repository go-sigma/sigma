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

package oss

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/storage"
	"github.com/go-sigma/sigma/pkg/utils"
)

func init() {
	utils.PanicIf(storage.RegisterDriverFactory(path.Base(reflect.TypeOf(factory{}).PkgPath()), &factory{}))
}

type factory struct{}

var _ storage.Factory = factory{}

// New ...
func (f factory) New(config configs.Configuration) (storage.StorageDriver, error) {
	client, err := oss.New(config.Storage.Oss.Endpoint, config.Storage.Oss.Ak, config.Storage.Oss.Sk,
		oss.ForcePathStyle(config.Storage.Oss.ForcePathStyle))
	if err != nil {
		return nil, err
	}
	bucket, err := client.Bucket(config.Storage.Oss.Bucket)
	if err != nil {
		return nil, err
	}

	return &alioss{
		client:        bucket,
		rootDirectory: strings.TrimPrefix(config.Storage.RootDirectory, "/"),
		bucket:        config.Storage.Oss.Bucket,
	}, nil
}

type alioss struct {
	client        *oss.Bucket
	rootDirectory string
	bucket        string
}

func (a *alioss) sanitizePath(p string) string {
	return strings.TrimPrefix(strings.TrimPrefix(path.Join(a.rootDirectory, p), "."), "/")
}

// Move moves an object stored at sourcePath to destPath, removing the
// original object.
// Note: This may be no more efficient than a copy followed by a delete for
// many implementations.
func (a *alioss) Move(ctx context.Context, srcPath string, dstPath string) error {
	srcPath = a.sanitizePath(srcPath)
	dstPath = a.sanitizePath(dstPath)

	header, err := a.client.GetObjectMeta(srcPath)
	if err != nil {
		return err
	}

	srcSize, err := strconv.ParseInt(header.Get(textproto.CanonicalMIMEHeaderKey("Content-Length")), 10, 64)
	if err != nil {
		log.Error().Err(err).Interface("Headers", header).Msg("Get object header failed")
		return fmt.Errorf("Convert header content-length to int failed: %v", err)
	}

	if srcSize <= storage.MultipartCopyThresholdSize {
		_, err := a.client.CopyObject(srcPath, dstPath)
		if err != nil {
			return err
		}
		return nil
	}

	createResp, err := a.client.InitiateMultipartUpload(dstPath, oss.EnableMd5())
	if err != nil {
		return err
	}

	numParts := (srcSize + storage.MultipartCopyChunkSize - 1) / storage.MultipartCopyChunkSize
	completedParts := make([]oss.UploadPart, numParts)
	errChan := make(chan error, numParts)
	limiter := make(chan struct{}, storage.MultipartCopyMaxConcurrency)

	for i := range completedParts {
		i := i
		go func() {
			limiter <- struct{}{}
			firstByte := int64(i) * storage.MultipartCopyChunkSize
			lastByte := firstByte + storage.MultipartCopyChunkSize
			if lastByte >= srcSize {
				lastByte = srcSize
			}
			partSize := lastByte - firstByte
			uploadResp, err := a.client.UploadPartCopy(createResp, a.bucket, srcPath, firstByte, partSize, i+1)
			if err == nil {
				completedParts[i] = uploadResp
			}
			errChan <- err
			<-limiter
		}()
	}

	for range completedParts {
		err := <-errChan
		if err != nil {
			return err
		}
	}

	_, err = a.client.CompleteMultipartUpload(createResp, completedParts)
	return err
}

// Delete recursively deletes all objects stored at "path" and its subpaths.
func (a *alioss) Delete(ctx context.Context, path string) error {
	objects := make([]string, 0, storage.MaxPaginationKeys)
	for {
		result, err := a.client.ListObjectsV2(oss.MaxKeys(storage.MaxPaginationKeys),
			oss.Prefix(storage.SanitizePath(a.rootDirectory, path)))
		if err != nil {
			return err
		}

		for _, r := range result.CommonPrefixes {
			err := a.Delete(ctx, r)
			if err != nil {
				return err
			}
		}

		for _, r := range result.Objects {
			objects = append(objects, r.Key)
		}

		for _, o := range objects {
			err = a.client.DeleteObject(o)
			if err != nil {
				return err
			}
		}

		objects = objects[:0]

		if !result.IsTruncated {
			break
		}
	}
	return nil
}

// Reader retrieves an io.ReadCloser for the content stored at "path"
// with a given byte offset.
func (a *alioss) Reader(ctx context.Context, path string) (io.ReadCloser, error) {
	return a.client.GetObject(a.sanitizePath(path))
}

// CreateUploadID creates a new multipart upload and returns an
// opaque upload ID.
func (a *alioss) CreateUploadID(ctx context.Context, path string) (string, error) {
	result, err := a.client.InitiateMultipartUpload(a.sanitizePath(path))
	if err != nil {
		return "", err
	}
	return result.UploadID, nil
}

// WritePart writes a part of a multipart upload.
func (a *alioss) UploadPart(ctx context.Context, path, uploadID string, partNumber int64, body io.Reader) (string, error) {
	file, err := os.CreateTemp("", "alioss")
	if err != nil {
		return "", err
	}
	_, err = io.Copy(file, body)
	if err != nil {
		return "", err
	}
	err = file.Sync()
	if err != nil {
		return "", err
	}
	stat, err := file.Stat()
	if err != nil {
		return "", err
	}
	err = file.Close()
	if err != nil {
		return "", err
	}
	defer func() {
		err := os.Remove(file.Name())
		if err != nil {
			log.Error().Err(err).Str("File", file.Name()).Msg("Remove upload temp file failed")
		}
	}()

	result, err := a.client.UploadPartFromFile(oss.InitiateMultipartUploadResult{
		Bucket:   a.bucket,
		Key:      a.sanitizePath(path),
		UploadID: uploadID,
	}, file.Name(), 0, stat.Size(), int(partNumber))
	if err != nil {
		return "", err
	}
	return result.ETag, nil
}

// CommitUpload commits a multipart upload.
func (a *alioss) CommitUpload(ctx context.Context, path string, uploadID string, parts []string) error {
	var ps = make([]oss.UploadPart, 0, len(parts))
	for index, p := range parts {
		ps = append(ps, oss.UploadPart{
			PartNumber: index + 1,
			ETag:       p,
		})
	}
	_, err := a.client.CompleteMultipartUpload(oss.InitiateMultipartUploadResult{
		Bucket:   a.bucket,
		Key:      a.sanitizePath(path),
		UploadID: uploadID,
	}, ps)
	return err
}

// AbortUpload aborts a multipart upload.
func (a *alioss) AbortUpload(ctx context.Context, path string, uploadID string) error {
	return a.client.AbortMultipartUpload(oss.InitiateMultipartUploadResult{
		Bucket:   a.bucket,
		Key:      a.sanitizePath(path),
		UploadID: uploadID,
	})
}

// Upload upload a file to the given path.
func (a *alioss) Upload(ctx context.Context, path string, body io.Reader) error {
	file, err := os.CreateTemp("", "alioss")
	if err != nil {
		return err
	}
	_, err = io.Copy(file, body)
	if err != nil {
		return err
	}
	err = file.Close()
	if err != nil {
		return err
	}
	defer func() {
		err := os.Remove(file.Name())
		if err != nil {
			log.Error().Err(err).Str("File", file.Name()).Msg("Remove upload temp file failed")
		}
	}()
	return a.client.UploadFile(a.sanitizePath(path), file.Name(), storage.MultipartCopyChunkSize, oss.EnableMd5())
}

// Redirect get a temporary link
func (a *alioss) Redirect(ctx context.Context, path string) (string, error) {
	return a.client.SignURL(a.sanitizePath(path), http.MethodGet, int64(consts.ObsPresignMaxTtl.Seconds()))
}
