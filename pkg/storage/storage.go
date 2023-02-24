package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/spf13/viper"
)

// FileInfo returns information about a given path. Inspired by os.FileInfo,
// it elides the base name method for a full path instead.
type FileInfo interface {
	// Name provides the full path of the target of this file info.
	Name() string

	// Size returns current length in bytes of the file. The return value can
	// be used to write to the end of the file at path. The value is
	// meaningless if IsDir returns true.
	Size() int64

	// ModTime returns the modification time for the file. For backends that
	// don't have a modification time, the creation time should be returned.
	ModTime() time.Time

	// IsDir returns true if the path is a directory.
	IsDir() bool
}

// StorageDriver is the interface for the storage driver
type StorageDriver interface {
	// Name returns the human-readable "name" of the driver, useful in error
	// messages and logging. By convention, this will just be the registration
	// name, but drivers may provide other information here.
	Name() string

	// Stat retrieves the FileInfo for the given path, including the current
	// size in bytes and the creation time.
	Stat(ctx context.Context, path string) (FileInfo, error)

	// Move moves an object stored at sourcePath to destPath, removing the
	// original object.
	// Note: This may be no more efficient than a copy followed by a delete for
	// many implementations.
	Move(ctx context.Context, sourcePath string, destPath string) error

	// Delete recursively deletes all objects stored at "path" and its subpaths.
	Delete(ctx context.Context, path string) error

	// Reader retrieves an io.ReadCloser for the content stored at "path"
	// with a given byte offset.
	// May be used to resume reading a stream by providing a nonzero offset.
	Reader(ctx context.Context, path string, offset int64) (io.ReadCloser, error)

	// CreateUploadID creates a new multipart upload and returns an
	// opaque upload ID.
	CreateUploadID(ctx context.Context, path string) (string, error)

	// WritePart writes a part of a multipart upload.
	UploadPart(ctx context.Context, path, uploadID string, partNumber int64, body io.Reader) (string, error)

	// CommitUpload commits a multipart upload.
	CommitUpload(ctx context.Context, path string, uploadID string, parts []string) error

	// AbortUpload aborts a multipart upload.
	AbortUpload(ctx context.Context, path string, uploadID string) error

	// Upload upload a file to the given path.
	Upload(ctx context.Context, path string, body io.Reader) error
}

// Factory is the interface for the storage driver factory
type Factory interface {
	New() (StorageDriver, error)
}

var driverFactories = make(map[string]Factory)

// RegisterDriverFactory registers a storage factory driver by name.
// If RegisterDriverFactory is called twice with the same name or if driver is nil, it panics.
func RegisterDriverFactory(name string, factory Factory) error {
	if _, ok := driverFactories[name]; ok {
		return fmt.Errorf("driver %q already registered", name)
	}
	driverFactories[name] = factory
	return nil
}

// Driver is the storage driver
var Driver StorageDriver

// Initialize initializes the storage driver
func Initialize() error {
	typ := viper.GetString("storage.type")
	factory, ok := driverFactories[typ]
	if !ok {
		return fmt.Errorf("driver %q not registered", typ)
	}
	var err error
	Driver, err = factory.New()
	if err != nil {
		return err
	}
	return nil
}
