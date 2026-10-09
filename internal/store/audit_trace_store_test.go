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

	"github.com/apache/airavata-custos/internal/tracing"
	"github.com/apache/airavata-custos/pkg/models"
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
		{SpanID: child, ParentSpanID: &root, Source: "comanage", EventType: "ComanageLookup", CreatedAt: time.Unix(2, 0)},
		{SpanID: leaf, ParentSpanID: &child, Source: "comanage", EventType: "ComanageClusterAccountAttached", CreatedAt: time.Unix(3, 0)},
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
		{SpanID: row2, ParentSpanID: &ghost, Source: "amie", EventType: "ORPHAN"},
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
	if !strings.Contains(w, "source = ANY(?)") {
		t.Errorf("where = %q, missing ANY clause", w)
	}
	if len(args) != 1 {
		t.Errorf("args len = %d, want 1", len(args))
	}
}

func TestBuildTraceWhereTimeAndQ(t *testing.T) {
	from := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	to := from.Add(24 * time.Hour)
	w, args := buildTraceWhere(TraceFilter{From: from, To: to, Q: "create"})
	if !strings.Contains(w, "event_time >= ?") || !strings.Contains(w, "event_time <= ?") {
		t.Errorf("where missing time clauses: %q", w)
	}
	if !strings.Contains(w, "trace_id ILIKE ?") || !strings.Contains(w, "entity_id = ?") || !strings.Contains(w, "details ILIKE ?") {
		t.Errorf("where missing a search clause: %q", w)
	}
	// 2 timestamps + 4 search values
	if len(args) != 6 {
		t.Errorf("args len = %d, want 6", len(args))
	}
}

// Make sure the trace status comes from the deliveries, so a delivery that succeeded after failed tries reads ok, not error.
func TestTraceStatusComesFromDeliveries(t *testing.T) {
	publisher, try1, handler1, try2 := mkSpan(0x01), mkSpan(0x02), mkSpan(0x03), mkSpan(0x04)
	rows := []rowEvent{
		{SpanID: publisher, EventType: "CLUSTER_ACCOUNT_APPROVED"},
		{SpanID: handler1, ParentSpanID: &try1, EventType: "ComanageProvisioningFailed"},
		{SpanID: try1, ParentSpanID: &publisher, EventType: "EVENT_DELIVERY_FAILED", EntityType: deliveryEntityType},
		{SpanID: try2, ParentSpanID: &publisher, EventType: "EVENT_DELIVERY_SUCCEEDED", EntityType: deliveryEntityType},
	}
	outside := append(rows, rowEvent{SpanID: mkSpan(0x05), ParentSpanID: &publisher, EventType: "PosixUsernameBuildFailed"})
	delivery := func(status models.EventDeliveryStatus) []models.TraceDelivery {
		return []models.TraceDelivery{{Status: status, Attempts: 2}}
	}
	cases := []struct {
		name       string
		rows       []rowEvent
		deliveries []models.TraceDelivery
		want       string
	}{
		{"delivery out of tries", rows, delivery(models.EventDeliveryFailed), tracing.StatusError},
		{"delivery still retrying", rows, delivery(models.EventDeliveryPending), tracing.StatusInProgress},
		{"delivery recovered on a later try", rows, delivery(models.EventDeliverySucceeded), tracing.StatusOk},
		{"step failed outside any delivery", outside, delivery(models.EventDeliverySucceeded), tracing.StatusError},
		{"nothing delivered, nothing failed", rows[:1], nil, tracing.StatusOk},
	}
	for _, c := range cases {
		if got := traceStatus(c.rows, c.deliveries); got != c.want {
			t.Errorf("%s: status = %q, want %q", c.name, got, c.want)
		}
	}
}

// Make sure the list gets how many deliveries are pending, done and failed,
// and the most tries a pending one has made, which it shows as "retrying N of 10".
func TestDeliveryCounts(t *testing.T) {
	got := deliveryCounts([]models.TraceDelivery{
		{Status: models.EventDeliveryPending, Attempts: 4},
		{Status: models.EventDeliveryPending, Attempts: 1},
		{Status: models.EventDeliverySucceeded, Attempts: 1},
		{Status: models.EventDeliveryFailed, Attempts: 10},
	})
	want := models.DeliveryCounts{Pending: 2, Succeeded: 1, Failed: 1, Attempts: 4}
	if got != want {
		t.Errorf("counts = %+v, want %+v", got, want)
	}
}

// The root is the first row whose parent span wrote no row, so a trace
// resumed from an unaudited span (an AMIE ingest, a publisher) still has one.
func TestSummarisePicksRootWithUnauditedParent(t *testing.T) {
	ingest, received := mkSpan(0x01), mkSpan(0x02)
	rows := []rowEvent{
		{SpanID: received, ParentSpanID: &ingest, Source: "amie", EventType: "PACKET_RECEIVED"},
		{SpanID: mkSpan(0x03), ParentSpanID: &received, Source: "comanage", EventType: "ComanageLookup"},
	}
	if op, src := summarise(rows); op != "PACKET_RECEIVED" || src != "amie" {
		t.Errorf("root = %s/%s, want PACKET_RECEIVED/amie", op, src)
	}
}
