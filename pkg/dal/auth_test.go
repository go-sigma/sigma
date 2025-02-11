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

package dal_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/dal"
	"github.com/go-sigma/sigma/pkg/dal/badger"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/logger"
	"github.com/go-sigma/sigma/pkg/modules/locker"
	"github.com/go-sigma/sigma/pkg/modules/locker/definition"
	"github.com/go-sigma/sigma/pkg/types/enums"
)

// import (
// 	"context"
// 	"fmt"
// 	"net/url"
// 	"os"
// 	"reflect"
// 	"testing"

// 	gonanoid "github.com/matoous/go-nanoid"
// 	"github.com/rs/zerolog/log"
// 	"github.com/stretchr/testify/assert"

// 	"github.com/go-sigma/sigma/pkg/configs"
// 	"github.com/go-sigma/sigma/pkg/dal"
// 	"github.com/go-sigma/sigma/pkg/dal/badger"
// 	"github.com/go-sigma/sigma/pkg/dal/dao"
// 	"github.com/go-sigma/sigma/pkg/dal/models"
// 	"github.com/go-sigma/sigma/pkg/logger"
// 	"github.com/go-sigma/sigma/pkg/modules/locker"
// 	"github.com/go-sigma/sigma/pkg/types/enums"
// )

func TestAuth(t *testing.T) {
	logger.SetLevel("debug")

	badgerDir, err := os.MkdirTemp("", "badger")
	require.NoError(t, err)

	digCon := dig.New()
	require.NoError(t, digCon.Provide(func() configs.Configuration {
		return configs.Configuration{
			Database: configs.ConfigurationDatabase{
				Type: enums.DatabaseSqlite3,
				Sqlite3: configs.ConfigurationDatabaseSqlite3{
					Path: fmt.Sprintf("%s.db", strings.ReplaceAll(uuid.Must(uuid.NewV7()).String(), "-", "")),
				},
			},
			Locker: configs.ConfigurationLocker{
				Type:   enums.LockerTypeBadger,
				Badger: configs.ConfigurationLockerBadger{},
				Prefix: "sigma-locker",
			},
			Badger: configs.ConfigurationBadger{
				Enabled: true,
				Path:    badgerDir,
			},
		}
	}))

	require.NoError(t, digCon.Provide(badger.New))
	require.NoError(t, digCon.Provide(func() (definition.Locker, error) { return locker.Initialize(digCon) }))
	require.NoError(t, dal.Initialize(digCon))

	ctx := log.Logger.WithContext(context.Background())

	nsMemberSvc := dao.NewNamespaceMemberServiceFactory().New()

	_, err = nsMemberSvc.AddNamespaceMember(ctx, 1, models.Namespace{ID: 1, Name: "library"}, enums.NamespaceRoleManager)
	require.NoError(t, err)
	err = dal.AuthEnforcer.LoadPolicy()
	require.NoError(t, err)

	passed, vals, err := dal.AuthEnforcer.EnforceEx("1", "library", "/v2/library/busybox/manifests/latest", "public", "GET")
	fmt.Println(vals)
	require.NoError(t, err)
	require.True(t, passed)
	passed, err = dal.AuthEnforcer.Enforce("1", "library", "/v2/library/busybox/manifests/sha256:xxx", "public", "GET")
	require.NoError(t, err)
	require.True(t, passed)

	// dal.AuthEnforcer.Enforcer.GetAllRolesByDomain()
}

// func TestAuth(t *testing.T) {
// 	logger.SetLevel("debug")

// 	assert.NoError(t, badger.Initialize(context.Background(), configs.Configuration{}))

// 	err := locker.Initialize(configs.Configuration{})
// 	assert.NoError(t, err)

// 	dbPath := fmt.Sprintf("%s.db", gonanoid.MustGenerate("abcdefghijklmnopqrstuvwxyz", 6))

// 	assert.NoError(t, dal.Initialize(configs.Configuration{
// 		Database: configs.ConfigurationDatabase{
// 			Type: enums.DatabaseSqlite3,
// 			Sqlite3: configs.ConfigurationDatabaseSqlite3{
// 				Path: dbPath,
// 			},
// 		},
// 	}))

