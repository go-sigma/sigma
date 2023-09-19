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

package dal

import (
	"fmt"
	"net/url"
	"os"
	"reflect"
	"testing"

	"github.com/alicebob/miniredis/v2"
	gonanoid "github.com/matoous/go-nanoid"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/go-sigma/sigma/pkg/logger"
)

func TestAuth(t *testing.T) {
	logger.SetLevel("debug")

	dbPath := fmt.Sprintf("%s.db", gonanoid.MustGenerate("abcdefghijklmnopqrstuvwxyz", 6))
	viper.SetDefault("database.type", "sqlite3")
	viper.SetDefault("database.sqlite3.path", dbPath)
	miniRedis := miniredis.RunT(t)
	viper.SetDefault("redis.url", "redis://"+miniRedis.Addr())

	err := Initialize()
	assert.NoError(t, err)

	added, _ := AuthEnforcer.AddPolicy("library_reader", "library", "DS$*/**$manifests$*", "public", "(GET)|(HEAD)", "allow")
	assert.True(t, added)
	added, _ = AuthEnforcer.AddRoleForUser("alice", "library_reader", "library")
	assert.True(t, added)

	passed, err := AuthEnforcer.Enforce("alice", "library", "/v2/library/busybox/manifests/latest", "public", "GET")
	assert.NoError(t, err)
	assert.True(t, passed)
	passed, err = AuthEnforcer.Enforce("alice", "library", "/v2/library/busybox/manifests/sha256:xxx", "public", "GET")
	assert.NoError(t, err)
	assert.True(t, passed)
	passed, err = AuthEnforcer.Enforce("alice", "library", "/v2/library/busybox/manifests/sha256:xxx", "public", "POST")
	assert.NoError(t, err)
	assert.False(t, passed)

	assert.NoError(t, os.Remove(dbPath))
}

func TestUrlMatchFunc(t *testing.T) {
	q := make(url.Values)
	q.Set("repository", "library/busybox")

	type args struct {
		args []any
	}
	tests := []struct {
		name    string
		args    args
		want    any
		wantErr bool
	}{
		{
			name: "normal-ds",
			args: args{
				args: []any{"/v2/?test=1", "DS$v2"},
			},
			want:    true,
			wantErr: false,
		}, {
			name: "parse-url-failed",
			args: args{
				args: []any{"::::////v2/?test=1", "DS$v2"},
			},
			want:    false,
			wantErr: true,
		}, {
			name: "normal-api-1",
			args: args{
				args: []any{"/api/namespaces/", "API$namespaces/"},
			},
			want:    true,
			wantErr: false,
		}, {
			name: "normal-api-2",
			args: args{
				args: []any{"/api/namespaces/10?" + q.Encode(), "API$*/**$namespaces/*"},
			},
			want:    true,
			wantErr: false,
		}, {
			name: "no-match",
			args: args{
				args: []any{"/v3/api/namespaces/10", ""},
			},
			want:    false,
			wantErr: false,
		}, {
			name: "catalog",
			args: args{
				args: []any{"/v2/_catalog", "DS$catalog"},
			},
			want:    true,
			wantErr: false,
		}, {
			name: "tags",
			args: args{
				args: []any{"/v2/library/busybox/tags/list", "DS$library/busybox$tags"},
			},
			want:    true,
			wantErr: false,
		}, {
			name: "blob_uploads",
			args: args{
				args: []any{"/v2/library/busybox/blobs/uploads/", "DS$library/busybox$blob_uploads"},
			},
			want:    true,
			wantErr: false,
		}, {
			name: "blobs",
			args: args{
				args: []any{"/v2/library/busybox/blobs/sha256:xxx", "DS$library/busybox$blobs$*"},
			},
			want:    true,
			wantErr: false,
		}, {
			name: "manifests",
			args: args{
				args: []any{"/v2/library/busybox/manifests/sha256:xxx", "DS$library/busybox$manifests$*"},
			},
			want:    true,
			wantErr: false,
		}, {
			name: "blob_uploads",
			args: args{
				args: []any{"/v2/library/busybox/blobs/uploads/sha256:xxx", "DS$library/busybox$blob_uploads$*"},
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := urlMatchFunc(tt.args.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("urlMatchFunc() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("urlMatchFunc() = %v, want %v", got, tt.want)
			}
		})
	}
}
