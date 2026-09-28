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
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSetLevel(t *testing.T) {
	SetLevel("debug")
	slog.Info("log level set to debug", "x", "x")
	SetLevel("info")
	slog.Error("log level set to info", "x", "x", "err", fmt.Errorf("hello"))
}

func TestReplaceLogAttrPreservesTime(t *testing.T) {
	ts := time.Date(2026, 7, 7, 21, 8, 34, 668000000, time.FixedZone("CST", 8*60*60))

	attr := replaceLogAttr(nil, slog.Time(slog.TimeKey, ts))

	require.Equal(t, slog.TimeKey, attr.Key)
	require.Equal(t, slog.KindTime, attr.Value.Kind())
	require.Equal(t, ts, attr.Value.Time())
}

func TestReplaceLogAttrTrimsSourceFile(t *testing.T) {
	attr := replaceLogAttr(nil, slog.Any(slog.SourceKey, &slog.Source{
		File: "github.com/go-sigma/sigma/pkg/app/helper.go",
		Line: 74,
	}))

	source, ok := attr.Value.Any().(*slog.Source)
	require.True(t, ok)
	require.Equal(t, "app/helper.go", source.File)
	require.Equal(t, 74, source.Line)
}

func TestFormatLogLineQuotesSQLWithSingleQuote(t *testing.T) {
	line := `time=2026-07-07T21:20:42 level=DEBUG msg="call database" sql="SELECT * FROM \"settings\" WHERE \"settings\".\"key\" = 'signing.private_key'" rows=0` + "\n"

	formatted := formatLogLine(line)

	require.Contains(t, formatted, `sql='SELECT * FROM "settings" WHERE "settings"."key" = \'signing.private_key\''`)
	require.NotContains(t, formatted, `\"settings\"`)
}

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
