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

package workq

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	oteltrace "go.opentelemetry.io/otel/trace"

	sigmatelemetry "github.com/go-sigma/sigma/pkg/telemetry"
)

func TestMarshalPayloadCarriesTraceContext(t *testing.T) {
	otel.SetTextMapPropagator(propagation.TraceContext{})

	traceID := oteltrace.TraceID{0x4b, 0xf9, 0x2f, 0x35, 0x77, 0xb3, 0x4d, 0xa6, 0xa3, 0xce, 0x92, 0x9d, 0x0e, 0x0e, 0x47, 0x36}
	spanID := oteltrace.SpanID{0x00, 0xf0, 0x67, 0xaa, 0x0b, 0xa9, 0x02, 0xb7}
	spanContext := oteltrace.NewSpanContext(oteltrace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: oteltrace.FlagsSampled,
		Remote:     true,
	})
	ctx := oteltrace.ContextWithRemoteSpanContext(t.Context(), spanContext)

	data, err := MarshalPayload(ctx, map[string]string{"hello": "world"})
	require.NoError(t, err)

	ctx, payload := UnmarshalPayload(context.Background(), data)
	require.JSONEq(t, `{"hello":"world"}`, string(payload))
	require.Equal(t, traceID.String(), sigmatelemetry.TraceIDFromContext(ctx))
}

func TestUnmarshalPayloadKeepsLegacyPayload(t *testing.T) {
	payload := []byte(`{"hello":"world"}`)

	ctx, got := UnmarshalPayload(t.Context(), payload)

	require.Equal(t, payload, got)
	require.Empty(t, sigmatelemetry.TraceIDFromContext(ctx))
}
