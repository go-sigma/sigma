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

package s3

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/storage"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

type awss3 struct {
	client        *s3.Client
	rootDirectory string
	bucket        string
}

func init() {
	utils.PanicIf(storage.RegisterDriverFactory(enums.StorageTypeS3, &factory{}))
}

type factory struct{}

var _ storage.Factory = factory{}

func (f factory) New(cfg configs.Configuration) (storage.StorageDriver, error) {
	c, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(cfg.Storage.S3.Region),
		config.WithCredentialsProvider(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     cfg.Storage.S3.Ak,
				SecretAccessKey: cfg.Storage.S3.Sk,
			}, nil
		})),
	)
	if err != nil {
		return nil, fmt.Errorf("new s3 config failed: %v", err)
	}
	return &awss3{
		client: s3.NewFromConfig(c, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.Storage.S3.Endpoint)
			o.UsePathStyle = cfg.Storage.S3.ForcePathStyle
		}),
		bucket:        cfg.Storage.S3.Bucket,
		rootDirectory: strings.TrimPrefix(cfg.Storage.RootDirectory, "/"),
	}, nil
}

func (a *awss3) sanitizePath(p string) string {
	return strings.TrimPrefix(strings.TrimPrefix(path.Join(a.rootDirectory, p), "."), "/")
}

// Move moves a file from srcPath to dstPath.
func (a *awss3) Move(ctx context.Context, srcPath string, dstPath string) error {
	srcPath = a.sanitizePath(srcPath)
	dstPath = a.sanitizePath(dstPath)

	srcFile, err := a.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(srcPath),
	})
	if err != nil {
		return fmt.Errorf("Head source path(%s) failed: %v", srcPath, err)
	}
	srcSize := ptr.To(srcFile.ContentLength)

	if srcSize <= storage.MultipartCopyThresholdSize {
		_, err := a.client.CopyObject(ctx, &s3.CopyObjectInput{
			Bucket:     aws.String(a.bucket),
			Key:        aws.String(dstPath),
			CopySource: aws.String(path.Join(a.bucket, srcPath)),
		})
		if err != nil {
			return err
		}
		return nil
	}

	createResp, err := a.client.CreateMultipartUpload(ctx, &s3.CreateMultipartUploadInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(dstPath),
	})
	if err != nil {
		return err
	}

	numParts := (srcSize + storage.MultipartCopyChunkSize - 1) / storage.MultipartCopyChunkSize
	completedParts := make([]types.CompletedPart, numParts)
	errChan := make(chan error, numParts)
	limiter := make(chan struct{}, storage.MultipartCopyMaxConcurrency)

	for i := range completedParts {
		i := int32(i) // nolint: gosec
		go func() {
			limiter <- struct{}{}
			firstByte := int64(i) * storage.MultipartCopyChunkSize
			lastByte := firstByte + storage.MultipartCopyChunkSize - 1
			if lastByte >= srcSize {
				lastByte = srcSize - 1
			}
			uploadResp, err := a.client.UploadPartCopy(ctx, &s3.UploadPartCopyInput{
				Bucket:          aws.String(a.bucket),
				CopySource:      aws.String(path.Join(a.bucket, srcPath)),
				Key:             aws.String(dstPath),
				PartNumber:      aws.Int32(i + 1),
				UploadId:        createResp.UploadId,
				CopySourceRange: aws.String(fmt.Sprintf("bytes=%d-%d", firstByte, lastByte)),
			})
			if err == nil {
				completedParts[i] = types.CompletedPart{
					ETag:       uploadResp.CopyPartResult.ETag,
					PartNumber: aws.Int32(i + 1),
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

	_, err = a.client.CompleteMultipartUpload(ctx, &s3.CompleteMultipartUploadInput{
		Bucket:          aws.String(a.bucket),
		Key:             aws.String(dstPath),
		UploadId:        createResp.UploadId,
		MultipartUpload: &types.CompletedMultipartUpload{Parts: completedParts},
	})
	return err
}

// Delete removes the object at the given path.
func (a *awss3) Delete(ctx context.Context, path string) error {
	path = storage.SanitizePath(a.rootDirectory, path)

	paginator := s3.NewListObjectsV2Paginator(a.client, &s3.ListObjectsV2Input{
		Bucket:  aws.String(a.bucket),
		Prefix:  aws.String(path),
		MaxKeys: aws.Int32(storage.MaxPaginationKeys),
	})

	var s3Objects = make([]types.ObjectIdentifier, 0, storage.MaxPaginationKeys)
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, key := range output.Contents {
			// Skip if we encounter a key that is not a subpath (so that deleting "/a" does not delete "/ab").
			if len(*key.Key) > len(path) && (*key.Key)[len(path)] != '/' {
				continue
			}
			s3Objects = append(s3Objects, types.ObjectIdentifier{
				Key: key.Key,
			})
		}
		if len(s3Objects) > 0 {
			resp, err := a.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
				Bucket: aws.String(a.bucket),
				Delete: &types.Delete{
					Objects: s3Objects,
					Quiet:   aws.Bool(false),
				},
			})
			if err != nil {
				return err
			}
			if len(resp.Errors) > 0 {
				// NOTE: AWS SDK s3.Error does not implement error interface which
				// is pretty intensely sad, so we have to do away with this for now.
				errs := make([]types.Error, 0, len(resp.Errors))
				// for _, err := range resp.Errors {
				errs = append(errs, resp.Errors...)
				// }
				err := errs[0]
				return fmt.Errorf("failed to delete objects: %s, code: %s", ptr.To(err.Message), ptr.To(err.Code))
			}
		}
		s3Objects = s3Objects[:0]
	}

	return nil
}