// 	ctx := log.Logger.WithContext(context.Background())
// 	namespaceMemberService := dao.NewNamespaceMemberServiceFactory().New()

// 	added, _ := dal.AuthEnforcer.AddPolicy(enums.NamespaceRoleManager.String(), "library", "DS$*/**$manifests$*", "public", "(GET)|(HEAD)", "allow")
// 	assert.True(t, added)

// 	_, err = namespaceMemberService.AddNamespaceMember(ctx, 1, models.Namespace{ID: 1, Name: "library"}, enums.NamespaceRoleManager)
// 	assert.NoError(t, err)
// 	err = dal.AuthEnforcer.LoadPolicy()
// 	assert.NoError(t, err)

// 	passed, err := dal.AuthEnforcer.Enforce("1", "library", "/v2/library/busybox/manifests/latest", "public", "GET")
// 	assert.NoError(t, err)
// 	assert.True(t, passed)
// 	passed, err = dal.AuthEnforcer.Enforce("1", "library", "/v2/library/busybox/manifests/sha256:xxx", "public", "GET")
// 	assert.NoError(t, err)
// 	assert.True(t, passed)

// 	passed, err = dal.AuthEnforcer.Enforce("1", "library", "/v2/library/busybox/manifests/sha256:xxx", "public", "POST")
// 	assert.NoError(t, err)
// 	assert.False(t, passed)

// 	assert.NoError(t, os.Remove(dbPath))
// }

// func TestUrlMatchFunc(t *testing.T) {
// 	q := make(url.Values)
// 	q.Set("repository", "library/busybox")

// 	type args struct {
// 		args []any
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    any
// 		wantErr bool
// 	}{
// 		{
// 			name: "normal-ds",
// 			args: args{
// 				args: []any{"/v2/?test=1", "DS$v2"},
// 			},
// 			want:    true,
// 			wantErr: false,
// 		}, {
// 			name: "parse-url-failed",
// 			args: args{
// 				args: []any{"::::////v2/?test=1", "DS$v2"},
// 			},
// 			want:    false,
// 			wantErr: true,
// 		}, {
// 			name: "normal-api-1",
// 			args: args{
// 				args: []any{"/api/namespaces/", "API$namespaces/"},
// 			},
// 			want:    true,
// 			wantErr: false,
// 		}, {
// 			name: "normal-api-2",
// 			args: args{
// 				args: []any{"/api/namespaces/10?" + q.Encode(), "API$*/**$namespaces/*"},
// 			},
// 			want:    true,
// 			wantErr: false,
// 		}, {
// 			name: "no-match",
// 			args: args{
// 				args: []any{"/v3/api/namespaces/10", ""},
// 			},
// 			want:    false,
// 			wantErr: false,
// 		}, {
// 			name: "catalog",
// 			args: args{
// 				args: []any{"/v2/_catalog", "DS$catalog"},
// 			},
// 			want:    true,
// 			wantErr: false,
// 		}, {
// 			name: "tags",
// 			args: args{
// 				args: []any{"/v2/library/busybox/tags/list", "DS$library/busybox$tags"},
// 			},
// 			want:    true,
// 			wantErr: false,
// 		}, {
// 			name: "blob_uploads",
// 			args: args{
// 				args: []any{"/v2/library/busybox/blobs/uploads/", "DS$library/busybox$blob_uploads"},
// 			},
// 			want:    true,
// 			wantErr: false,
// 		}, {
// 			name: "blobs",
// 			args: args{
// 				args: []any{"/v2/library/busybox/blobs/sha256:xxx", "DS$library/busybox$blobs$*"},
// 			},
// 			want:    true,
// 			wantErr: false,
// 		}, {
// 			name: "manifests",
// 			args: args{
// 				args: []any{"/v2/library/busybox/manifests/sha256:xxx", "DS$library/busybox$manifests$*"},
// 			},
// 			want:    true,
// 			wantErr: false,
// 		}, {
// 			name: "blob_uploads",
// 			args: args{
// 				args: []any{"/v2/library/busybox/blobs/uploads/sha256:xxx", "DS$library/busybox$blob_uploads$*"},
// 			},
// 			want:    true,
// 			wantErr: false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := dal.UrlMatchFunc(tt.args.args...)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("urlMatchFunc() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("urlMatchFunc() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
