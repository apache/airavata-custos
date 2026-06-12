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

// Package audit carries cross-cutting audit metadata on ctx so writers do not
// have to pass it through every call site.
package audit

import "context"

type sourceKeyType struct{}

var sourceKey sourceKeyType

// WithSource tags ctx with the source name. Connector subscribers and HTTP
// entry points set this once; downstream audit writes pick it up automatically.
func WithSource(ctx context.Context, source string) context.Context {
	if source == "" {
		return ctx
	}
	return context.WithValue(ctx, sourceKey, source)
}

// SourceFromContext returns the source tagged on ctx, or "" if none.
func SourceFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(sourceKey).(string); ok {
		return v
	}
	return ""
}
