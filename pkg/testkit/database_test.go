// Copyright 2026 sigma
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

package testkit

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/go-sigma/sigma/pkg/api/enums"
)

func TestNewDatabaseName(t *testing.T) {
	namePattern := regexp.MustCompile(`^[a-z][a-z0-9_]+$`)
	seen := make(map[string]struct{}, 100)

	for range 100 {
		name := NewDatabaseName()
		if len(name) > 63 {
			t.Fatalf("database name length = %d, want <= 63", len(name))
		}
		if !namePattern.MatchString(name) {
			t.Fatalf("database name %q is not identifier-safe", name)
		}
		if _, ok := seen[name]; ok {
			t.Fatalf("database name %q is duplicated", name)
		}
		seen[name] = struct{}{}
	}
}

func TestDatabaseTypeDefault(t *testing.T) {
	value, exists := os.LookupEnv(EnvDatabaseType)
	require.NoError(t, os.Unsetenv(EnvDatabaseType))
	t.Cleanup(func() {
		if exists {
			require.NoError(t, os.Setenv(EnvDatabaseType, value))
			return
		}
		require.NoError(t, os.Unsetenv(EnvDatabaseType))
	})

	databaseType, err := DatabaseType(enums.DatabaseSqlite3)
	require.NoError(t, err)
	require.Equal(t, enums.DatabaseSqlite3, databaseType)

	databaseType, err = DatabaseType(enums.DatabasePostgresql)
	require.NoError(t, err)
	require.Equal(t, enums.DatabasePostgresql, databaseType)

	_, err = DatabaseType("")
	require.Error(t, err)
}

func TestDatabaseTypeFromEnvironment(t *testing.T) {
	for _, databaseType := range []enums.Database{
		enums.DatabaseSqlite3,
		enums.DatabaseTurso,
		enums.DatabaseMysql,
		enums.DatabasePostgresql,
	} {
		t.Run(databaseType.String(), func(t *testing.T) {
			t.Setenv(EnvDatabaseType, databaseType.String())

			actual, err := DatabaseType(enums.DatabaseSqlite3)
			require.NoError(t, err)
			require.Equal(t, databaseType, actual)
		})
	}
}

func TestDatabaseTypeInvalid(t *testing.T) {
	for _, value := range []string{"", " ", "oracle"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv(EnvDatabaseType, value)
			_, err := DatabaseType(enums.DatabaseSqlite3)
			require.Error(t, err)
		})
	}
}

func TestPrepareSQLiteDatabase(t *testing.T) {
	first, err := PrepareDatabase(t, enums.DatabaseSqlite3)
	require.NoError(t, err)
	second, err := PrepareDatabase(t, enums.DatabaseSqlite3)
	require.NoError(t, err)

	require.NotEqual(t, first.Sqlite3.Path, second.Sqlite3.Path)
	require.Contains(t, first.Sqlite3.Path, "mode=memory")
}

func TestPrepareTursoDatabase(t *testing.T) {
	databaseConfig, err := PrepareDatabase(t, enums.DatabaseTurso)
	require.NoError(t, err)

	require.Equal(t, enums.DatabaseTurso, databaseConfig.Type)
	require.Equal(t, ".db", filepath.Ext(databaseConfig.Turso.DSN))
	require.DirExists(t, filepath.Dir(databaseConfig.Turso.DSN))
}
