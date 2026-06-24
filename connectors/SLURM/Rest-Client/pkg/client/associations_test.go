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

// cli/internal/client/associations_test.go
package client

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestListAssociationsByAccount(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/slurmdb/v0.0.41/associations" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if got, _ := url.QueryUnescape(r.URL.RawQuery); got != "account=eng" {
			t.Fatalf("query = %q", got)
		}
		_, _ = w.Write([]byte(`{"associations":[{"account":"eng","cluster":"artisan","user":"alice","id_association":5}]}`))
	}))
	defer srv.Close()
	c := New(srv.URL, "root", "t", "41")
	assocs, err := c.ListAssociations(AssocFilter{Account: "eng"})
	if err != nil {
		t.Fatal(err)
	}
	if len(assocs) != 1 || assocs[0].User != "alice" || assocs[0].ID != 5 {
		t.Errorf("assocs = %+v", assocs)
	}
}

func TestCreateAssociation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Support to list associations so we can verify the association was created successfully
		if r.Method == "GET" && r.URL.Path == "/slurmdb/v0.0.41/associations" {
			q := r.URL.Query()
			if q.Get("account") != "eng" || q.Get("cluster") != "artisan" || q.Get("user") != "alice" {
				t.Errorf("unexpected list query = %q", r.URL.RawQuery)
				http.Error(w, "bad query", http.StatusBadRequest)
				return
			}
			_, _ = w.Write([]byte(`{"associations":[{"account":"eng","cluster":"artisan","user":"alice","id_association":5}]}`))
			return
		}

		if r.Method != "POST" || r.URL.Path != "/slurmdb/v0.0.41/associations" {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		b, _ := io.ReadAll(r.Body)
		var payload struct {
			Associations []Association `json:"associations"`
		}
		_ = json.Unmarshal(b, &payload)
		if len(payload.Associations) != 1 || payload.Associations[0].Account != "eng" ||
			payload.Associations[0].User != "alice" {
			t.Errorf("payload = %+v", payload)
		}
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()
	c := New(srv.URL, "root", "t", "41")
	err := c.UpsertAssociation(Association{Account: "eng", Cluster: "artisan", User: "alice"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDeleteAssociation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" || r.URL.Path != "/slurmdb/v0.0.41/association" {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("account") != "eng" || r.URL.Query().Get("user") != "alice" {
			t.Fatalf("query = %v", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()
	c := New(srv.URL, "root", "t", "41")
	if err := c.DeleteAssociation(AssocFilter{Account: "eng", User: "alice"}); err != nil {
		t.Fatal(err)
	}
}
