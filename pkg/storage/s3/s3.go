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
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/spf13/viper"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/storage"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

type awss3 struct {
	S3            *s3.S3
	uploader      *s3manager.Uploader
	rootDirectory string
	bucket        string
}

func init() {
	utils.PanicIf(storage.RegisterDriverFactory(path.Base(reflect.TypeOf(factory{}).PkgPath()), &factory{}))
}

type factory struct{}

var _ storage.Factory = factory{}

func (f factory) New(config configs.Configuration) (storage.StorageDriver, error) {
	endpoint := config.Storage.S3.Endpoint
	region := config.Storage.S3.Region
	ak := config.Storage.S3.Ak
	sk := config.Storage.S3.Sk
	bucket := config.Storage.S3.Bucket
	forcePathStyle := config.Storage.S3.ForcePathStyle

	sess, err := session.NewSession(&aws.Config{
		Endpoint:         aws.String(endpoint),
		Region:           aws.String(region),
		S3ForcePathStyle: aws.Bool(forcePathStyle),
		Credentials:      credentials.NewStaticCredentials(ak, sk, ""),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create new session with aws config: %v", err)
	}
	return &awss3{
		S3:            s3.New(sess),
		uploader:      s3manager.NewUploader(sess),
		bucket:        bucket,
		rootDirectory: strings.TrimPrefix(viper.GetString("storage.rootDirectory"), "/"),
	}, nil
}

func (a *awss3) sanitizePath(p string) string {
	return strings.TrimPrefix(strings.TrimPrefix(path.Join(a.rootDirectory, p), "."), "/")
}

// Move moves a file from srcPath to dstPath.
func (a *awss3) Move(ctx context.Context, srcPath string, dstPath string) error {
	srcPath = a.sanitizePath(srcPath)
	dstPath = a.sanitizePath(dstPath)

	srcFile, err := a.S3.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(srcPath),
	})
	if err != nil {
		return fmt.Errorf("Head source path(%s) failed: %v", srcPath, err)
	}
	srcSize := ptr.To(srcFile.ContentLength)

	if srcSize <= storage.MultipartCopyThresholdSize {
		_, err := a.S3.CopyObject(&s3.CopyObjectInput{
			Bucket:     aws.String(a.bucket),
			Key:        aws.String(dstPath),
			CopySource: aws.String(path.Join(a.bucket, srcPath)),
		})
		if err != nil {
			return err
		}
		return nil
	}

	createResp, err := a.S3.CreateMultipartUpload(&s3.CreateMultipartUploadInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(dstPath),
	})
	if err != nil {
		return err
	}

	numParts := (srcSize + storage.MultipartCopyChunkSize - 1) / storage.MultipartCopyChunkSize
	completedParts := make([]*s3.CompletedPart, numParts)
	errChan := make(chan error, numParts)
	limiter := make(chan struct{}, storage.MultipartCopyMaxConcurrency)

	for i := range completedParts {
		i := int64(i)
		go func() {
			limiter <- struct{}{}
			firstByte := i * storage.MultipartCopyChunkSize
			lastByte := firstByte + storage.MultipartCopyChunkSize - 1
			if lastByte >= srcSize {
				lastByte = srcSize - 1
			}
			uploadResp, err := a.S3.UploadPartCopy(&s3.UploadPartCopyInput{
				Bucket:          aws.String(a.bucket),
				CopySource:      aws.String(path.Join(a.bucket, srcPath)),
				Key:             aws.String(dstPath),
				PartNumber:      aws.Int64(i + 1),
				UploadId:        createResp.UploadId,
				CopySourceRange: aws.String(fmt.Sprintf("bytes=%d-%d", firstByte, lastByte)),
			})
			if err == nil {
				completedParts[i] = &s3.CompletedPart{
					ETag:       uploadResp.CopyPartResult.ETag,
					PartNumber: aws.Int64(i + 1),
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

	_, err = a.S3.CompleteMultipartUpload(&s3.CompleteMultipartUploadInput{
		Bucket:          aws.String(a.bucket),
		Key:             aws.String(dstPath),
		UploadId:        createResp.UploadId,
		MultipartUpload: &s3.CompletedMultipartUpload{Parts: completedParts},
	})
	return err
}

// Delete removes the object at the given path.
func (a *awss3) Delete(ctx context.Context, path string) error {
	path = storage.SanitizePath(a.rootDirectory, path)

	s3Objects := make([]*s3.ObjectIdentifier, 0, storage.MaxPaginationKeys)
	listObjectsInput := &s3.ListObjectsV2Input{
		Bucket: aws.String(a.bucket),
		Prefix: aws.String(path),
	}

	for {
		// list all the objects
		resp, err := a.S3.ListObjectsV2(listObjectsInput)

		// resp.Contents can only be empty on the first call
		// if there were no more results to return after the first call, resp.IsTruncated would have been false
		// and the loop would exit without recalling ListObjects
		if err != nil || len(resp.Contents) == 0 {
			return fmt.Errorf("failed to list objects: %w", err)
		}

		for _, key := range resp.Contents {
			// Skip if we encounter a key that is not a subpath (so that deleting "/a" does not delete "/ab").
			if len(*key.Key) > len(path) && (*key.Key)[len(path)] != '/' {
				continue
			}
			s3Objects = append(s3Objects, &s3.ObjectIdentifier{
				Key: key.Key,
			})
		}

		// Delete objects only if the list is not empty, otherwise S3 API returns a cryptic error
		if len(s3Objects) > 0 {
			// NOTE: according to AWS docs https://docs.aws.amazon.com/AmazonS3/latest/API/API_ListObjectsV2.html
			// by default the response returns up to 1,000 key names. The response _might_ contain fewer keys but it will never contain more.
			// 10000 keys is coincidentally (?) also the max number of keys that can be deleted in a single Delete operation, so we'll just smack
			// Delete here straight away and reset the object slice when successful.
			resp, err := a.S3.DeleteObjects(&s3.DeleteObjectsInput{
				Bucket: aws.String(a.bucket),
				Delete: &s3.Delete{
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
				errs := make([]error, 0, len(resp.Errors))
				for _, err := range resp.Errors {
					errs = append(errs, errors.New(err.String()))
				}
				return fmt.Errorf("failed to delete objects: %w", errs[0])
			}
		}
		// NOTE: we don't want to reallocate
		// the slice so we simply "reset" it
		s3Objects = s3Objects[:0]

		// resp.Contents must have at least one element or we would have returned not found
		listObjectsInput.StartAfter = resp.Contents[len(resp.Contents)-1].Key

		// from the s3 api docs, IsTruncated "specifies whether (true) or not (false) all of the results were returned"
		// if everything has been returned, break
		if resp.IsTruncated == nil || !*resp.IsTruncated {
			break
		}
	}

	return nil
}

// Reader returns a reader for the given path.
func (a *awss3) Reader(ctx context.Context, path string, offset int64) (io.ReadCloser, error) {
	resp, err := a.S3.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(a.sanitizePath(path)),
		Range:  aws.String("bytes=" + strconv.FormatInt(offset, 10) + "-"),
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == s3.ErrCodeNoSuchKey {
				fmt.Println(254, awsErr.Error())
				return nil, os.ErrNotExist
			}
			return nil, awsErr
		}
		return nil, err
	}
	return resp.Body, err
}

// CreateUploadID creates a new upload ID.
func (a *awss3) CreateUploadID(ctx context.Context, path string) (string, error) {
	resp, err := a.S3.CreateMultipartUploadWithContext(ctx, &s3.CreateMultipartUploadInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(a.sanitizePath(path)),
	})
	if err != nil {
		return "", err
	}
	return aws.StringValue(resp.UploadId), nil
}

// UploadPart uploads a part of an object.
func (a *awss3) UploadPart(ctx context.Context, path, uploadID string, partNumber int64, body io.Reader) (string, error) {
	_, err := a.uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(a.sanitizePath(path)),
		Body:   body,
	})
	if err != nil {
		return "", err
	}
	resp, err := a.S3.UploadPartCopyWithContext(ctx, &s3.UploadPartCopyInput{
		Bucket:     aws.String(a.bucket),
		Key:        aws.String(a.sanitizePath(path)),
		UploadId:   aws.String(uploadID),
		PartNumber: aws.Int64(partNumber),
		CopySource: aws.String(a.bucket + "/" + a.sanitizePath(path)),
	})
	if err != nil {
		return "", err
	}
	return aws.StringValue(resp.CopyPartResult.ETag), nil
}

