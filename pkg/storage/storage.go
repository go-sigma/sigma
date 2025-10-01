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

package storage

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/types/enums"
)

const (
	// MultipartCopyThresholdSize ...
	MultipartCopyThresholdSize = 64 << 20 // 64MB
	// MultipartCopyChunkSize ...
	MultipartCopyChunkSize = 64 << 20 // 64MB
	// MultipartCopyMaxConcurrency ...
	MultipartCopyMaxConcurrency = 100 // 100 goroutines
	// MaxPaginationKeys ...
	MaxPaginationKeys = 1000 // 1000 keys
)

//go:generate mockgen -destination=storage_mocks.go -package=storage github.com/go-sigma/sigma/pkg/storage StorageDriver

// StorageDriver is the interface for the storage driver
type StorageDriver interface {
	// Move moves an object stored at sourcePath to destPath, removing the
	// original object.
	// Note: This may be no more efficient than a copy followed by a delete for
	// many implementations.
	Move(ctx context.Context, sourcePath string, destPath string) error

	// Delete recursively deletes all objects stored at "path" and its subpaths.
	Delete(ctx context.Context, path string) error

	// Reader retrieves an io.ReadCloser for the content stored at "path"
	// with a given byte offset.
	Reader(ctx context.Context, path string) (io.ReadCloser, error)

	// CreateUploadID creates a new multipart upload and returns an
	// opaque upload ID.
	CreateUploadID(ctx context.Context, path string) (string, error)

	// WritePart writes a part of a multipart upload.
	UploadPart(ctx context.Context, path, uploadID string, partNumber int32, body io.Reader) (string, error)

	// CommitUpload commits a multipart upload.
	CommitUpload(ctx context.Context, path string, uploadID string, parts []string) error

	// AbortUpload aborts a multipart upload.
	AbortUpload(ctx context.Context, path string, uploadID string) error

	// Upload upload a file to the given path.
	Upload(ctx context.Context, path string, body io.Reader) error

	// Redirect get a temporary link
	Redirect(ctx context.Context, path string) (string, error)
}

// Factory is the interface for the storage driver factory
type Factory interface {
	New(config configs.Configuration) (StorageDriver, error)
}

var factories = make(map[enums.StorageType]Factory)

// Register registers a storage factory driver by name.
// If Register is called twice with the same name or if driver is nil, it panics.
func Register(name enums.StorageType, factory Factory) error {
	if _, ok := factories[name]; ok {
		return fmt.Errorf("driver %q already registered", name)
	}
	factories[name] = factory
	return nil
}

// Initialize initializes the storage driver
func Initialize(config configs.Configuration) (StorageDriver, error) {
	factory, ok := factories[config.Storage.Type]
	if !ok {
		return nil, fmt.Errorf("driver %q not registered", config.Storage.Type)
	}
	return factory.New(config)
}

// SanitizePath ...
func SanitizePath(rootDirectory, p string) string {
	if rootDirectory == "" || rootDirectory == "." || rootDirectory == "./" || rootDirectory == "/" {
		return p
	}
	return path.Join(strings.TrimPrefix(rootDirectory, "./"), strings.TrimPrefix(p, "./"))
}
