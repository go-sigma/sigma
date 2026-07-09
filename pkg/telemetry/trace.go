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

package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	"github.com/go-sigma/sigma/pkg/config"
)

// Init configures the global tracer provider. Safe to call once at process start.
// If config.Trace.Enabled is false, it returns nil, nil (otel defaults to no-op).
func Init(cfg *config.Configuration) (func(context.Context) error, error) {
	if !cfg.Trace.Enabled {
		return nil, nil // tracing disabled; otel defaults to no-op
	}

	serviceName := cfg.Trace.ServiceName
	if serviceName == "" {
		serviceName = "sigma"
	}

	// 1. Exporter
	var exporter sdktrace.SpanExporter
	var err error
	switch cfg.Trace.Protocol {
	case "http", "":
		opts := []otlptracehttp.Option{otlptracehttp.WithEndpoint(cfg.Trace.Endpoint)}
		if cfg.Trace.Insecure {
			opts = append(opts, otlptracehttp.WithInsecure())
		}
		exporter, err = otlptracehttp.New(context.Background(), opts...)
	case "grpc":
		opts := []otlptracegrpc.Option{otlptracegrpc.WithEndpoint(cfg.Trace.Endpoint)}
		if cfg.Trace.Insecure {
			opts = append(opts, otlptracegrpc.WithInsecure())
		}
		exporter, err = otlptracegrpc.New(context.Background(), opts...)
	default:
		return nil, fmt.Errorf("unknown trace protocol %q", cfg.Trace.Protocol)
	}
	if err != nil {
		return nil, fmt.Errorf("create OTLP exporter: %w", err)
	}

	// 2. Resource (service.name, etc.)
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("build OTLP resource: %w", err)
	}

	// 3. Sampler
	ratio := cfg.Trace.SampleRatio
	if ratio <= 0 {
		ratio = 1.0
	}

	// 4. Provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(ratio))),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// 5. Shutdown function — caller defers this
	shutdown := func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		return tp.Shutdown(ctx)
	}
	return shutdown, nil
}
