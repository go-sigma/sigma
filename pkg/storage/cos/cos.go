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

package cos

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	gonanoid "github.com/matoous/go-nanoid"
	"github.com/tencentyun/cos-go-sdk-v5"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/storage"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
)

func init() {
	utils.PanicIf(storage.RegisterDriverFactory(enums.StorageTypeCos, &factory{}))
}

type factory struct{}

var _ storage.Factory = factory{}

// New ...
func (f factory) New(config configs.Configuration) (storage.StorageDriver, error) {
	u, err := url.Parse(config.Storage.Cos.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("Config [storage.cos.endpoint] is invalid")
	}

	c := cos.NewClient(&cos.BaseURL{BucketURL: u}, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  config.Storage.Cos.Ak,
			SecretKey: config.Storage.Cos.Sk,
		},
	})

	return &tencentcos{
		client:        c,
		domain:        u.Host,
		rootDirectory: strings.TrimPrefix(config.Storage.RootDirectory, "/"),
	}, nil
}

type tencentcos struct {
	client        *cos.Client
	domain        string
	rootDirectory string
}

// Move moves a file from srcPath to dstPath.
func (t *tencentcos) Move(ctx context.Context, srcPath, dstPath string) (err error) {
	srcPath = storage.SanitizePath(t.rootDirectory, srcPath)
	dstPath = storage.SanitizePath(t.rootDirectory, dstPath)

	srcFile, err := t.client.Object.Head(ctx, srcPath, nil)
	if err != nil {
		return err
	}
	srcSize := srcFile.ContentLength

	if srcSize < storage.MultipartCopyThresholdSize {
		_, _, err := t.client.Object.Copy(ctx, dstPath,
			fmt.Sprintf("%s/%s", t.domain, srcPath), &cos.ObjectCopyOptions{})
		return err
	}

	uploadRes, _, err := t.client.Object.InitiateMultipartUpload(ctx, dstPath, nil)
	if err != nil {
		return err
	}
	uploadID := uploadRes.UploadID

	numParts := (srcSize + storage.MultipartCopyChunkSize - 1) / storage.MultipartCopyChunkSize
	completedParts := make([]cos.Object, numParts)
	errChan := make(chan error, numParts)
	limiter := make(chan struct{}, storage.MultipartCopyMaxConcurrency)

	for i := range completedParts {
		i := i
		go func() {
			limiter <- struct{}{}
			firstByte := i * storage.MultipartCopyChunkSize
			lastByte := firstByte + storage.MultipartCopyChunkSize - 1
			if int64(lastByte) >= srcSize {
				lastByte = int(srcSize - 1)
			}
			partResp, _, err := t.client.Object.CopyPart(ctx, dstPath, uploadID, i+1,
				fmt.Sprintf("%s/%s", t.domain, srcPath), &cos.ObjectCopyPartOptions{
					XCosCopySourceRange: fmt.Sprintf("bytes=%d-%d", firstByte, lastByte),
				})
			if err == nil {
				completedParts[i] = cos.Object{
					ETag:       partResp.ETag,
					PartNumber: i + 1,
				}
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

	_, _, err = t.client.Object.CompleteMultipartUpload(ctx, dstPath, uploadID, &cos.CompleteMultipartUploadOptions{
		Parts: completedParts,
	})
	return err
}

// Delete removes the object at the given path.
// Note: if you delete 'test' then the following file and dir will be deleted:
// 'test/file', 'test', 'test/'
func (t *tencentcos) Delete(ctx context.Context, path string) error {
	path = storage.SanitizePath(t.rootDirectory, path)

	objects := make([]cos.Object, 0, storage.MaxPaginationKeys)
	var marker string
	for {
		opt := &cos.BucketGetOptions{
			Prefix:  path,
			Marker:  marker,
			MaxKeys: storage.MaxPaginationKeys,
		}
		resp, _, err := t.client.Bucket.Get(ctx, opt)
		if err != nil {
			return fmt.Errorf("List objects failed: %v", err)
		}
		if len(resp.Contents) == 0 {
			break
		}

		for _, obj := range resp.Contents {
			if len(obj.Key) > len(path) && obj.Key[len(path)] != '/' {
				continue
			}
			objects = append(objects, cos.Object{
				Key: obj.Key,
			})
		}

		if len(objects) > 0 {
			resp, _, err := t.client.Object.DeleteMulti(ctx, &cos.ObjectDeleteMultiOptions{
				Quiet:   false,
				Objects: objects,
			})
			if err != nil {
				return err
			}

			if len(resp.DeletedObjects) != len(objects) {
				errs := make([]error, 0, len(resp.Errors))
				for _, err := range resp.Errors {
					errs = append(errs, fmt.Errorf(err.Message))
				}
				return fmt.Errorf("Delete objects failed: %+v", errs)
			}
		}
		objects = objects[:0]
		marker = resp.NextMarker
		if !resp.IsTruncated {
			break
		}
	}
	return nil
}

// Reader returns a reader for the given path.
func (t *tencentcos) Reader(ctx context.Context, path string) (io.ReadCloser, error) {
	opt := &cos.ObjectGetOptions{}
	resp, err := t.client.Object.Get(ctx, path, opt)
	if err != nil {
		return nil, fmt.Errorf("Get object failed: %v", err)
	}
	return resp.Body, nil
}

// CreateUploadID creates a new upload ID.
func (t *tencentcos) CreateUploadID(ctx context.Context, path string) (string, error) {
	resp, _, err := t.client.Object.InitiateMultipartUpload(ctx, path, &cos.InitiateMultipartUploadOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to create multipart upload: %v", err)
	}
	return resp.UploadID, nil
}

// UploadPart WritePart writes a part of a multipart upload.
func (t *tencentcos) UploadPart(ctx context.Context, rPath, uploadID string, partNumber int64, body io.Reader) (string, error) {
	partPath := path.Join(storage.SanitizePath(t.rootDirectory, consts.BlobUploadParts), rPath, fmt.Sprintf("%s-%s", uploadID, gonanoid.MustGenerate(consts.Alphanum, 6)))
	_, err := t.client.Object.Put(ctx, partPath, body, nil)
	if err != nil {
		return "", fmt.Errorf("Upload part failed: %v", err)
	}
	res, _, err := t.client.Object.CopyPart(ctx, rPath, uploadID, int(partNumber), fmt.Sprintf("%s/%s", t.domain, partPath), nil)
	if err != nil {
		return "", fmt.Errorf("Upload part failed: %v", err)
	}
	err = t.Delete(ctx, partPath)
	if err != nil {
		return "", fmt.Errorf("Delete the part file failed: %v", err)
	}
	return res.ETag, nil
}

// CommitUpload commits a multipart upload.
func (t *tencentcos) CommitUpload(ctx context.Context, path, uploadID string, parts []string) error {
	completeParts := make([]cos.Object, len(parts))
	for i, part := range parts {
		completeParts[i] = cos.Object{
			ETag:       part,
			PartNumber: i + 1,
		}
	}
	_, _, err := t.client.Object.CompleteMultipartUpload(ctx, path, uploadID, &cos.CompleteMultipartUploadOptions{Parts: completeParts})
	return err
}

// AbortUpload aborts a multipart upload.
func (t *tencentcos) AbortUpload(ctx context.Context, path string, uploadID string) error {
	_, err := t.client.Object.AbortMultipartUpload(ctx, path, uploadID)
	return err
}

// Upload upload a file to the given path.
func (t *tencentcos) Upload(ctx context.Context, path string, body io.Reader) error {
	_, err := t.client.Object.Put(ctx, path, body, nil)
	return err
}

// Redirect get a temporary link
func (t *tencentcos) Redirect(ctx context.Context, path string) (string, error) {
	opt := &cos.ObjectGetOptions{}
	url, err := t.client.Object.GetPresignedURL2(ctx, http.MethodGet, path, consts.ObsPresignMaxTtl, opt)
	if err != nil {
		return "", fmt.Errorf("Get object failed: %v", err)
	}
	return url.String(), nil
}
