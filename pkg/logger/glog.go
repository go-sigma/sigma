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

// LevelTrace is the log level below Debug. It is used for high frequency
// polling queries and is only enabled when the configured log level is trace.
const LevelTrace = slog.LevelDebug - 4

// databaseTraceQuietKey is the context key marking a context whose database
// queries are logged at LevelTrace instead of Debug.
type databaseTraceQuietKey struct{}

// WithQuietDatabaseTrace returns a context that downgrades the database trace
// logs of the queries executed with it from debug to trace level. Long running
// pollers use it to keep the debug log readable.
func WithQuietDatabaseTrace(ctx context.Context) context.Context {
	return context.WithValue(ctx, databaseTraceQuietKey{}, true)
}

// databaseTraceQuiet reports whether ctx was marked by WithQuietDatabaseTrace.
func databaseTraceQuiet(ctx context.Context) bool {
	quiet, ok := ctx.Value(databaseTraceQuietKey{}).(bool)
	return ok && quiet
}

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
	if !slog.Default().Enabled(ctx, databaseTraceLevel(ctx)) {
		return
	}
	args := []any{"call", databaseTraceCall(ctx), "elapsed", time.Since(begin).String()}

	sql, rows := f()
	if sql != "" {
		args = append(args, "sql", sql)
	}
	if rows > -1 {
		args = append(args, "rows", rows)
	}
	logDatabaseTrace(ctx, args...)
}

// databaseTraceCall reports the caller site used for gorm trace logs. When ctx
// is marked quiet, or the configured log level is not debug/trace, it returns
// "database" without walking the stack.
func databaseTraceCall(ctx context.Context) string {
	if databaseTraceQuiet(ctx) {
		return "database"
	}
	return databaseTraceCallFrom(runtime.Caller)
}

// databaseTraceCallFrom is the testable core of databaseTraceCall. caller is
// runtime.Caller in production and a fake in tests.
func databaseTraceCallFrom(caller func(skip int) (uintptr, string, int, bool)) string {
	logLevel := config.GetConfig().Log.Level
	if logLevel == enums.LogLevelDebug || logLevel == enums.LogLevelTrace {
		for i := range 15 {
			_, file, n, ok := caller(i)
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

// databaseTraceLevel returns the level used for database trace logs. Queries
// issued by pollers are downgraded one level below debug so that they only show
// up when trace logging is enabled.
func databaseTraceLevel(ctx context.Context) slog.Level {
	if databaseTraceQuiet(ctx) {
		return LevelTrace
	}
	return slog.LevelDebug
}

func logDatabaseTrace(ctx context.Context, args ...any) {
	level := databaseTraceLevel(ctx)
	if !slog.Default().Enabled(ctx, level) {
		return
	}
	record := slog.NewRecord(time.Now(), level, "call database", 0)
	record.Add(args...)
	_ = slog.Default().Handler().Handle(ctx, record)
}
