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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/go-sigma/sigma/pkg/configs"
)

func TestBigFileMove(t *testing.T) {
	ctx := context.Background()
	var config = configs.Configuration{
		Storage: configs.ConfigurationStorage{
			Oss: configs.ConfigurationStorageOss{
				Ak:             os.Getenv("OSS_AK"),
				Sk:             os.Getenv("OSS_SK"),
				Bucket:         os.Getenv("OSS_BUCKET"),
				Endpoint:       os.Getenv("OSS_ENDPOINT"),
				ForcePathStyle: false,
			},
			RootDirectory: "sigma",
		},
	}
	f := factory{}
	driver, err := f.New(config)
	assert.NoError(t, err)

	var bigFile = "test-big-file.bin"
	originalFile, _ := os.Create(bigFile)
	for i := 0; i < 100; i++ { // generate 100MB
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
	var config = configs.Configuration{
		Storage: configs.ConfigurationStorage{
			Oss: configs.ConfigurationStorageOss{
				Ak:             os.Getenv("OSS_AK"),
				Sk:             os.Getenv("OSS_SK"),
				Bucket:         os.Getenv("OSS_BUCKET"),
				Endpoint:       os.Getenv("OSS_ENDPOINT"),
				ForcePathStyle: false,
			},
			RootDirectory: "sigma",
		},
	}
	f := factory{}
	driver, err := f.New(config)
	assert.NoError(t, err)

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
	var config = configs.Configuration{
		Storage: configs.ConfigurationStorage{
			Oss: configs.ConfigurationStorageOss{
				Ak:             os.Getenv("OSS_AK"),
				Sk:             os.Getenv("OSS_SK"),
				Bucket:         os.Getenv("OSS_BUCKET"),
				Endpoint:       os.Getenv("OSS_ENDPOINT"),
				ForcePathStyle: false,
			},
			RootDirectory: "sigma",
		},
	}
	f := factory{}
	driver, err := f.New(config)
	assert.NoError(t, err)

	uploadID, err := driver.CreateUploadID(ctx, "upload-test")
	assert.NoError(t, err)
	var bigFile1 = "test-big-file1.bin"
	originalFile1, _ := os.Create(bigFile1)
	for i := 0; i < 1; i++ { // 1M
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
	for i := 0; i < 1; i++ { // 1M
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
