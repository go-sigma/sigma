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

package filesystem

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/go-sigma/sigma/pkg/configs"
)

func TestNew(t *testing.T) {
	var config = configs.Configuration{}
	config.Storage.Filesystem.Path = "test"
	config.Storage.RootDirectory = "storage"

	f := factory{}
	driver, err := f.New(config)
	assert.NoError(t, err)
	assert.NotNil(t, driver)

	err = os.WriteFile("test/storage/unit-test", []byte("test"), 0600)
	assert.NoError(t, err)
	err = driver.Move(context.Background(), "unit-test", "unit-test-2")
	assert.NoError(t, err)
	_, err = os.Stat("test/storage/unit-test")
	assert.True(t, errors.Is(err, os.ErrNotExist))
	_, err = os.Stat("test/storage/unit-test-2")
	assert.NoError(t, err)
	err = driver.Delete(context.Background(), "test/storage/unit-test-2")
	assert.NoError(t, err)

	err = os.WriteFile("test/storage/unit-test", []byte("test"), 0600)
	assert.NoError(t, err)
	reader, err := driver.Reader(context.Background(), "unit-test")
	assert.NoError(t, err)
	assert.NotNil(t, reader)
	dataBytes, err := io.ReadAll(reader)
	assert.NoError(t, err)
	assert.Equal(t, "test", string(dataBytes))
	reader, err = driver.Reader(context.Background(), "unit-test")
	assert.NoError(t, err)
	assert.NotNil(t, reader)
	dataBytes, err = io.ReadAll(reader)
	assert.NoError(t, err)
	assert.Equal(t, "test", string(dataBytes))
	reader, err = driver.Reader(context.Background(), "unit-test")
	assert.NoError(t, err)
	dataBytes, err = io.ReadAll(reader)
	assert.NoError(t, err)
	assert.Equal(t, "test", string(dataBytes))
	err = driver.Delete(context.Background(), "unit-test")
	assert.NoError(t, err)

	uploadID, err := driver.CreateUploadID(context.Background(), "unit-test-path")
	assert.NoError(t, err)
	assert.NotEmpty(t, uploadID)
	tag1, err := driver.UploadPart(context.Background(), "unit-test-path", uploadID, 1, strings.NewReader("test"))
	assert.NoError(t, err)
	assert.NotEmpty(t, tag1)
	tag2, err := driver.UploadPart(context.Background(), "unit-test-path", uploadID, 2, strings.NewReader("hello"))
	assert.NoError(t, err)
	assert.NotEmpty(t, tag2)
	err = driver.CommitUpload(context.Background(), "unit-test-path", uploadID, []string{tag1, tag2})
	assert.NoError(t, err)
	reader, err = driver.Reader(context.Background(), "unit-test-path")
	assert.NoError(t, err)
	assert.NotNil(t, reader)
	dataBytes, err = io.ReadAll(reader)
	assert.NoError(t, err)
	assert.Equal(t, "test"+"hello", string(dataBytes))
	err = driver.Delete(context.Background(), "unit-test-path")
	assert.NoError(t, err)

	err = driver.CommitUpload(context.Background(), "unit-test-path", uploadID, []string{tag1, tag2})
	assert.Error(t, err)
	err = os.RemoveAll("unit-test-path.fake")
	assert.NoError(t, err)

	uploadID, err = driver.CreateUploadID(context.Background(), "unit-test-path")
	assert.NoError(t, err)
	assert.NotEmpty(t, uploadID)
	tag1, err = driver.UploadPart(context.Background(), "unit-test-path", uploadID, 1, strings.NewReader("test"))
	assert.NoError(t, err)
	assert.NotEmpty(t, tag1)
	err = driver.AbortUpload(context.Background(), "unit-test-path", uploadID)
	assert.NoError(t, err)

	err = driver.Upload(context.Background(), "unit-test-path", strings.NewReader("test"))
	assert.NoError(t, err)
	reader, err = driver.Reader(context.Background(), "unit-test-path")
	assert.NoError(t, err)
	assert.NotNil(t, reader)
	dataBytes, err = io.ReadAll(reader)
	assert.NoError(t, err)
	assert.Equal(t, "test", string(dataBytes))
	err = driver.Delete(context.Background(), "unit-test-path")
	assert.NoError(t, err)

	err = driver.Upload(context.Background(), "unit-test-path", nil)
	assert.Error(t, err)
}
