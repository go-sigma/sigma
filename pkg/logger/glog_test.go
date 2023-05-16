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

package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type MockWriter struct {
	Entries []map[string]interface{}
}

func NewMockWriter() *MockWriter {
	return &MockWriter{make([]map[string]interface{}, 0)}
}

func (m *MockWriter) Write(p []byte) (int, error) {
	entry := map[string]interface{}{}

	if err := json.Unmarshal(p, &entry); err != nil {
		panic(fmt.Sprintf("Failed to parse JSON %v: %s", p, err.Error()))
	}

	m.Entries = append(m.Entries, entry)

	return len(p), nil
}

func (m *MockWriter) Reset() {
	m.Entries = make([]map[string]interface{}, 0)
}

func Test_Logger_Sqlite(t *testing.T) {
	logger := NewMockWriter()

	z := zerolog.New(logger)

	now := time.Now()

	zLogger := ZLogger{}

	assert.NotNil(t, zLogger.LogMode(0))

	zLogger.Error(context.Background(), "error %s", "error")
	zLogger.Warn(context.Background(), "warn %s", "warn")
	zLogger.Info(context.Background(), "info %s", "info")

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{NowFunc: func() time.Time { return now }, Logger: zLogger})

	if err != nil {
		panic(err)
	}

	db = db.WithContext(z.WithContext(context.Background()))

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
		logger.Reset()

		err := c.run()

		if err != nil && !c.err_ok {
			t.Fatalf("Unexpected error: %s (%T)", err, err)
		}

		// TODO: Must get from log entries
		entries := logger.Entries

		if got, want := len(entries), 1; got != want {
			t.Errorf("Logger logged %d items, want %d items", got, want)
		} else {
			fieldByName := entries[0]

			if got, want := fieldByName["sql"].(string), c.sql; got != want {
				t.Errorf("Logged sql was %q, want %q", got, want)
			}
		}
	}
}
