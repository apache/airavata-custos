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

// Package httputil holds small HTTP helpers shared across the binary
// (tracing middleware, logging middleware, admin handlers).
package httputil

import "net/http"

// StatusRecorder captures the response status without buffering the body.
type StatusRecorder struct {
	http.ResponseWriter
	Status int
}

func (sr *StatusRecorder) WriteHeader(code int) {
	sr.Status = code
	sr.ResponseWriter.WriteHeader(code)
}
