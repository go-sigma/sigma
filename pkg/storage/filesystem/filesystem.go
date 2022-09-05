package filesystem

import (
	"context"
	"io"
	"os"

	"github.com/ximager/ximager/pkg/storage"
)

const (
	// name is the name of the filesystem storage driver
	name = "filesystem"
)

// fs is the filesystem storage driver
type fs struct{}

// New returns a new filesystem storage driver
func New() storage.StorageDriver {
	return &fs{}
}

// Name returns the name of the filesystem storage driver
func (f *fs) Name() string {
	return name
}

// Stat returns the file info for the given path
func (f *fs) Stat(ctx context.Context, path string) (storage.FileInfo, error) {
	return os.Stat(path)
}

// Move moves a file from sourcePath to destPath
func (f *fs) Move(ctx context.Context, sourcePath string, destPath string) error {
	return os.Rename(sourcePath, destPath)
}

// Delete deletes a file at the given path
func (f *fs) Delete(ctx context.Context, path string) error {
	return os.RemoveAll(path)
}

// Reader returns a reader for the file at the given path
func (f *fs) Reader(ctx context.Context, path string, offset int64) (io.ReadCloser, error) {
	fp, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	seekPos, err := fp.Seek(offset, io.SeekStart)
	if err != nil {
		fp.Close() // nolint: errcheck
		return nil, err
	} else if seekPos < offset {
		fp.Close() // nolint: errcheck
		return nil, err
	}
	return fp, nil
}

// CreateUploadID creates a new multipart upload and returns an
// opaque upload ID.
func (f *fs) CreateUploadID(ctx context.Context, path string) (string, error) {
	panic("implement me")
}

// WritePart writes a part of a multipart upload.
func (f *fs) UploadPart(ctx context.Context, path, uploadID string, partNumber int64, body io.Reader) (string, error) {
	panic("implement me")
}

// CommitUpload commits a multipart upload.
func (f *fs) CommitUpload(ctx context.Context, uploadID string, parts []string, path string) error {
	panic("implement me")
}