// Reader returns a reader for the given path.
func (a *awss3) Reader(ctx context.Context, path string) (io.ReadCloser, error) {
	resp, err := a.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(a.sanitizePath(path)),
	})
	if err != nil {
		return nil, err
	}
	return resp.Body, err
}

// CreateUploadID creates a new upload ID.
func (a *awss3) CreateUploadID(ctx context.Context, path string) (string, error) {
	resp, err := a.client.CreateMultipartUpload(ctx, &s3.CreateMultipartUploadInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(a.sanitizePath(path)),
	})
	if err != nil {
		return "", err
	}
	return ptr.To(resp.UploadId), nil
}

// UploadPart uploads a part of an object.
func (a *awss3) UploadPart(ctx context.Context, path, uploadID string, partNumber int32, body io.Reader) (string, error) {
	file, err := os.CreateTemp("", "s3")
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

	fd, err := os.Open(file.Name())
	if err != nil {
		return "", err
	}
	defer fd.Close() // nolint: errcheck

	resp, err := a.client.UploadPart(ctx, &s3.UploadPartInput{
		Bucket:     aws.String(a.bucket),
		Key:        aws.String(a.sanitizePath(path)),
		UploadId:   aws.String(uploadID),
		PartNumber: aws.Int32(partNumber),
		Body:       fd,
	})
	if err != nil {
		return "", err
	}
	return ptr.To(resp.ETag), nil
}

// CommitUpload commits an upload.
func (a *awss3) CommitUpload(ctx context.Context, path, uploadID string, parts []string) error {
	completedParts := make([]types.CompletedPart, len(parts))
	for i, part := range parts {
		completedParts[i] = types.CompletedPart{
			ETag:       aws.String(part),
			PartNumber: aws.Int32(int32(i + 1)), // nolint: gosec
		}
	}
	_, err := a.client.CompleteMultipartUpload(ctx, &s3.CompleteMultipartUploadInput{
		Bucket:          aws.String(a.bucket),
		Key:             aws.String(a.sanitizePath(path)),
		UploadId:        aws.String(uploadID),
		MultipartUpload: &types.CompletedMultipartUpload{Parts: completedParts},
	})
	return err
}

// AbortUpload aborts an upload.
func (a *awss3) AbortUpload(ctx context.Context, path string, uploadID string) error {
	_, err := a.client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
		Bucket:   aws.String(a.bucket),
		Key:      aws.String(storage.SanitizePath(a.rootDirectory, path)),
		UploadId: aws.String(uploadID),
	})
	return err
}

// Upload uploads a file to the given path.
func (a *awss3) Upload(ctx context.Context, path string, body io.Reader) error {
	_, err := a.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(storage.SanitizePath(a.rootDirectory, path)),
		Body:   body,
	})
	return err
}

// Redirect get a temporary link
func (a *awss3) Redirect(ctx context.Context, path string) (string, error) {
	req, err := s3.NewPresignClient(a.client).PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(a.sanitizePath(path)),
	}, s3.WithPresignExpires(consts.ObsPresignMaxTtl))
	if err != nil {
		return "", err
	}
	return req.URL, nil
}
