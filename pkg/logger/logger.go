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
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/lmittmann/tint"

	"github.com/go-sigma/sigma/pkg/api/enums"
)

const logTimeFormat = "2006-01-02T15:04:05"

// SetLevel configures the global slog level and installs a TextHandler
// writing to stdout with source info.
func SetLevel(levelStr string) {
	var level slog.Level
	switch enums.LogLevel(levelStr) {
	case enums.LogLevelTrace, enums.LogLevelDebug:
		level = slog.LevelDebug
	case enums.LogLevelInfo:
		level = slog.LevelInfo
	case enums.LogLevelWarn:
		level = slog.LevelWarn
	case enums.LogLevelError, enums.LogLevelFatal, enums.LogLevelPanic:
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}
	inner := tint.NewTextHandler(logWriter{Writer: os.Stdout}, &tint.Options{
		Level:       level,
		AddSource:   true,
		ReplaceAttr: replaceLogAttr,
		TimeFormat:  logTimeFormat,
	})
	slog.SetDefault(slog.New(NewTraceHandler(inner)))
}

func replaceLogAttr(_ []string, attr slog.Attr) slog.Attr {
	if attr.Key == slog.SourceKey {
		if source, ok := attr.Value.Any().(*slog.Source); ok && source != nil {
			shortSource := *source
			shortSource.File = trimSourceFile(shortSource.File)
			return slog.Any(slog.SourceKey, &shortSource)
		}
	}
	return attr
}

func trimSourceFile(file string) string {
	lastSlash := strings.LastIndex(file, "/")
	if lastSlash < 0 {
		return file
	}
	prevSlash := strings.LastIndex(file[:lastSlash], "/")
	if prevSlash < 0 {
		return file
	}
	return file[prevSlash+1:]
}

type logWriter struct {
	io.Writer
}

func (w logWriter) Write(p []byte) (int, error) {
	formatted := formatLogLine(string(p))
	_, err := w.Writer.Write([]byte(formatted))
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func formatLogLine(line string) string {
	const sqlPrefix = ` sql="`

	start := strings.Index(line, sqlPrefix)
	if start < 0 {
		return line
	}
	valueStart := start + len(" sql=")
	valueEnd := findQuotedValueEnd(line, valueStart)
	if valueEnd < 0 {
		return line
	}
	sql, err := strconv.Unquote(line[valueStart : valueEnd+1])
	if err != nil {
		return line
	}
	return line[:valueStart] + "'" + escapeSQLLogValue(sql) + "'" + line[valueEnd+1:]
}

func findQuotedValueEnd(line string, start int) int {
	escaped := false
	for i := start + 1; i < len(line); i++ {
		switch {
		case escaped:
			escaped = false
		case line[i] == '\\':
			escaped = true
		case line[i] == '"':
			return i
		}
	}
	return -1
}

func escapeSQLLogValue(sql string) string {
	replacer := strings.NewReplacer(
		`\`, `\\`,
		`'`, `\'`,
		"\n", `\n`,
		"\r", `\r`,
		"\t", `\t`,
	)
	return replacer.Replace(sql)
}

// Logger is a thin stdlib-style wrapper used by pkg/infra/workq/redis.
type Logger struct{}

// Debug logs a message at Debug level.
func (l *Logger) Debug(args ...any) {
	slog.Debug(fmt.Sprintf("%v", args...))
}

// Info logs a message at Info level.
func (l *Logger) Info(args ...any) {
	slog.Info(fmt.Sprintf("%v", args...))
}

// Warn logs a message at Warning level.
func (l *Logger) Warn(args ...any) {
	slog.Warn(fmt.Sprintf("%v", args...))
}

// Error logs a message at Error level.
func (l *Logger) Error(args ...any) {
	slog.Error(fmt.Sprintf("%v", args...))
}

// Fatal logs a message at Error level and exits with status 1.
func (l *Logger) Fatal(args ...any) {
	slog.Error(fmt.Sprintf("%v", args...))
	os.Exit(1)
}

// Fatal is a package-level helper replacing the previous log.Fatal() usage.
// It logs at Error level and exits with status 1.
func Fatal(msg string, args ...any) {
	slog.Error(msg, args...)
	os.Exit(1)
}
