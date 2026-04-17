// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package httputil provides shared HTTP utilities like context keys and JSON helpers.
package httputil

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/apache/airavata-custos/signer/internal/store"
)

type contextKey string

const (
	clientConfigKey contextKey = "clientConfig"
	sourceIPKey     contextKey = "sourceIP"
	userIdentityKey contextKey = "userIdentity"
)

type UserIdentityContext struct {
	Issuer    string
	Subject   string
	Email     string
	Principal string
}

func WithUserIdentity(ctx context.Context, id *UserIdentityContext) context.Context {
	return context.WithValue(ctx, userIdentityKey, id)
}

func UserIdentityFromContext(ctx context.Context) *UserIdentityContext {
	if v := ctx.Value(userIdentityKey); v != nil {
		return v.(*UserIdentityContext)
	}
	return nil
}

func WithClientConfig(ctx context.Context, cfg *store.ClientConfig) context.Context {
	return context.WithValue(ctx, clientConfigKey, cfg)
}

func ClientConfigFromContext(ctx context.Context) *store.ClientConfig {
	if v := ctx.Value(clientConfigKey); v != nil {
		return v.(*store.ClientConfig)
	}
	return nil
}

func WithSourceIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, sourceIPKey, ip)
}

func SourceIPFromContext(ctx context.Context) string {
	if v := ctx.Value(sourceIPKey); v != nil {
		return v.(string)
	}
	return ""
}

func WriteJSONError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   code,
		"message": message,
	})
}

func WriteJSONErrorWithExtra(w http.ResponseWriter, status int, code, message string, extra map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp := map[string]string{
		"error":   code,
		"message": message,
	}
	for k, v := range extra {
		resp[k] = v
	}
	json.NewEncoder(w).Encode(resp)
}
