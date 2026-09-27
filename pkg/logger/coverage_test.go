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

package logger

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/config"
)

func TestSetLevelAllBranches(t *testing.T) {
	previous := slog.Default()
	t.Cleanup(func() { slog.SetDefault(previous) })

	for _, level := range []string{"trace", "debug", "info", "warn", "error", "fatal", "panic", "unknown"} {
		t.Run(level, func(t *testing.T) {
			SetLevel(level)
			require.NotNil(t, slog.Default())
		})
	}
}

func TestTrimSourceFileBranches(t *testing.T) {
	require.Equal(t, "helper.go", trimSourceFile("helper.go"))
	require.Equal(t, "app/helper.go", trimSourceFile("app/helper.go"))
	require.Equal(t, "app/helper.go", trimSourceFile("github.com/go-sigma/sigma/pkg/app/helper.go"))
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }

func TestLogWriterWrite(t *testing.T) {
	var buf bytes.Buffer
	written, err := logWriter{Writer: &buf}.Write([]byte("plain line\n"))
	require.NoError(t, err)
	require.Equal(t, len("plain line\n"), written)
	require.Equal(t, "plain line\n", buf.String())

	_, err = logWriter{Writer: failingWriter{}}.Write([]byte("plain line\n"))
	require.Error(t, err)
}

func TestFormatLogLineErrorBranches(t *testing.T) {
	require.Equal(t, "no sql here\n", formatLogLine("no sql here\n"))
	require.Equal(t, `x sql="unterminated`, formatLogLine(`x sql="unterminated`))
	require.Equal(t, "x sql=\"\\q\"\n", formatLogLine("x sql=\"\\q\"\n"))
}

func TestLogDatabaseTraceDisabled(t *testing.T) {
	previous := slog.Default()
	t.Cleanup(func() { slog.SetDefault(previous) })
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelInfo})))

	logDatabaseTrace(context.Background(), "sql", "select 1")
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

func TestLoggerMethods(t *testing.T) {
	l := &Logger{}
	l.Debug("debug")
	l.Info("info")
	l.Warn("warn")
	l.Error("error")
}

func TestLoggerFatalExits(t *testing.T) {
	if os.Getenv("SIGMA_TEST_LOGGER_FATAL") == "1" {
		(&Logger{}).Fatal("fatal logger")
		return
	}

	//nolint:gosec // re-exec the test binary to observe the exit code
	cmd := exec.Command(os.Args[0], "-test.run=TestLoggerFatalExits")
	cmd.Env = append(os.Environ(), "SIGMA_TEST_LOGGER_FATAL=1")
	out, err := cmd.CombinedOutput()
	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, 1, exitErr.ExitCode())
	require.Contains(t, string(out), "fatal logger")
}

func TestFatalExits(t *testing.T) {
	if os.Getenv("SIGMA_TEST_FATAL") == "1" {
		Fatal("fatal package")
		return
	}

	//nolint:gosec // re-exec the test binary to observe the exit code
	cmd := exec.Command(os.Args[0], "-test.run=TestFatalExits")
	cmd.Env = append(os.Environ(), "SIGMA_TEST_FATAL=1")
	out, err := cmd.CombinedOutput()
	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, 1, exitErr.ExitCode())
	require.Contains(t, string(out), "fatal package")
}

func TestTraceHandler(t *testing.T) {
	var buf bytes.Buffer
	handler := NewTraceHandler(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	require.True(t, handler.Enabled(context.Background(), slog.LevelInfo))

	// Without a trace id the record is passed through unchanged.
	require.NoError(t, handler.Handle(context.Background(), slog.NewRecord(time.Now(), slog.LevelInfo, "no-trace", 0)))
	require.NotContains(t, buf.String(), TraceKey)

	buf.Reset()
	tp := sdktrace.NewTracerProvider()
	ctx, span := tp.Tracer("logger-test").Start(context.Background(), "op")
	defer span.End()
	require.NoError(t, handler.Handle(ctx, slog.NewRecord(time.Now(), slog.LevelInfo, "with-trace", 0)))
	require.Contains(t, buf.String(), TraceKey)
	require.Contains(t, buf.String(), span.SpanContext().TraceID().String())

	require.NotNil(t, handler.WithAttrs([]slog.Attr{slog.String("key", "value")}))
	require.NotNil(t, handler.WithGroup("group"))
}
