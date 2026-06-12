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

package store

import (
	"strings"
	"testing"
	"time"
)

func mkSpan(b byte) string {
	hi := "0123456789abcdef"[b>>4]
	lo := "0123456789abcdef"[b&0x0f]
	return strings.Repeat(string(hi)+string(lo), 8)
}

func TestBuildTreeNestsByParent(t *testing.T) {
	root := mkSpan(0x01)
	child := mkSpan(0x02)
	leaf := mkSpan(0x03)
	rows := []rowEvent{
		{SpanID: root, Source: "amie", EventType: "CREATE_PERSON", CreatedAt: time.Unix(1, 0)},
		{SpanID: child, ParentSpanID: root, Source: "comanage", EventType: "ComanageLookup", CreatedAt: time.Unix(2, 0)},
		{SpanID: leaf, ParentSpanID: child, Source: "comanage", EventType: "ComanageClusterAccountAttached", CreatedAt: time.Unix(3, 0)},
	}
	tree := buildTree(rows)
	if got := len(tree.Children); got != 1 {
		t.Fatalf("top level = %d, want 1", got)
	}
	rootNode := tree.Children[0]
	if rootNode.EventType != "CREATE_PERSON" {
		t.Errorf("root event_type = %q", rootNode.EventType)
	}
	if got := len(rootNode.Children); got != 1 {
		t.Fatalf("root children = %d, want 1", got)
	}
	if got := len(rootNode.Children[0].Children); got != 1 {
		t.Fatalf("grandchild count = %d, want 1", got)
	}
	if rootNode.Children[0].Children[0].EventType != "ComanageClusterAccountAttached" {
		t.Errorf("leaf event_type = %q", rootNode.Children[0].Children[0].EventType)
	}
}

func TestBuildTreeOrphansBecomeTopLevel(t *testing.T) {
	row1 := mkSpan(0x10)
	ghost := mkSpan(0xEE)
	row2 := mkSpan(0x11)
	rows := []rowEvent{
		{SpanID: row1, Source: "amie", EventType: "ROOT"},
		// parent references a span that did not write an audit row
		{SpanID: row2, ParentSpanID: ghost, Source: "amie", EventType: "ORPHAN"},
	}
	tree := buildTree(rows)
	if got := len(tree.Children); got != 2 {
		t.Fatalf("top level = %d, want 2", got)
	}
}

func TestEventStatusViaToTraceEvent(t *testing.T) {
	row := rowEvent{EventType: "ComanageProvisioningFailed"}
	if got := row.toTraceEvent().Status; got != "error" {
		t.Errorf("status = %q, want error", got)
	}
	row2 := rowEvent{EventType: "CREATE_PERSON"}
	if got := row2.toTraceEvent().Status; got != "ok" {
		t.Errorf("status = %q, want ok", got)
	}
}

func TestBuildTraceWhereEmpty(t *testing.T) {
	w, args := buildTraceWhere(TraceFilter{})
	if w != "" {
		t.Errorf("empty filter where = %q", w)
	}
	if len(args) != 0 {
		t.Errorf("empty filter args = %v", args)
	}
}

func TestBuildTraceWhereSources(t *testing.T) {
	w, args := buildTraceWhere(TraceFilter{Sources: []string{"amie", "comanage"}})
	if !strings.Contains(w, "u.source IN (?,?)") {
		t.Errorf("where = %q, missing IN clause", w)
	}
	if len(args) != 2 {
		t.Errorf("args len = %d, want 2", len(args))
	}
}

func TestBuildTraceWhereTimeAndQ(t *testing.T) {
	from := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	to := from.Add(24 * time.Hour)
	w, args := buildTraceWhere(TraceFilter{From: from, To: to, Q: "create"})
	if !strings.Contains(w, "u.created_at >= ?") || !strings.Contains(w, "u.created_at <= ?") {
		t.Errorf("where missing time clauses: %q", w)
	}
	if !strings.Contains(w, "u.trace_id LIKE ?") {
		t.Errorf("where missing trace_id LIKE: %q", w)
	}
	// 2 timestamps + 2 LIKE values
	if len(args) != 4 {
		t.Errorf("args len = %d, want 4", len(args))
	}
}

func TestMatchesStatusFilter(t *testing.T) {
	if !matchesStatusFilter("ok", nil) {
		t.Error("empty filter should match")
	}
	if !matchesStatusFilter("error", []string{"ok", "error"}) {
		t.Error("match should succeed")
	}
	if matchesStatusFilter("in_progress", []string{"ok"}) {
		t.Error("filter mismatch should reject")
	}
}

func TestDominantSourcePicksRoot(t *testing.T) {
	rows := []rowEvent{
		{Source: "comanage", ParentSpanID: mkSpan(0x01)},
		{Source: "amie"}, // root: nil parent
	}
	if got := dominantSource(rows); got != "amie" {
		t.Errorf("dominant = %q, want amie", got)
	}
}
