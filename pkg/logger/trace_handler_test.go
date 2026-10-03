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
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestTraceHandler(t *testing.T) {
	var buf bytes.Buffer
	handler := NewTraceHandler(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	require.True(t, handler.Enabled(t.Context(), slog.LevelInfo))

	// Without a trace id the record is passed through unchanged.
	require.NoError(t, handler.Handle(t.Context(), slog.NewRecord(time.Now(), slog.LevelInfo, "no-trace", 0)))
	require.NotContains(t, buf.String(), TraceKey)

	buf.Reset()
	tp := sdktrace.NewTracerProvider()
	ctx, span := tp.Tracer("logger-test").Start(t.Context(), "op")
	defer span.End()
	require.NoError(t, handler.Handle(ctx, slog.NewRecord(time.Now(), slog.LevelInfo, "with-trace", 0)))
	require.Contains(t, buf.String(), TraceKey)
	require.Contains(t, buf.String(), span.SpanContext().TraceID().String())

	require.NotNil(t, handler.WithAttrs([]slog.Attr{slog.String("key", "value")}))
	require.NotNil(t, handler.WithGroup("group"))
}
