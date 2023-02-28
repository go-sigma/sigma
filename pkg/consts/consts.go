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

package consts

const (
	// AppName represents the app name
	AppName = "XImager"
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
	// DefaultTimePattern time pattern
	DefaultTimePattern = "2006-01-02 15:04:05"
)
