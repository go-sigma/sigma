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

package logger

import (
	"context"
	"io"

	"github.com/go-sigma/sigma/pkg/infra/registry"
)

// BuilderLogger ...
type BuilderLogger interface {
	// Write write log to object storage or database
	Write(builderID, runnerID string) io.WriteCloser
	// Read get log from object storage or database
	Read(ctx context.Context, id string) (io.Reader, error)
}

// Driver is the builder logger driver, maybe implement by s3, database, etc.
var Driver BuilderLogger

// Factory is the interface for the builder logger factory
type Factory interface {
	New() (BuilderLogger, error)
}

// DriverFactories ...
var DriverFactories = registry.New[Factory]()

func Initialize() error {
	factory, err := DriverFactories.Get("database")
	if err != nil {
		return err
	}

	Driver, err = factory.New()
	if err != nil {
		return err
	}

	return nil
}
