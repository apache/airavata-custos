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

package main

import (
	"path/filepath"
	"testing"
)

// TestEdgesLeaveConnectorHandlers analyses the repository and checks every edge
// starts at a connector handler subscribed to its source event.
func TestEdgesLeaveConnectorHandlers(t *testing.T) {
	repoRoot, _ = filepath.Abs("../..")
	st := analyse()
	handlers := map[[2]string]bool{}
	for _, s := range st.subs {
		if s.Connector != "core" {
			handlers[[2]string{s.Topic, st.connName(s.Connector) + "." + actor(s.Handler)}] = true
		}
	}
	g := st.graph()
	subs := map[string]int{}
	for _, e := range g.Events {
		subs[e.ID] = e.Subs
	}
	if len(g.Edges) == 0 {
		t.Fatal("no edges")
	}
	for _, e := range g.Edges {
		switch {
		case e.Kind != "conn":
			t.Errorf("%s → %s: kind %q, want conn", e.From, e.To, e.Kind)
		case subs[e.From] == 0:
			t.Errorf("%s → %s: edge out of a leaf", e.From, e.To)
		case !handlers[[2]string{e.From, e.Actor}]:
			t.Errorf("%s → %s: %s does not handle %s", e.From, e.To, e.Actor, e.From)
		}
	}
}
