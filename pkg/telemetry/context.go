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
	"maps"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// TraceIDFromContext returns the active OpenTelemetry trace id from ctx.
func TraceIDFromContext(ctx context.Context) string {
	sc := oteltrace.SpanContextFromContext(ctx)
	if !sc.HasTraceID() {
		return ""
	}
	return sc.TraceID().String()
}

// CarrierFromContext serializes the active propagation context.
func CarrierFromContext(ctx context.Context) map[string]string {
	if TraceIDFromContext(ctx) == "" {
		return nil
	}
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	if len(carrier) == 0 {
		return nil
	}
	result := make(map[string]string, len(carrier))
	maps.Copy(result, carrier)
	return result
}

// ContextWithCarrier extracts a propagation carrier into ctx.
func ContextWithCarrier(ctx context.Context, carrier map[string]string) context.Context {
	if len(carrier) == 0 {
		return ctx
	}
	return otel.GetTextMapPropagator().Extract(ctx, propagation.MapCarrier(carrier))
}

// TraceIDFromCarrier returns the trace id encoded in a propagation carrier.
func TraceIDFromCarrier(carrier map[string]string) string {
	return TraceIDFromContext(ContextWithCarrier(context.Background(), carrier))
}

// TraceIDFromCarrierBytes returns the trace id encoded in a JSON propagation carrier.
func TraceIDFromCarrierBytes(data []byte) string {
	var carrier map[string]string
	if err := json.Unmarshal(data, &carrier); err != nil {
		return ""
	}
	return TraceIDFromCarrier(carrier)
}

// ClearCarrier removes propagation headers from carrier.
func ClearCarrier(carrier map[string]string) map[string]string {
	if len(carrier) == 0 {
		return carrier
	}
	fields := append([]string{"traceparent", "tracestate", "baggage"}, otel.GetTextMapPropagator().Fields()...)
	for _, field := range fields {
		for key := range carrier {
			if strings.EqualFold(key, field) {
				delete(carrier, key)
			}
		}
	}
	return carrier
}

// InjectCarrier adds the active propagation context to carrier.
func InjectCarrier(ctx context.Context, carrier map[string]string) map[string]string {
	carrier = ClearCarrier(carrier)
	traceCarrier := CarrierFromContext(ctx)
	if len(traceCarrier) == 0 {
		return carrier
	}
	if carrier == nil {
		carrier = make(map[string]string, len(traceCarrier))
	}
	maps.Copy(carrier, traceCarrier)
	return carrier
}
