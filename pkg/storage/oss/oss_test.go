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

package oss

import (
	"bytes"
	"context"
	"crypto/rand"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-sigma/sigma/pkg/config"
)

func newTestDriver(t *testing.T) *alioss {
	t.Helper()

	var missing []string
	ak := os.Getenv("OSS_AK")
	if ak == "" {
		missing = append(missing, "OSS_AK")
	}
	sk := os.Getenv("OSS_SK")
	if sk == "" {
		missing = append(missing, "OSS_SK")
	}
	bucket := os.Getenv("OSS_BUCKET")
	if bucket == "" {
		missing = append(missing, "OSS_BUCKET")
	}
	endpoint := os.Getenv("OSS_ENDPOINT")
	if endpoint == "" {
		missing = append(missing, "OSS_ENDPOINT")
	}
	if len(missing) > 0 {
		slices.Sort(missing)
		t.Skipf("OSS integration test requires %s", strings.Join(missing, ", "))
	}

	driver, err := factory{}.New(&config.Configuration{
		Storage: config.ConfigurationStorage{
			Oss: config.ConfigurationStorageOss{
				Ak:             ak,
				Sk:             sk,
				Bucket:         bucket,
				Endpoint:       endpoint,
				ForcePathStyle: false,
			},
			RootDirectory: "sigma",
		},
	})
	require.NoError(t, err)
	require.NotNil(t, driver)
	return driver.(*alioss)
}

func TestBigFileMove(t *testing.T) {
	ctx := context.Background()
	driver := newTestDriver(t)
	var err error

	var bigFile = "test-big-file.bin"
	originalFile, _ := os.Create(bigFile)
	for range 100 { // generate 100MB
		data := make([]byte, 1<<20)
		_, _ = rand.Read(data)
		_, _ = originalFile.Write(data)
	}
	err = originalFile.Close()
	assert.NoError(t, err)

	defer os.Remove(bigFile) // nolint: errcheck

	bigFileBytes, err := os.ReadFile(bigFile)
	assert.NoError(t, err)

	err = driver.Upload(ctx, "big-file", bytes.NewReader(bigFileBytes))
	assert.NoError(t, err)
	err = driver.Move(ctx, "big-file", "big-file-move-to")
	assert.NoError(t, err)

	err = driver.Delete(ctx, "big-file")
	assert.NoError(t, err)
	err = driver.Delete(ctx, "big-file-move-to")
	assert.NoError(t, err)
}

func TestDelete(t *testing.T) {
	ctx := context.Background()
	driver := newTestDriver(t)
	var err error

	err = driver.Upload(ctx, "dir/unit-test/unit-test/test.txt", strings.NewReader("test"))
	assert.NoError(t, err)
	err = driver.Upload(ctx, "dir/unit-test/test.txt", strings.NewReader("test"))
	assert.NoError(t, err)
	err = driver.Upload(ctx, "dir/test.txt", strings.NewReader("test"))
	assert.NoError(t, err)

	err = driver.Delete(ctx, "dir")
	assert.NoError(t, err)
}

func TestMultiUpload(t *testing.T) {
	ctx := context.Background()
	driver := newTestDriver(t)
	var err error

	uploadID, err := driver.CreateUploadID(ctx, "upload-test")
	assert.NoError(t, err)
	var bigFile1 = "test-big-file1.bin"
	originalFile1, _ := os.Create(bigFile1)
	for range 1 { // 1M
		data := make([]byte, 1<<20)
		_, _ = rand.Read(data)
		_, _ = originalFile1.Write(data)
	}
	_ = originalFile1.Close()
	file1Bytes, err := os.ReadFile(bigFile1)
	assert.NoError(t, err)
	defer os.RemoveAll(bigFile1) // nolint: errcheck
	etag1, err := driver.UploadPart(ctx, "upload-test", uploadID, 1, bytes.NewReader(file1Bytes))
	assert.NoError(t, err)
	var bigFile2 = "test-big-file2.bin"
	originalFile2, _ := os.Create(bigFile2)
	for range 1 { // 1M
		data := make([]byte, 1<<20)
		_, _ = rand.Read(data)
		_, _ = originalFile2.Write(data)
	}
	_ = originalFile1.Close()
	file2, _ := os.Open(bigFile2)
	defer file2.Close()          // nolint: errcheck
	defer os.RemoveAll(bigFile2) // nolint: errcheck
	etag2, err := driver.UploadPart(ctx, "upload-test", uploadID, 2, file2)
	assert.NoError(t, err)
	err = driver.CommitUpload(ctx, "upload-test", uploadID, []string{etag1, etag2})
	assert.NoError(t, err)
	err = driver.Move(ctx, "upload-test", "upload-test-move-to")
	assert.NoError(t, err)
	err = driver.Delete(ctx, "upload-test-move-to")
	assert.NoError(t, err)
	err = driver.Delete(ctx, "upload-test")
	assert.NoError(t, err)
}
