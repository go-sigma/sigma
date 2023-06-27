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

package dal

import (
	"fmt"
	"os"
	"testing"

	gonanoid "github.com/matoous/go-nanoid"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/ximager/ximager/pkg/logger"
)

func TestAuth(t *testing.T) {
	logger.SetLevel("debug")

	dbPath := fmt.Sprintf("%s.db", gonanoid.MustGenerate("abcdefghijklmnopqrstuvwxyz", 6))
	viper.SetDefault("database.type", "sqlite3")
	viper.SetDefault("database.sqlite3.path", dbPath)

	err := Initialize()
	assert.NoError(t, err)

	added, _ := AuthEnforcer.AddPolicy("library_reader", "library", `/v2/(?:(?:(?:[a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9])(?:\.(?:[a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9]))*|\[(?:[a-fA-F0-9:]+)\])(?::[0-9]+)?/)?[a-z0-9]+(?:(?:[._]|__|[-]+)[a-z0-9]+)*(?:/[a-z0-9]+(?:(?:[._]|__|[-]+)[a-z0-9]+)*)*/manifests/[\w][\w.-]{0,127}|[a-z0-9]+(?:[.+_-][a-z0-9]+)*:[a-zA-Z0-9=_-]+`, "public", "(GET)|(HEAD)", "allow")
	assert.True(t, added)
	added, _ = AuthEnforcer.AddRoleForUser("alice", "library_reader", "library")
	assert.True(t, added)

	passed, err := AuthEnforcer.Enforce("alice", "library", "/v2/xxx/xxx/manifests/ssss", "public", "GET")
	assert.NoError(t, err)
	assert.True(t, passed)
	passed, err = AuthEnforcer.Enforce("alice", "library", "/v2/xxx/xxx/manifests1/sha256:xxx", "public", "GET")
	assert.NoError(t, err)
	assert.True(t, passed)

	assert.NoError(t, os.Remove(dbPath))
}
