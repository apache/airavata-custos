// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package tracing

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

type Mode int

const (
	ModeProduction Mode = iota
	ModeNoop
)

const tracerName = "custos"

type InitConfig struct {
	Mode        Mode
	Logger      *slog.Logger
	ServiceName string
}

func Init(cfg InitConfig) (func(context.Context) error, error) {
	if cfg.Mode == ModeNoop {
		otel.SetTracerProvider(noop.NewTracerProvider())
		otel.SetTextMapPropagator(propagation.TraceContext{})
		return func(context.Context) error { return nil }, nil
	}

	serviceName := cfg.ServiceName
	if serviceName == "" {
		serviceName = "custos"
	}

	hostname, _ := os.Hostname()
	instanceID := fmt.Sprintf("%s-%d", hostname, os.Getpid())

	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceInstanceID(instanceID),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("tracing: build resource: %w", err)
	}

	// No SpanProcessor: spans live in ctx for ID propagation only.
	tp := sdktrace.NewTracerProvider(sdktrace.WithResource(res))

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return func(ctx context.Context) error {
		return tp.Shutdown(ctx)
	}, nil
}

type parentSpanIDKeyType struct{}
type lastBusinessSpanIDKeyType struct{}

var (
	parentSpanIDKey       parentSpanIDKeyType
	lastBusinessSpanIDKey lastBusinessSpanIDKeyType
)

// Start opens a span and stamps the audit parent in ctx. bus.* spans are
// skipped so audit parents jump over the bus to the nearest business span.
func Start(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	newCtx, span := otel.Tracer(tracerName).Start(ctx, name, opts...)

	if strings.HasPrefix(name, "bus.") {
		return newCtx, span
	}

	if last, ok := ctx.Value(lastBusinessSpanIDKey).(trace.SpanID); ok && last.IsValid() {
		newCtx = context.WithValue(newCtx, parentSpanIDKey, last)
	} else if parent := trace.SpanFromContext(ctx).SpanContext().SpanID(); parent.IsValid() {
		newCtx = context.WithValue(newCtx, parentSpanIDKey, parent)
	}

	newCtx = context.WithValue(newCtx, lastBusinessSpanIDKey, span.SpanContext().SpanID())
	return newCtx, span
}

func ParentSpanIDFromContext(ctx context.Context) []byte {
	if p, ok := ctx.Value(parentSpanIDKey).(trace.SpanID); ok && p.IsValid() {
		out := make([]byte, 8)
		copy(out, p[:])
		return out
	}
	return nil
}

func FromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

func IDsFromContext(ctx context.Context) (traceID, spanID string) {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return "", ""
	}
	return sc.TraceID().String(), sc.SpanID().String()
}

// IDsBytesFromContext returns the active trace_id (16 bytes) and span_id (8 bytes),
// or nil/nil if no recording span is on ctx.
func IDsBytesFromContext(ctx context.Context) (traceID, spanID []byte) {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return nil, nil
	}
	tid := sc.TraceID()
	sid := sc.SpanID()
	return append([]byte(nil), tid[:]...), append([]byte(nil), sid[:]...)
}
