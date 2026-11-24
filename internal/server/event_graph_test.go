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

package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetEventGraph(t *testing.T) {
	if _, err := eventGraph.ReadFile(eventGraphFile); err != nil {
		t.Skip("run go generate ./cmd/server/")
	}
	rec := httptest.NewRecorder()
	(&Server{}).getEventGraph(rec, httptest.NewRequest(http.MethodGet, "/events/graph", nil))
	var g struct{ Events, Edges []json.RawMessage }
	if err := json.Unmarshal(rec.Body.Bytes(), &g); err != nil || rec.Code != http.StatusOK || len(g.Events) == 0 || len(g.Edges) == 0 {
		t.Fatalf("status %d, %d events, %d edges, %v", rec.Code, len(g.Events), len(g.Edges), err)
	}
}
