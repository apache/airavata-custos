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
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/apache/airavata-custos/internal/httputil"
)

// Middleware opens a root span per request and writes X-Trace-Id on the response.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := r.Pattern
		if route == "" {
			route = r.URL.Path
		}
		name := "http." + r.Method + " " + route

		ctx, span := Start(r.Context(), name)
		defer span.End()

		span.SetAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.route", route),
			attribute.String("source", "http"),
		)

		if tid, _ := IDsFromContext(ctx); tid != "" {
			w.Header().Set("X-Trace-Id", tid)
		}

		sw := &httputil.StatusRecorder{ResponseWriter: w, Status: http.StatusOK}

		// Re-panic so net/http's recover still logs and serves 500.
		defer func() {
			if rec := recover(); rec != nil {
				span.RecordError(fmt.Errorf("http handler panic: %v", rec))
				span.SetStatus(codes.Error, "panic")
				panic(rec)
			}
		}()

		next.ServeHTTP(sw, r.WithContext(ctx))

		span.SetAttributes(attribute.Int("http.status_code", sw.Status))
		if sw.Status >= 500 {
			span.SetStatus(codes.Error, http.StatusText(sw.Status))
		}
	})
}
