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
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func newTestClient(t *testing.T, baseURL string) *Client {
	t.Helper()
	return New(Config{
		RegistryURL:   baseURL,
		COID:          2,
		APIUser:       "co_2.testuser",
		APIKey:        "key-primary",
		UnixClusterID: 1,
		HTTPTimeout:   5 * time.Second,
	})
}

func TestCreatePerson_SendsBasicAuthAndDecodes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || !strings.HasSuffix(r.URL.Path, "/api/co/2/core/v1/people") {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		user, pass, ok := r.BasicAuth()
		if !ok || user != "co_2.testuser" || pass != "key-primary" {
			t.Errorf("auth header: ok=%v user=%q pass=%q", ok, user, pass)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `[{"identifier":"Person100099","type":"comanage_id","login":false,"status":"A"}]`)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL+"/registry")
	c.cfg.RegistryURL = srv.URL // override to match server root
	out, err := c.CreatePerson([]byte(`{}`))
	if err != nil {
		t.Fatalf("CreatePerson: %v", err)
	}
	if len(out) != 1 || out[0].Identifier != "Person100099" || out[0].Type != "comanage_id" {
		t.Errorf("got %+v", out)
	}
}

func TestDo_RetriesOn5xx(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "ok")
	}))
	defer srv.Close()

	c := New(Config{
		RegistryURL: srv.URL,
		COID:        2,
		APIUser:     "u",
		APIKey:      "k",
		HTTPTimeout: 2 * time.Second,
	})
	resp, body, err := c.Do("GET", srv.URL+"/anything", nil)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: got %d", resp.StatusCode)
	}
	if got := string(body); got != "ok" {
		t.Errorf("body: %q", got)
	}
	if calls.Load() != 3 {
		t.Errorf("calls: got %d, want 3", calls.Load())
	}
}

func TestCreateCoGroup_EnvelopeShape(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/co_groups.json") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var req CoGroupCreateRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.RequestType != "CoGroups" || req.Version != "1.0" || len(req.CoGroups) != 1 {
			t.Errorf("envelope: %+v", req)
		}
		g := req.CoGroups[0]
		if g.CoId != 2 || g.Name != "custos-test" || g.GroupType != "CL" || g.Auto {
			t.Errorf("group fields: %+v", g)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"ResponseType":"NewObject","Version":"1.0","ObjectType":"CoGroup","Id":"42"}`)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	c.cfg.RegistryURL = srv.URL
	id, err := c.CreateCoGroup("custos-test", "Primary group for custos-test")
	if err != nil {
		t.Fatalf("CreateCoGroup: %v", err)
	}
	if id != 42 {
		t.Errorf("id: got %d, want 42", id)
	}
}

func TestCreateIdentifierOnGroup_UsesNestedPersonShape(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req IdentifierCreateRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("decode: %v", err)
		}
		ident := req.Identifiers[0]
		if ident.Person.Type != "Group" || ident.Person.Id != 25 {
			t.Errorf("Person field: %+v (need {Type:Group,Id:25})", ident.Person)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"ResponseType":"NewObject","Version":"1.0","ObjectType":"Identifier","Id":"7"}`)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	c.cfg.RegistryURL = srv.URL
	id, err := c.CreateIdentifierOnGroup("custos-alice", "uid", 25)
	if err != nil {
		t.Fatalf("CreateIdentifierOnGroup: %v", err)
	}
	if id != 7 {
		t.Errorf("id: %d", id)
	}
}

func TestCreateUnixClusterGroup_URLAndBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/unix_cluster/unix_cluster_groups.json") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var req UnixClusterGroupCreateRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("decode: %v", err)
		}
		one := req.UnixClusterGroups[0]
		if one.UnixClusterId != 1 || one.CoGroupId != 25 {
			t.Errorf("fields: %+v", one)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"ResponseType":"NewObject","Version":"1.0","ObjectType":"UnixClusterGroup","Id":"3"}`)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	c.cfg.RegistryURL = srv.URL
	id, err := c.CreateUnixClusterGroup(25)
	if err != nil {
		t.Fatalf("CreateUnixClusterGroup: %v", err)
	}
	if id != 3 {
		t.Errorf("id: %d", id)
	}
}

func TestFindUnixClusterGroup_MatchesConfiguredCluster(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("cogroupid") != "25" {
			t.Errorf("missing cogroupid filter: %s", r.URL.RawQuery)
		}
		_, _ = io.WriteString(w, `{
            "ResponseType":"UnixClusterGroups","Version":"1.0",
            "UnixClusterGroups":[
                {"Version":"1.0","Id":7,"UnixClusterId":9,"CoGroupId":25},
                {"Version":"1.0","Id":8,"UnixClusterId":1,"CoGroupId":25}
            ]
        }`)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	c.cfg.RegistryURL = srv.URL

	id, err := c.FindUnixClusterGroup(25)
	if err != nil {
		t.Fatalf("FindUnixClusterGroup: %v", err)
	}
	if id != 8 {
		t.Errorf("id: got %d, want 8 (UnixClusterId=1 matches configured cluster)", id)
	}
}

func TestFindUnixClusterGroup_NoMatchReturnsZero(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"ResponseType":"UnixClusterGroups","Version":"1.0","UnixClusterGroups":[]}`)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	c.cfg.RegistryURL = srv.URL

	id, err := c.FindUnixClusterGroup(99)
	if err != nil {
		t.Fatalf("FindUnixClusterGroup: %v", err)
	}
	if id != 0 {
		t.Errorf("id: got %d, want 0", id)
	}
}

func TestGetPersonComposite_PreservesRawJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || !strings.Contains(r.URL.Path, "/api/co/2/core/v1/people/Person100099") {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"CoPerson":{"meta":{"id":99}},"Name":[],"Identifier":[]}`)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	c.cfg.RegistryURL = srv.URL
	raw, err := c.GetPersonComposite("Person100099")
	if err != nil {
		t.Fatalf("GetPersonComposite: %v", err)
	}
	if !strings.Contains(string(raw), `"id":99`) {
		t.Errorf("raw missing expected substring: %s", string(raw))
	}
}

func TestFindCoGroupByName_ExactMatchOnly(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{
            "ResponseType":"CoGroups","Version":"1.0",
            "CoGroups":[
                {"Version":"1.0","Id":11,"CoId":2,"Name":"custos-alice"},
                {"Version":"1.0","Id":12,"CoId":2,"Name":"custos-bob"}
            ]
        }`)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	c.cfg.RegistryURL = srv.URL

	id, err := c.FindCoGroupByName("custos-bob")
	if err != nil {
		t.Fatalf("FindCoGroupByName: %v", err)
	}
	if id != 12 {
		t.Errorf("id: got %d, want 12", id)
	}
	id, err = c.FindCoGroupByName("custos-nope")
	if err != nil {
		t.Fatalf("FindCoGroupByName miss: %v", err)
	}
	if id != 0 {
		t.Errorf("miss should return 0, got %d", id)
	}
}
