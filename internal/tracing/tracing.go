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
	"errors"
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
)

const tracerName = "custos"

// provider mints the spans whose ids audit rows carry, so it is a real SDK
// provider from the start rather than the otel global, which is a no-op until set.
var provider trace.TracerProvider = sdktrace.NewTracerProvider()

// SetProvider replaces the span provider, for Init and for tests that record
// spans, and returns the previous one.
func SetProvider(tp trace.TracerProvider) (prev trace.TracerProvider) {
	prev, provider = provider, tp
	return prev
}

type InitConfig struct {
	Logger      *slog.Logger
	ServiceName string
}

func Init(cfg InitConfig) (func(context.Context) error, error) {
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

	SetProvider(tp)
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
	newCtx, span := provider.Tracer(tracerName).Start(ctx, name, opts...)

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

func ParentSpanIDFromContext(ctx context.Context) *string {
	if p, ok := ctx.Value(parentSpanIDKey).(trace.SpanID); ok && p.IsValid() {
		s := p.String()
		return &s
	}
	return nil
}

// ContextWithSpanContext resumes a saved span, so the spans opened from here
// on are its children and audit rows written here take it as their parent.
// An invalid one leaves ctx unchanged and is logged since it should not happen.
func ContextWithSpanContext(ctx context.Context, traceID, spanID string) context.Context {
	tid, terr := trace.TraceIDFromHex(traceID)
	sid, serr := trace.SpanIDFromHex(spanID)
	if err := errors.Join(terr, serr); err != nil {
		slog.WarnContext(ctx, "ignoring invalid span context", "trace_id", traceID, "span_id", spanID, "error", err)
		return ctx
	}
	ctx = context.WithValue(ctx, parentSpanIDKey, sid)
	return trace.ContextWithRemoteSpanContext(ctx, trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    tid,
		SpanID:     sid,
		TraceFlags: trace.FlagsSampled,
		Remote:     true,
	}))
}

// ErrNoSpan rejects a write made outside a span, since every audit row and event needs a trace.
var ErrNoSpan = errors.New("tracing: no active span")

func IDsFromContext(ctx context.Context) (traceID, spanID string, err error) {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return "", "", ErrNoSpan
	}
	return sc.TraceID().String(), sc.SpanID().String(), nil
}
