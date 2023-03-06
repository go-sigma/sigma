// Copyright 2023 XImager
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

package s3

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	viper.SetDefault("storage.s3.endpoint", "http://localhost:9000")
	viper.SetDefault("storage.s3.region", "cn-north-1")
	viper.SetDefault("storage.s3.ak", "ximager")
	viper.SetDefault("storage.s3.sk", "ximager-ximager")
	viper.SetDefault("storage.s3.bucket", "ximager")
	viper.SetDefault("storage.s3.forcePathStyle", true)

	var f = factory{}
	driver, err := f.New()
	assert.NoError(t, err)
	assert.NotNil(t, driver)
}
