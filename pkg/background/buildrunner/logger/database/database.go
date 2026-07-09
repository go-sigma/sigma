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

package database

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log/slog"
	"path"
	"reflect"

	builderlogger "github.com/go-sigma/sigma/pkg/background/buildrunner/logger"
	"github.com/go-sigma/sigma/pkg/dal/query"
	repobuilder "github.com/go-sigma/sigma/pkg/dal/repository/builder"
)

func init() {
	builderlogger.DriverFactories.MustRegister(path.Base(reflect.TypeFor[factory]().PkgPath()), &factory{})
}

type factory struct{}

var _ builderlogger.Factory = factory{}

// New returns a new filesystem storage driver
func (f factory) New() (builderlogger.BuilderLogger, error) {
	return &database{
		builderRepository: repobuilder.NewBuilderRepository(),
	}, nil
}

type database struct {
	builderRepository repobuilder.BuilderRepository
}

func (d *database) Write(builderID, runnerID string) io.WriteCloser {
	buffer := new(bytes.Buffer)
	gw := gzip.NewWriter(buffer)
	return &writer{
		builderID: builderID,
		runnerID:  runnerID,
		gw:        gw,
		data:      buffer,
		db:        d.builderRepository,
	}
}

type writer struct {
	builderID string
	runnerID  string
	gw        *gzip.Writer
	data      *bytes.Buffer
	db        repobuilder.BuilderRepository
}

// Write writes the given bytes to the underlying storage
func (w *writer) Write(p []byte) (n int, err error) {
	return w.gw.Write(p)
}

// Close closes the writer and flushes the data to the underlying storage
func (w *writer) Close() error {
	err := w.gw.Flush()
	if err != nil {
		return err
	}
	builderRepository := w.db
	data := w.data.Bytes()
	slog.Info("create builder log success", "len", len(data))
	updates := map[string]any{
		query.BuilderRunner.Log.ColumnName().String(): data,
		query.BuilderRunner.ID.ColumnName().String():  w.runnerID,
	}
	return builderRepository.UpdateRunner(context.Background(), w.builderID, w.runnerID, updates)
}

// Read returns a reader for the given id
func (d *database) Read(ctx context.Context, id string) (io.Reader, error) {
	builderRepository := d.builderRepository
	builderLog, err := builderRepository.GetRunner(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get builder log: %v", err)
	}
	return bytes.NewReader(builderLog.Log), nil
}
