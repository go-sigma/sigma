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
	"bytes"
	"encoding/json/v2"
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/config"
)

type MockWriter struct {
	Entries []map[string]any
	buf     bytes.Buffer
}

func NewMockWriter() *MockWriter {
	return &MockWriter{Entries: make([]map[string]any, 0)}
}

func (m *MockWriter) Write(p []byte) (int, error) {
	entry := map[string]any{}
	if err := json.Unmarshal(p, &entry); err != nil {
		panic(fmt.Sprintf("Failed to parse JSON %v: %s", p, err.Error()))
	}
	m.Entries = append(m.Entries, entry)
	return len(p), nil
}

func (m *MockWriter) Reset() {
	m.Entries = make([]map[string]any, 0)
	m.buf.Reset()
}

func Test_Logger_Sqlite(t *testing.T) {
	writer := NewMockWriter()
	prev := slog.Default()
	defer slog.SetDefault(prev)
	lg := slog.New(slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(lg)

	ctx := t.Context()

	now := time.Now()

	zLogger := ZLogger{}

	assert.NotNil(t, zLogger.LogMode(0))

	zLogger.Error(ctx, "error %s", "error")
	zLogger.Warn(ctx, "warn %s", "warn")
	zLogger.Info(ctx, "info %s", "info")

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{NowFunc: func() time.Time { return now }, Logger: zLogger})
	if err != nil {
		panic(err)
	}
	db = db.WithContext(ctx)

	type Post struct {
		Title, Body string
	}
	db.AutoMigrate(&Post{}) // nolint: errcheck

	cases := []struct {
		run    func() error
		sql    string
		err_ok bool
	}{
		{
			run: func() error { return db.Create(&Post{Title: "awesome"}).Error },
			sql: fmt.Sprintf(
				"INSERT INTO `posts` (`title`,`body`) VALUES (%q,%q)",
				"awesome", "",
			),
			err_ok: false,
		},
		{
			run:    func() error { return db.Model(&Post{}).Find(&[]*Post{}).Error },
			sql:    "SELECT * FROM `posts`",
			err_ok: false,
		},
		{
			run: func() error {
				return db.Where(&Post{Title: "awesome", Body: "This is awesome post !"}).First(&Post{}).Error
			},
			sql: fmt.Sprintf(
				"SELECT * FROM `posts` WHERE `posts`.`title` = %q AND `posts`.`body` = %q ORDER BY `posts`.`title` LIMIT 1",
				"awesome", "This is awesome post !",
			),
			err_ok: true,
		},
		{
			run:    func() error { return db.Raw("THIS is,not REAL sql").Scan(&Post{}).Error },
			sql:    "THIS is,not REAL sql",
			err_ok: true,
		},
	}

	for _, c := range cases {
		writer.Reset()

		err := c.run()
		if err != nil && !c.err_ok {
			t.Fatalf("Unexpected error: %s (%T)", err, err)
		}

		entries := writer.Entries
		if got, want := len(entries), 1; got != want {
			t.Errorf("Logger logged %d items, want %d items", got, want)
		} else {
			fieldByName := entries[0]
			if got, want := fieldByName["msg"].(string), "call database"; got != want {
				t.Errorf("Logged msg was %q, want %q", got, want)
			}
			if _, ok := fieldByName["source"]; ok {
				t.Errorf("Logged source was present, want omitted")
			}
			if _, ok := fieldByName["call"]; !ok {
				t.Errorf("Logged call was absent, want present")
			}
			if got, want := fieldByName["sql"].(string), c.sql; got != want {
				t.Errorf("Logged sql was %q, want %q", got, want)
			}
		}
	}
}

func TestLogDatabaseTraceDisabled(t *testing.T) {
	previous := slog.Default()
	t.Cleanup(func() { slog.SetDefault(previous) })
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelInfo})))

	logDatabaseTrace(t.Context(), "sql", "select 1")
}

func TestDatabaseTraceCallDebug(t *testing.T) {
	cfg := config.GetConfig()
	previous := cfg.Log.Level
	t.Cleanup(func() { cfg.Log.Level = previous })

	cfg.Log.Level = enums.LogLevelDebug
	require.NotEmpty(t, databaseTraceCall())

	cfg.Log.Level = enums.LogLevelTrace
	require.NotEmpty(t, databaseTraceCall())
}

func TestDatabaseTraceCallFrom(t *testing.T) {
	cfg := config.GetConfig()
	previous := cfg.Log.Level
	t.Cleanup(func() { cfg.Log.Level = previous })

	// At debug level, excluded frames are skipped and the first business frame wins.
	cfg.Log.Level = enums.LogLevelDebug
	frames := []string{
		"github.com/go-sigma/sigma/pkg/logger/glog.go",
		"github.com/go-sigma/sigma/pkg/dal/query/user.gen.go",
		"github.com/go-sigma/sigma/pkg/dal/repository/registry/artifact.go",
		"github.com/go-sigma/sigma/pkg/service/analytics/analytics.go",
	}
	index := 0
	caller := func(int) (uintptr, string, int, bool) {
		file := frames[index]
		index++
		return 0, file, index, true
	}
	require.Equal(t, "analytics/analytics.go:4", databaseTraceCallFrom(caller))

	// The walk stops when the stack is exhausted.
	empty := func(int) (uintptr, string, int, bool) { return 0, "", 0, false }
	require.Equal(t, "database", databaseTraceCallFrom(empty))

	// Above debug level the stack is never walked.
	cfg.Log.Level = enums.LogLevelInfo
	require.Equal(t, "database", databaseTraceCallFrom(func(int) (uintptr, string, int, bool) {
		t.Fatal("caller must not be invoked above debug level")
		return 0, "", 0, false
	}))
}
