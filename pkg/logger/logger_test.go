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
	"log/slog"
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
