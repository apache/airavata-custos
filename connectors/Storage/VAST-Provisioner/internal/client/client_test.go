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
	"encoding/json"
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

// Make sure the view is created on the given path with the policy and protocols, and creating it again when it already exists is fine.
func TestCreateView(t *testing.T) {
	tests := []struct {
		name     string
		create   int
		existing string
		wantErr  bool
	}{
		{name: "created", create: 200},
		{name: "already exists", create: 400, existing: `[{"path":"/scratch/custos-jdoe"}]`},
		{name: "refused and no view", create: 400, existing: `[]`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sent map[string]any
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.URL.Path == "/api/token/":
					_, _ = w.Write([]byte(`{"access":"fresh"}`))
				case r.Method == http.MethodGet:
					_, _ = w.Write([]byte(tt.existing))
				default:
					_ = json.NewDecoder(r.Body).Decode(&sent)
					w.WriteHeader(tt.create)
				}
			}))
			defer srv.Close()
			err := New(srv.URL, "u", "p", 1).CreateView(context.Background(), "/scratch/custos-jdoe", 4, []string{"NFS", "NFS4"})
			if (err != nil) != tt.wantErr {
				t.Fatalf("err=%v", err)
			}
			want := `{"create_dir":false,"path":"/scratch/custos-jdoe","policy_id":4,"protocols":["NFS","NFS4"],"tenant_id":1}`
			if got, _ := json.Marshal(sent); string(got) != want {
				t.Fatalf("sent %s, want %s", got, want)
			}
		})
	}
}

// Make sure the view policy is matched by its exact name, and a missing one is an error.
func TestViewPolicyID(t *testing.T) {
	tests := []struct {
		name, policies string
		want           int
		wantErr        bool
	}{
		{name: "found", policies: `[{"id":4,"name":"local-access"}]`, want: 4},
		{name: "missing", policies: `[]`, wantErr: true},
		{name: "only a similar name", policies: `[{"id":5,"name":"local-access-old"}]`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/token/" {
					_, _ = w.Write([]byte(`{"access":"fresh"}`))
					return
				}
				_, _ = w.Write([]byte(tt.policies))
			}))
			defer srv.Close()
			id, err := New(srv.URL, "u", "p", 1).ViewPolicyID(context.Background(), "local-access")
			if id != tt.want || (err != nil) != tt.wantErr {
				t.Fatalf("id=%d err=%v", id, err)
			}
		})
	}
}
