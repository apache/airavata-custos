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

package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateFolder(t *testing.T) {
	tests := []struct {
		name         string
		create, stat int
		wantErr      bool
	}{
		{name: "created", create: 200},
		{name: "already exists", create: 503, stat: 200},
		{name: "unknown owner", create: 503, stat: 503, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.URL.Path == "/api/token/":
					_, _ = w.Write([]byte(`{"access":"fresh"}`))
				case r.Header.Get("Authorization") != "Bearer fresh":
					w.WriteHeader(http.StatusForbidden)
				case r.URL.Path == "/api/folders/stat_path/":
					w.WriteHeader(tt.stat)
				default:
					w.WriteHeader(tt.create)
				}
			}))
			defer srv.Close()
			err := New(srv.URL, "u", "p", 1).CreateFolder(context.Background(), "/home/custos-jdoe", "custos-jdoe", "custos-jdoe", 0o700)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err=%v", err)
			}
		})
	}
}
