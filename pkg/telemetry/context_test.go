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

package telemetry

import (
	"context"
	"encoding/json/v2"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// setPropagator installs the composite propagator the package expects and
// restores the previous one when the test finishes.
func setPropagator(t *testing.T) {
	t.Helper()
	prev := otel.GetTextMapPropagator()
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	t.Cleanup(func() { otel.SetTextMapPropagator(prev) })
}

// recordingContext returns a context carrying a sampled span and the trace id
// of that span.
func recordingContext(t *testing.T) (context.Context, string) {
	t.Helper()
	tp := sdktrace.NewTracerProvider()
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })

	ctx, span := tp.Tracer("telemetry-test").Start(t.Context(), "op")
	t.Cleanup(func() { span.End() })
	return ctx, span.SpanContext().TraceID().String()
}

func TestTraceIDFromContext(t *testing.T) {
	require.Empty(t, TraceIDFromContext(t.Context()))

	ctx, want := recordingContext(t)
	require.Equal(t, want, TraceIDFromContext(ctx))
}

func TestCarrierFromContext(t *testing.T) {
	setPropagator(t)

	require.Nil(t, CarrierFromContext(t.Context()))

	ctx, want := recordingContext(t)
	carrier := CarrierFromContext(ctx)
	require.NotNil(t, carrier)
	require.Contains(t, carrier, "traceparent")
	require.Equal(t, want, TraceIDFromCarrier(carrier))
}

func TestCarrierFromContextWithoutPropagator(t *testing.T) {
	prev := otel.GetTextMapPropagator()
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator())
	t.Cleanup(func() { otel.SetTextMapPropagator(prev) })

	ctx, _ := recordingContext(t)
	require.Nil(t, CarrierFromContext(ctx))
}

func TestContextWithCarrier(t *testing.T) {
	setPropagator(t)

	base := t.Context()
	require.Equal(t, base, ContextWithCarrier(base, nil))
	require.Equal(t, base, ContextWithCarrier(base, map[string]string{}))

	ctx, want := recordingContext(t)
	got := TraceIDFromContext(ContextWithCarrier(base, CarrierFromContext(ctx)))
	require.Equal(t, want, got)
}

func TestTraceIDFromCarrier(t *testing.T) {
	setPropagator(t)

	require.Empty(t, TraceIDFromCarrier(nil))

	ctx, want := recordingContext(t)
	require.Equal(t, want, TraceIDFromCarrier(CarrierFromContext(ctx)))
}

func TestTraceIDFromCarrierBytes(t *testing.T) {
	setPropagator(t)

	require.Empty(t, TraceIDFromCarrierBytes([]byte("not-json")))

	ctx, want := recordingContext(t)
	data, err := json.Marshal(CarrierFromContext(ctx))
	require.NoError(t, err)
	require.Equal(t, want, TraceIDFromCarrierBytes(data))
}

func TestClearCarrier(t *testing.T) {
	setPropagator(t)

	empty := map[string]string{}
	require.Equal(t, empty, ClearCarrier(empty))

	carrier := map[string]string{
		"Traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
		"BAGGAGE":     "hello=world",
		"tracestate":  "vendor=value",
		"user-agent":  "sigma",
	}
	require.Equal(t, map[string]string{"user-agent": "sigma"}, ClearCarrier(carrier))
}

func TestInjectCarrier(t *testing.T) {
	setPropagator(t)

	// Without an active span, stale propagation headers are dropped.
	got := InjectCarrier(t.Context(), map[string]string{
		"traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
		"user-agent":  "sigma",
	})
	require.Equal(t, map[string]string{"user-agent": "sigma"}, got)

	// Without an active span and a nil carrier, nothing is created.
	require.Nil(t, InjectCarrier(t.Context(), nil))

	// With an active span and a nil carrier, a new carrier is created.
	ctx, want := recordingContext(t)
	got = InjectCarrier(ctx, nil)
	require.Equal(t, want, TraceIDFromCarrier(got))
}

func TestInjectCarrierClearsStalePropagationHeaders(t *testing.T) {
	setPropagator(t)

	carrier := map[string]string{
		"Traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
		"baggage":     "hello=world",
		"user-agent":  "sigma",
	}

	carrier = InjectCarrier(t.Context(), carrier)

	require.Equal(t, map[string]string{"user-agent": "sigma"}, carrier)
}
