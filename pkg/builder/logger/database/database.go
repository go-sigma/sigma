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

	"github.com/rs/zerolog/log"

	builderlogger "github.com/go-sigma/sigma/pkg/builder/logger"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/dal/query"
)

func init() {
	builderlogger.DriverFactories[path.Base(reflect.TypeOf(factory{}).PkgPath())] = &factory{}
}

type factory struct{}

var _ builderlogger.Factory = factory{}

// New returns a new filesystem storage driver
func (f factory) New() (builderlogger.BuilderLogger, error) {
	return &database{
		builderServiceFactory: dao.NewBuilderServiceFactory(),
	}, nil
}

type database struct {
	builderServiceFactory dao.BuilderServiceFactory
}

func (d *database) Write(builderID, runnerID int64) io.WriteCloser {
	buffer := new(bytes.Buffer)
	gw := gzip.NewWriter(buffer)
	return &writer{
		builderID: builderID,
		runnerID:  runnerID,
		gw:        gw,
		data:      buffer,
		db:        d.builderServiceFactory,
	}
}

type writer struct {
	builderID int64
	runnerID  int64
	gw        *gzip.Writer
	data      *bytes.Buffer
	db        dao.BuilderServiceFactory
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
	builderService := w.db.New()
	data := w.data.Bytes()
	log.Info().Int("len", len(data)).Msg("Create builder log success")
	updates := map[string]any{
		query.BuilderRunner.Log.ColumnName().String(): data,
		query.BuilderRunner.ID.ColumnName().String():  w.runnerID,
	}
	return builderService.UpdateRunner(context.Background(), w.builderID, w.runnerID, updates)
}

// Read returns a reader for the given id
func (d *database) Read(ctx context.Context, id int64) (io.Reader, error) {
	builderService := d.builderServiceFactory.New()
	builderLog, err := builderService.GetRunner(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("Failed to get builder log: %v", err)
	}
	return bytes.NewReader(builderLog.Log), nil
}
