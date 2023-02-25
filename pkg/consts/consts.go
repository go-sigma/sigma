// The MIT License (MIT)
//
// Copyright © 2023 Tosone <i@tosone.cn>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package consts

const (
	// APIVersion represents the API version
	APIVersion = "v2"
	// APIVersionKey represents the API version key
	APIVersionKey = "Docker-Distribution-API-Version"
	// APIVersionValue represents the API version value
	APIVersionValue = "registry/2.0"
	// UploadUUID represents the upload uuid in header
	UploadUUID = "Docker-Upload-UUID"
	// ContentDigest represents the content digest in header
	ContentDigest = "Docker-Content-Digest"
	// Blobs represents a blobs
	// file always represent like: blobs/{algo}/xx/xx/{digest}
	Blobs = "blobs"
	// BlobUploads represent blob uploads
	// file always represent like: blob_uploads/{upload_id}
	BlobUploads = "blob_uploads"
	// DefaultTimePattern 时间格式
	DefaultTimePattern = "2006-01-02 15:04:05"
)
