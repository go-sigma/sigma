package s3

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/spf13/viper"

	"github.com/ximager/ximager/pkg/storage"
)

const (
	name = "s3"
)

type awss3 struct {
	S3            *s3.S3
	uploader      *s3manager.Uploader
	rootDirectory string
	bucket        string
}

func init() {
	err := storage.RegisterDriverFactory(name, &factory{})
	if err != nil {
		panic(fmt.Sprintf("fail to register driver factory: %v", err))
	}
}

type factory struct{}

var _ storage.Factory = factory{}

func (f factory) New() (storage.StorageDriver, error) {
	endpoint := viper.GetString("storage.s3.endpoint")
	region := viper.GetString("storage.s3.region")
	ak := viper.GetString("storage.s3.ak")
	sk := viper.GetString("storage.s3.sk")
	bucket := viper.GetString("storage.s3.bucket")
	forcePathStyle := viper.GetBool("storage.s3.forcePathStyle")

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
		rootDirectory: viper.GetString("storage.rootDirectory"),
	}, nil
}

// New returns a new instance of the s3 storage driver.
func (a *awss3) Name() string {
	return name
}

type fileInfo struct {
	name    string
	size    int64
	modTime time.Time
	isDir   bool
}

var _ storage.FileInfo = fileInfo{}

func (f fileInfo) Name() string {
	return f.name
}

func (f fileInfo) Size() int64 {
	return f.size
}

func (f fileInfo) ModTime() time.Time {
	return f.modTime
}

func (f fileInfo) IsDir() bool {
	return f.isDir
}

// Stat returns the file info for the given path.
func (a *awss3) Stat(ctx context.Context, path string) (storage.FileInfo, error) {
	path = a.sanitizePath(path)
	resp, err := a.S3.ListObjectsV2WithContext(ctx, &s3.ListObjectsV2Input{
		Bucket:  aws.String(a.bucket),
		Prefix:  aws.String(path),
		MaxKeys: aws.Int64(1),
	})
	if err != nil {
		return nil, err
	}

	fi := fileInfo{name: path}
	if len(resp.Contents) == 1 {
		if *resp.Contents[0].Key != path {
			fi.isDir = true
		} else {
			fi.isDir = false
			fi.size = *resp.Contents[0].Size
			fi.modTime = *resp.Contents[0].LastModified
		}
	}
	return fi, nil
}

const (
	multipartCopyThresholdSize  = 32 << 20 // 32MB
	multipartCopyChunkSize      = 32 << 20 // 32MB
	multipartCopyMaxConcurrency = 100      // 100 goroutines
	maxPaginationKeys           = 1000     // 1000 keys
)

func (a *awss3) sanitizePath(p string) string {
	return strings.TrimPrefix(strings.TrimPrefix(path.Join(a.rootDirectory, p), "."), "/")
}

func (a *awss3) Move(ctx context.Context, sourcePath string, destPath string) error {
	sourcePath = a.sanitizePath(sourcePath)
	destPath = a.sanitizePath(destPath)

	fileInfo, err := a.Stat(ctx, sourcePath)
	if err != nil {
		return err
	}

	if fileInfo.Size() <= multipartCopyThresholdSize {
		_, err := a.S3.CopyObject(&s3.CopyObjectInput{
			Bucket:     aws.String(a.bucket),
			Key:        aws.String(destPath),
			CopySource: aws.String(path.Join(a.bucket, sourcePath)),
		})
		if err != nil {
			return err
		}
		return nil
	}

	createResp, err := a.S3.CreateMultipartUpload(&s3.CreateMultipartUploadInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(a.sanitizePath(destPath)),
	})
	if err != nil {
		return err
	}

	numParts := (fileInfo.Size() + multipartCopyChunkSize - 1) / multipartCopyChunkSize
	completedParts := make([]*s3.CompletedPart, numParts)
	errChan := make(chan error, numParts)
	limiter := make(chan struct{}, multipartCopyMaxConcurrency)

	for i := range completedParts {
		i := int64(i)
		go func() {
			limiter <- struct{}{}
			firstByte := i * multipartCopyChunkSize
			lastByte := firstByte + multipartCopyChunkSize - 1
			if lastByte >= fileInfo.Size() {
				lastByte = fileInfo.Size() - 1
			}
			uploadResp, err := a.S3.UploadPartCopy(&s3.UploadPartCopyInput{
				Bucket:          aws.String(a.bucket),
				CopySource:      aws.String(path.Join(a.bucket, sourcePath)),
				Key:             aws.String(destPath),
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
		Key:             aws.String(a.sanitizePath(destPath)),
		UploadId:        createResp.UploadId,
		MultipartUpload: &s3.CompletedMultipartUpload{Parts: completedParts},
	})
	return err
}

func (a *awss3) Delete(ctx context.Context, path string) error {
	path = a.sanitizePath(path)

	s3Objects := make([]*s3.ObjectIdentifier, 0, maxPaginationKeys)
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

func (a *awss3) Reader(ctx context.Context, path string, offset int64) (io.ReadCloser, error) {
	resp, err := a.S3.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(a.sanitizePath(path)),
		Range:  aws.String("bytes=" + strconv.FormatInt(offset, 10) + "-"),
	})
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

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
		Key:      aws.String(a.sanitizePath(path)),
		UploadId: aws.String(uploadID),
	})
	return err
}

// Upload uploads a file to the given path.
func (a *awss3) Upload(ctx context.Context, path string, body io.Reader) error {
	_, err := a.uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(a.sanitizePath(path)),
		Body:   body,
	})
	if err != nil {
		return err
	}
	return nil
}
