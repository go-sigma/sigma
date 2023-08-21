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
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/tencentyun/cos-go-sdk-v5"

	"github.com/go-sigma/sigma/pkg/configs"
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
	u, err := url.Parse(config.Storage.Cos.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("Config [storage.cos.endpoint] is invalid")
	}

	c := cos.NewClient(&cos.BaseURL{BucketURL: u}, &http.Client{
		Timeout: 3 * time.Second,
		Transport: &cos.AuthorizationTransport{
			SecretID:  config.Storage.Cos.Ak,
			SecretKey: config.Storage.Cos.Sk,
		},
	})

	return &tencentcos{
		client: c,
		domain: u.Host,
	}, nil
}

type tencentcos struct {
	client *cos.Client
	domain string
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

type CopyResults struct {
	PartNumber int
	Resp       *cos.Response
	Err        error
	Res        *cos.CopyPartResult
}

// Stat retrieves the FileInfo for the given path, including the current
// size in bytes and the creation time.
func (t *tencentcos) Stat(ctx context.Context, path string) (storage.FileInfo, error) {
	resp, err := t.client.Object.Head(ctx, path, &cos.ObjectHeadOptions{})
	if err != nil {
		if resp.Response.StatusCode == 404 {
			return nil, fmt.Errorf("%s", resp.Response.Status)
		}
		return nil, err
	}
	fi := fileInfo{name: path}
	lastModified, _ := time.Parse(time.RFC1123, resp.Header.Get("Last-Modified"))
	fi.size = resp.ContentLength
	fi.modTime = lastModified
	return fi, nil
}

func (t *tencentcos) Move(ctx context.Context, sourcePath, destPath string) (err error) {
	endpoint, _ := url.Parse(viper.GetString("storage.cos.endpoint"))
	sourceURL := fmt.Sprintf("%s/%s", endpoint.Host, sourcePath)
	sUrl := strings.SplitN(sourceURL, "/", 2)
	if len(sUrl) < 2 {
		err := fmt.Errorf("sourceURL format error: %s", sourceURL)
		return err
	}
	var resp *cos.Response
	resp, err = t.client.Object.Head(ctx, sUrl[1], nil)
	if err != nil {
		return err
	}

	totalBytes := resp.ContentLength
	u := fmt.Sprintf("%s/%s", sUrl[0], url.QueryEscape(sUrl[1]))
	opt := &cos.MultiCopyOptions{}

	chunks, partNum, err := SplitSizeIntoChunks(totalBytes, opt.PartSize*1024*1024)
	if err != nil {
		return err
	}
	if partNum == 0 || (totalBytes < storage.MultipartCopyThresholdSize) {
		_, _, err := t.client.Object.Copy(ctx, destPath, sourceURL, opt.OptCopy)
		return err
	}
	var uploadID string
	uploadRes, _, err := t.client.Object.InitiateMultipartUpload(ctx, destPath, nil)
	if err != nil {
		return err
	}
	uploadID = uploadRes.UploadID

	var poolSize int
	if opt.ThreadPoolSize > 0 {
		poolSize = opt.ThreadPoolSize
	} else {
		poolSize = 1
	}

	chjobs := make(chan *cos.CopyJobs, storage.MultipartCopyMaxConcurrency)
	chresults := make(chan *CopyResults)
	optcom := &cos.CompleteMultipartUploadOptions{}

	for w := 1; w <= poolSize; w++ {
		go func(chjobs <-chan *cos.CopyJobs, chresults chan<- *CopyResults) {
			for j := range chjobs {
				var copyres CopyResults
				j.Opt.XCosCopySourceRange = fmt.Sprintf("bytes=%d-%d", j.Chunk.OffSet, j.Chunk.OffSet+j.Chunk.Size-1)
				rt := j.RetryTimes
				for {
					res, resp, err := t.client.Object.CopyPart(ctx, j.Name, j.UploadId, j.Chunk.Number, j.Opt.XCosCopySource, j.Opt)
					copyres.PartNumber = j.Chunk.Number
					copyres.Resp = resp
					copyres.Res = res
					copyres.Err = err
					if err != nil {
						rt--
						if rt == 0 {
							chresults <- &copyres
							break
						}
						if resp != nil && resp.StatusCode < 499 && resp.StatusCode >= 400 {
							chresults <- &copyres
							break
						}
						time.Sleep(10 * time.Millisecond)
						continue
					}
					chresults <- &copyres
					break
				}
			}
		}(chjobs, chresults)
	}

	go func() {
		for _, chunk := range chunks {
			partOpt := &cos.ObjectCopyPartOptions{
				XCosCopySource: u,
			}
			job := &cos.CopyJobs{
				Name:       destPath,
				RetryTimes: 3,
				UploadId:   uploadID,
				Chunk:      chunk,
				Opt:        partOpt,
			}
			chjobs <- job
		}
		close(chjobs)
	}()
	err = nil
	for i := 0; i < partNum; i++ {
		res := <-chresults
		if res.Resp.StatusCode != 200 {
			err = fmt.Errorf("UploadID %s, part %d failed to get resp content. error: %s", uploadID, res.PartNumber, res.Err.Error())
			fmt.Println(err)
			continue
		}
		etag := res.Res.ETag
		optcom.Parts = append(optcom.Parts, cos.Object{
			PartNumber: res.PartNumber, ETag: etag},
		)
	}
	close(chresults)
	if err != nil {
		return err
	}
	sort.Sort(cos.ObjectList(optcom.Parts))

	_, _, err = t.client.Object.CompleteMultipartUpload(ctx, destPath, uploadID, optcom)
	if err != nil {
		t.client.Object.AbortMultipartUpload(ctx, destPath, uploadID)
		return err
	}
	return err
}

// Delete recursively deletes all objects stored at "path" and its subpaths.
// Delete 删除指定路径的对象
func (t *tencentcos) Delete(ctx context.Context, path string) error {
	cosObjects := make([]cos.Object, 0, storage.MaxPaginationKeys)
	var marker string
	for {
		// 获取对象列表
		opt := &cos.BucketGetOptions{
			Prefix: path,
			Marker: marker,
		}
		resp, _, err := t.client.Bucket.Get(ctx, opt)
		if err != nil || len(resp.Contents) == 0 {
			return fmt.Errorf("failed to list objects: %v", err)
		}

		// 添加要删除的对象
		for _, obj := range resp.Contents {
			if len(obj.Key) > len(path) && obj.Key[len(path)] != '/' {
				continue
			}
			cosObjects = append(cosObjects, cos.Object{
				Key: obj.Key,
			})
		}

		// 删除对象
		if len(cosObjects) > 0 {
			resp, _, err := t.client.Object.DeleteMulti(ctx, &cos.ObjectDeleteMultiOptions{
				Quiet:   false,
				Objects: cosObjects,
			})
			if err != nil {
				return err
			}

			if len(resp.DeletedObjects) != len(cosObjects) {
				errs := make([]error, 0, len(resp.Errors))
				for _, err := range resp.Errors {
					errs = append(errs, fmt.Errorf(err.Message))
				}
				return fmt.Errorf("failed to delete objects: %v", errs)
			}
		}
		cosObjects = cosObjects[:0]
		marker = resp.NextMarker
		if resp.IsTruncated {
			continue
		}
		break
	}
	return nil
}

// Reader retrieves an io.ReadCloser for the content stored at "path"
// with a given byte offset.
// May be used to resume reading a stream by providing a nonzero offset.
// Reader 返回读取指定路径的 Reader
func (t *tencentcos) Reader(ctx context.Context, path string, offset int64) (io.ReadCloser, error) {
	// 设置读取范围
	opt := &cos.ObjectGetOptions{
		Range: fmt.Sprintf("bytes=%s-", strconv.FormatInt(offset, 10)),
	}

	// 获取对象
	resp, err := t.client.Object.Get(ctx, path, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %v", err)
	}
	return resp.Body, nil
}

// CreateUploadID creates a new multipart upload and returns an
func (t *tencentcos) CreateUploadID(ctx context.Context, path string) (string, error) {
	// 创建分块上传任务
	resp, _, err := t.client.Object.InitiateMultipartUpload(ctx, path, &cos.InitiateMultipartUploadOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to create multipart upload: %v", err)
	}
	return resp.UploadID, nil
}

// UploadPart WritePart writes a part of a multipart upload.
func (t *tencentcos) UploadPart(ctx context.Context, path, uploadID string, partNumber int64, body io.Reader) (string, error) {

	newPath := fmt.Sprintf("%s-%d", path, partNumber)
	sourceURL := fmt.Sprintf("%s/%s", t.client.Host, newPath)
	_, err := t.client.Object.Put(ctx, newPath, body, nil)
	res, _, err := t.client.Object.CopyPart(ctx, path, uploadID, int(partNumber), sourceURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to upload part: %v", err)
	}
	err = t.Delete(ctx, newPath)
	if err != nil {
		return "", fmt.Errorf("failed to delete source part: %v", err)
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

// // helpers
// func encodeURIComponent(s string, excluded ...[]byte) string {
// 	var b bytes.Buffer
// 	written := 0

// 	for i, n := 0, len(s); i < n; i++ {
// 		c := s[i]

// 		switch c {
// 		case '-', '_', '.', '!', '~', '*', '\'', '(', ')':
// 			continue
// 		default:
// 			// Unreserved according to RFC 3986 sec 2.3
// 			if 'a' <= c && c <= 'z' {

// 				continue

// 			}
// 			if 'A' <= c && c <= 'Z' {

// 				continue

// 			}
// 			if '0' <= c && c <= '9' {

// 				continue
// 			}
// 			if len(excluded) > 0 {
// 				conti := false
// 				for _, ch := range excluded[0] {
// 					if ch == c {
// 						conti = true
// 						break
// 					}
// 				}
// 				if conti {
// 					continue
// 				}
// 			}
// 		}

// 		b.WriteString(s[written:i])
// 		fmt.Fprintf(&b, "%%%02X", c)
// 		written = i + 1
// 	}

// 	if written == 0 {
// 		return s
// 	}
// 	b.WriteString(s[written:])
// 	return b.String()
// }

func SplitSizeIntoChunks(totalBytes int64, partSize int64) ([]cos.Chunk, int, error) {
	var partNum int64
	if partSize > 0 {
		if partSize < 1024*1024 {
			return nil, 0, fmt.Errorf("partSize>=1048576 is required")
		}
		partNum = totalBytes / partSize
		if partNum >= 10000 {
			return nil, 0, fmt.Errorf("Too many parts, out of 10000")
		}
	} else {
		partNum, partSize = DividePart(totalBytes, 16)
	}

	var chunks []cos.Chunk
	var chunk = cos.Chunk{}
	for i := int64(0); i < partNum; i++ {
		chunk.Number = int(i + 1)
		chunk.OffSet = i * partSize
		chunk.Size = partSize
		chunks = append(chunks, chunk)
	}

	if totalBytes%partSize > 0 {
		chunk.Number = len(chunks) + 1
		chunk.OffSet = int64(len(chunks)) * partSize
		chunk.Size = totalBytes % partSize
		chunks = append(chunks, chunk)
		partNum++
	}

	return chunks, int(partNum), nil
}

func DividePart(fileSize int64, last int) (int64, int64) {
	partSize := int64(last * 1024 * 1024)
	partNum := fileSize / partSize
	for partNum >= 10000 {
		partSize = partSize * 2
		partNum = fileSize / partSize
	}
	return partNum, partSize
}
