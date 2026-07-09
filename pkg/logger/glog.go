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
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"strings"
	"time"

	"gorm.io/gorm/logger"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/config"
)

// ZLogger is the logger for gorm
type ZLogger struct{}

// LogMode is the log mode
func (l ZLogger) LogMode(logger.LogLevel) logger.Interface {
	return l
}

// Error is the error log
func (l ZLogger) Error(ctx context.Context, msg string, opts ...any) {
	slog.ErrorContext(ctx, fmt.Sprintf(msg, opts...))
}

// Warn is the warn log
func (l ZLogger) Warn(ctx context.Context, msg string, opts ...any) {
	slog.WarnContext(ctx, fmt.Sprintf(msg, opts...))
}

// Info is the info log
func (l ZLogger) Info(ctx context.Context, msg string, opts ...any) {
	slog.InfoContext(ctx, fmt.Sprintf(msg, opts...))
}

// Trace is the trace log
func (l ZLogger) Trace(ctx context.Context, begin time.Time, f func() (string, int64), err error) {
	args := []any{"call", databaseTraceCall(), "elapsed", time.Since(begin).String()}

	sql, rows := f()
	if sql != "" {
		args = append(args, "sql", sql)
	}
	if rows > -1 {
		args = append(args, "rows", rows)
	}
	logDatabaseTrace(ctx, args...)
}

func databaseTraceCall() string {
	logLevel := config.GetConfig().Log.Level
	if logLevel == enums.LogLevelDebug || logLevel == enums.LogLevelTrace {
		for i := range 15 {
			_, file, n, ok := runtime.Caller(i)
			if !ok {
				break
			}
			if strings.HasPrefix(file, "github.com/go-sigma/sigma/pkg") {
				if strings.HasPrefix(file, "github.com/go-sigma/sigma/pkg/dal/query") ||
					strings.HasPrefix(file, "github.com/go-sigma/sigma/pkg/dal/repository/registry") ||
					strings.HasPrefix(file, "github.com/go-sigma/sigma/pkg/logger") {
					continue
				}
				lastIndex := strings.LastIndex(file, "/")
				left := file[:lastIndex]
				file = file[lastIndex+1:]
				if strings.Contains(left, "/") {
					lastIndex = strings.LastIndex(left, "/")
					file = left[lastIndex+1:] + "/" + file
				}
				return fmt.Sprintf("%s:%d", file, n)
			}
		}
	}
	return "database"
}

func logDatabaseTrace(ctx context.Context, args ...any) {
	if !slog.Default().Enabled(ctx, slog.LevelDebug) {
		return
	}
	record := slog.NewRecord(time.Now(), slog.LevelDebug, "call database", 0)
	record.Add(args...)
	_ = slog.Default().Handler().Handle(ctx, record)
}
