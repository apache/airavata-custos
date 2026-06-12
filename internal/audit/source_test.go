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

package audit

import (
	"context"
	"testing"
)

func TestSourceFromContext_DefaultEmpty(t *testing.T) {
	if got := SourceFromContext(context.Background()); got != "" {
		t.Fatalf("expected empty source, got %q", got)
	}
}

func TestWithSource_StampsValue(t *testing.T) {
	ctx := WithSource(context.Background(), "comanage")
	if got := SourceFromContext(ctx); got != "comanage" {
		t.Fatalf("got %q, want comanage", got)
	}
}

func TestWithSource_EmptyDoesNotStamp(t *testing.T) {
	ctx := WithSource(context.Background(), "")
	if got := SourceFromContext(ctx); got != "" {
		t.Fatalf("empty source should not stamp ctx, got %q", got)
	}
}

func TestWithSource_OverwritesPrior(t *testing.T) {
	ctx := WithSource(context.Background(), "amie")
	ctx = WithSource(ctx, "comanage")
	if got := SourceFromContext(ctx); got != "comanage" {
		t.Fatalf("expected newer value to win, got %q", got)
	}
}
