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
	"encoding/json"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/go-sigma/sigma/pkg/api/enums"
	sigmatelemetry "github.com/go-sigma/sigma/pkg/telemetry"
)

type envelope struct {
	Payload      json.RawMessage   `json:"payload"`
	TraceContext map[string]string `json:"trace_context,omitempty"`
}

// MarshalPayload serializes a task payload with the current trace context.
func MarshalPayload(ctx context.Context, payload any) ([]byte, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return json.Marshal(envelope{
		Payload:      raw,
		TraceContext: sigmatelemetry.CarrierFromContext(ctx),
	})
}

// UnmarshalPayload unwraps a task payload and restores the stored trace context.
func UnmarshalPayload(ctx context.Context, data []byte) (context.Context, []byte) {
	var msg envelope
	if err := json.Unmarshal(data, &msg); err != nil || len(msg.Payload) == 0 {
		return ctx, data
	}
	return sigmatelemetry.ContextWithCarrier(ctx, msg.TraceContext), msg.Payload
}

// StartConsumerSpan marks the asynchronous work queue consume boundary.
func StartConsumerSpan(ctx context.Context, topic enums.Daemon) (context.Context, oteltrace.Span) {
	return otel.Tracer("github.com/go-sigma/sigma/pkg/infra/workq").Start(
		ctx,
		"workq.consume",
		oteltrace.WithSpanKind(oteltrace.SpanKindConsumer),
		oteltrace.WithAttributes(attribute.String("messaging.destination.name", topic.String())),
	)
}