// CommitUpload commits an upload.
func (a *awss3) CommitUpload(ctx context.Context, path, uploadID string, parts []string) error {
	completedParts := make([]*s3.CompletedPart, len(parts))
	for i, part := range parts {
		completedParts[i] = &s3.CompletedPart{
			ETag:       aws.String(part),
			PartNumber: aws.Int64(int64(i + 1)),
		}
	}
	_, err := a.S3.CompleteMultipartUploadWithContext(ctx, &s3.CompleteMultipartUploadInput{
		Bucket:          aws.String(a.bucket),
		Key:             aws.String(a.sanitizePath(path)),
		UploadId:        aws.String(uploadID),
		MultipartUpload: &s3.CompletedMultipartUpload{Parts: completedParts},
	})
	return err
}

// AbortUpload aborts an upload.
func (a *awss3) AbortUpload(ctx context.Context, path string, uploadID string) error {
	_, err := a.S3.AbortMultipartUploadWithContext(ctx, &s3.AbortMultipartUploadInput{
		Bucket:   aws.String(a.bucket),
		Key:      aws.String(storage.SanitizePath(a.rootDirectory, path)),
		UploadId: aws.String(uploadID),
	})
	return err
}

// Upload uploads a file to the given path.
func (a *awss3) Upload(ctx context.Context, path string, body io.Reader) error {
	_, err := a.uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(storage.SanitizePath(a.rootDirectory, path)),
		Body:   body,
	})
	return err
}

// Redirect get a temporary link
func (a *awss3) Redirect(ctx context.Context, path string) (string, error) {
	req, _ := a.S3.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(a.sanitizePath(path)),
	})
	return req.Presign(consts.ObsPresignMaxTtl)
}
