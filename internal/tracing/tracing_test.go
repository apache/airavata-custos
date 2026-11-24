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
	"testing"
	"time"
)

func TestInitProductionMintsValidIDs(t *testing.T) {
	shutdown, err := Init(InitConfig{ServiceName: "custos-test"})
	if err != nil {
		t.Fatalf("Init(production) returned error: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = shutdown(ctx)
	}()

	ctx, span := Start(context.Background(), "test.production")
	defer span.End()

	traceID, spanID, _ := IDsFromContext(ctx)
	if traceID == "" || spanID == "" {
		t.Fatalf("expected non-empty IDs from production ctx, got trace=%q span=%q", traceID, spanID)
	}
}

func TestStartCapturesParentSpanID(t *testing.T) {
	shutdown, err := Init(InitConfig{ServiceName: "custos-test"})
	if err != nil {
		t.Fatalf("Init returned error: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = shutdown(ctx)
	}()

	ctx1, span1 := Start(context.Background(), "outer")
	defer span1.End()
	span1ID := span1.SpanContext().SpanID()

	ctx2, span2 := Start(ctx1, "inner")
	defer span2.End()

	got := *ParentSpanIDFromContext(ctx2)
	if len(got) != 16 {
		t.Fatalf("expected 16 hex chars, got %d", len(got))
	}
	if got != span1ID.String() {
		t.Errorf("parent mismatch: want %s got %s", span1ID.String(), got)
	}

	if root := ParentSpanIDFromContext(ctx1); root != nil {
		t.Errorf("expected empty parent on root span ctx, got %s", *root)
	}
}
