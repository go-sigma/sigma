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
)
