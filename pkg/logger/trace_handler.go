// Copyright 2025 sigma
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
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

// TraceKey is the attribute name under which the trace id is recorded.
const TraceKey = "trace_id"

// traceHandler wraps a slog.Handler and adds the OpenTelemetry trace id
// from the request context to every log record that has one.
type traceHandler struct {
	inner slog.Handler
}

// NewTraceHandler wraps the given handler so that the trace id carried in
// the logging context (if any) is added to each record.
func NewTraceHandler(h slog.Handler) slog.Handler {
	return &traceHandler{inner: h}
}

// Enabled reports whether the handler will process records at the given level.
func (h *traceHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

// Handle extracts the trace id from ctx and attaches it to the record
// before delegating to the wrapped handler.
func (h *traceHandler) Handle(ctx context.Context, record slog.Record) error {
	sc := trace.SpanContextFromContext(ctx)
	if sc.HasTraceID() {
		record.AddAttrs(slog.String(TraceKey, sc.TraceID().String()))
	}
	return h.inner.Handle(ctx, record)
}

// WithAttrs returns a new handler whose attributes consist of both the
// receiver's attributes and the argument list.
func (h *traceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &traceHandler{inner: h.inner.WithAttrs(attrs)}
}

// WithGroup returns a handler with the given group appended to existing groups.
func (h *traceHandler) WithGroup(name string) slog.Handler {
	return &traceHandler{inner: h.inner.WithGroup(name)}
}
