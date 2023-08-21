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

package cos

import (
	"context"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/tencentyun/cos-go-sdk-v5"

	"github.com/go-sigma/sigma/pkg/configs"
)

func TestNew(t *testing.T) {
	viper.Reset()
	viper.SetDefault("storage.cos.endpoint", "https://sigma-1251887554.cos.ap-beijing.myqcloud.com")
	viper.SetDefault("storage.cos.ak", "AKID04s63l4XbV7RU5mZPyHlxtYFshu1OpKY")
	viper.SetDefault("storage.cos.sk", "94j8HBCwZyEVdVBkPmkThiV2wkIabNbY")

	ctx := context.Background()
	var f = factory{}
	driver, err := f.New(configs.Configuration{
		Storage: configs.ConfigurationStorage{
			Cos: configs.ConfigurationStorageCos{
				Endpoint: "https://xxx.cos.ap-beijing.myqcloud.com",
				Ak:       "xxx",
				Sk:       "xxx",
			},
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, driver)

	err = driver.Upload(ctx, "unit-test", strings.NewReader("test"))
	assert.NoError(t, err)
	err = driver.Delete(ctx, "unit-test")
	assert.NoError(t, err)

	_, err = driver.Stat(ctx, "none-exist")
	assert.Error(t, err)

	err = driver.Move(ctx, "none-exist", "none-exist")
	assert.Error(t, err)

	_, err = driver.CreateUploadID(ctx, "failed")
	assert.NoError(t, err)

	err = driver.Upload(ctx, "dir/unit-test", strings.NewReader("test"))
	assert.NoError(t, err)

	fileInfo, err := driver.Stat(ctx, "dir/unit-test")
	assert.NoError(t, err)
	assert.Equal(t, fileInfo.IsDir(), false)
	assert.Equal(t, fileInfo.Name(), "dir/unit-test")
	assert.Equal(t, fileInfo.Size(), int64(4))
	assert.NotNil(t, fileInfo.ModTime())

	err = driver.Move(ctx, "dir/unit-test", "dir/unit-test-to")
	assert.NoError(t, err)

	_, err = driver.Stat(ctx, "none-exist")
	assert.Equal(t, false, cos.IsNotFoundError(err))

	// reader, err := driver.Reader(ctx, "dir/unit-test/README.md", 0)
	// assert.NoError(t, err)
	// contentBytes, err := io.ReadAll(reader)
	// assert.NoError(t, err)
	// assert.NotNil(t, string(contentBytes))

	// var bigFile1 = "test-big-file1.bin"
	// uploadID, err := driver.CreateUploadID(ctx, "upload-test")
	// assert.NoError(t, err)
	// originalFile1, _ := os.Create(bigFile1)
	// for i := 0; i < 100; i++ { // 100M
	// 	data := make([]byte, 1<<20)
	// 	_, _ = rand.Read(data)
	// 	_, _ = originalFile1.Write(data)
	// }
	// _ = originalFile1.Close()

	// file1, _ := os.Open(bigFile1)
	// defer file1.Close()          // nolint: errcheck
	// defer os.RemoveAll(bigFile1) // nolint: errcheck
	// etag1, err := driver.UploadPart(ctx, "upload-test", uploadID, 1, file1)
	// assert.NoError(t, err)

	// var bigFile2 = "test-big-file2.bin"
	// originalFile2, _ := os.Create(bigFile2)
	// for i := 0; i < 100; i++ { // 100M
	// 	data := make([]byte, 1<<20)
	// 	_, _ = rand.Read(data)
	// 	_, _ = originalFile2.Write(data)
	// }
	// _ = originalFile2.Close()
	// file2, _ := os.Open(bigFile2)
	// defer file2.Close()          // nolint: errcheck
	// defer os.RemoveAll(bigFile2) // nolint: errcheck
	// etag2, err := driver.UploadPart(ctx, "upload-test", uploadID, 2, file2)
	// assert.NoError(t, err)
	// err = driver.CommitUpload(ctx, "upload-test", uploadID, []string{etag1, etag2})
	// assert.NoError(t, err)
	// err = driver.Move(ctx, "upload-test", "upload-test-move-to")
	// assert.NoError(t, err)
	// f2i, err := driver.Stat(ctx, "upload-test-move-to")
	// assert.NoError(t, err)
	// assert.Equal(t, int64(200*1<<20), f2i.Size())
	// err = driver.Delete(ctx, "upload-test-move-to")
	// assert.NoError(t, err)

	// uploadID, err = driver.CreateUploadID(ctx, "upload-test-move-to")
	// assert.NoError(t, err)
	// err = driver.AbortUpload(ctx, "upload-test-move-to", uploadID)
	// assert.NoError(t, err)
}
