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

package obs

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"path"
	"reflect"

	buildlogger "github.com/go-sigma/sigma/pkg/background/build/logger"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/storage"
	"github.com/go-sigma/sigma/pkg/utils"
)

func init() {
	buildlogger.LogStoreFactories.MustRegister(path.Base(reflect.TypeFor[factory]().PkgPath()), &factory{})
}

type factory struct{}

var _ buildlogger.LogStoreFactory = factory{}

// New returns an object-storage-backed build log store.
func (f factory) New() (buildlogger.LogStore, error) {
	storageDriver, err := storage.Initialize(config.GetConfig())
	if err != nil {
		return nil, err
	}
	return &obs{
		storage: storageDriver,
	}, nil
}

type obs struct {
	storage storage.StorageDriver
}

// Write returns a writer for the given id
func (o *obs) Write(builderID, runnerID string) io.WriteCloser {
	buffer := new(bytes.Buffer)
	gw := gzip.NewWriter(buffer)
	return &writer{
		builderID: builderID,
		runnerID:  runnerID,
		gw:        gw,
		data:      buffer,
		storage:   o.storage,
	}
}

type writer struct {
	builderID string
	runnerID  string
	gw        *gzip.Writer
	data      *bytes.Buffer
	storage   storage.StorageDriver
}

// Write writes the given bytes to the writer
func (w *writer) Write(p []byte) (n int, err error) {
	return w.gw.Write(p)
}

// Close closes the writer
func (w *writer) Close() error {
	err := w.gw.Flush()
	if err != nil {
		return err
	}
	return w.storage.Upload(context.Background(),
		fmt.Sprintf("%s.log.gz", path.Join(consts.BuilderLogs, fmt.Sprintf("%s/%s", w.builderID, w.runnerID))),
		bytes.NewReader(w.data.Bytes()))
}

// Read returns a reader for the given id
func (o *obs) Read(ctx context.Context, id string) (io.Reader, error) {
	return o.storage.Reader(ctx, fmt.Sprintf("%s.log.gz", path.Join(consts.BuilderLogs, utils.DirWithSlash(id))))
}
