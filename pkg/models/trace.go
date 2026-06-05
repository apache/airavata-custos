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

package models

import (
	"encoding/hex"
	"encoding/json"
	"time"
)

type TraceSummary struct {
	TraceID       []byte    `db:"trace_id" json:"-"`
	RootOperation string    `db:"root_operation" json:"root_operation"`
	Source        string    `db:"source" json:"source"`
	Status        string    `db:"status" json:"status"`
	StartedAt     time.Time `db:"started_at" json:"started_at"`
	EndedAt       time.Time `db:"ended_at" json:"ended_at"`
	EventCount    int       `db:"event_count" json:"event_count"`
}

func (t TraceSummary) MarshalJSON() ([]byte, error) {
	type alias TraceSummary
	return json.Marshal(struct {
		TraceID string `json:"trace_id"`
		alias
	}{
		TraceID: hex.EncodeToString(t.TraceID),
		alias:   alias(t),
	})
}

type TraceEvent struct {
	SpanID       []byte    `db:"span_id" json:"-"`
	ParentSpanID []byte    `db:"parent_span_id" json:"-"`
	Source       string    `db:"source" json:"source"`
	Action       string    `db:"action" json:"action"`
	EntityType   string    `db:"entity_type" json:"entity_type,omitempty"`
	EntityID     string    `db:"entity_id" json:"entity_id,omitempty"`
	Summary      string    `db:"summary" json:"summary,omitempty"`
	Status       string    `db:"status" json:"status"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

func (e TraceEvent) MarshalJSON() ([]byte, error) {
	type alias TraceEvent
	out := struct {
		SpanID       string `json:"span_id"`
		ParentSpanID string `json:"parent_span_id,omitempty"`
		alias
	}{
		SpanID: hex.EncodeToString(e.SpanID),
		alias:  alias(e),
	}
	if len(e.ParentSpanID) > 0 {
		out.ParentSpanID = hex.EncodeToString(e.ParentSpanID)
	}
	return json.Marshal(out)
}

type TraceNode struct {
	TraceEvent
	Children []*TraceNode `json:"children"`
}

// Children is always emitted (non-nil) so clients can iterate without a check.
func (n TraceNode) MarshalJSON() ([]byte, error) {
	type alias TraceEvent
	out := struct {
		SpanID       string `json:"span_id"`
		ParentSpanID string `json:"parent_span_id,omitempty"`
		alias
		Children []*TraceNode `json:"children"`
	}{
		SpanID:   hex.EncodeToString(n.SpanID),
		alias:    alias(n.TraceEvent),
		Children: n.Children,
	}
	if len(n.ParentSpanID) > 0 {
		out.ParentSpanID = hex.EncodeToString(n.ParentSpanID)
	}
	if out.Children == nil {
		out.Children = []*TraceNode{}
	}
	return json.Marshal(out)
}
