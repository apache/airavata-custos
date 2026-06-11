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
	"encoding/json"
	"time"
)

type TraceSummary struct {
	TraceID       string    `db:"trace_id" json:"trace_id"`
	RootOperation string    `db:"root_operation" json:"root_operation"`
	Source        string    `db:"source" json:"source"`
	Status        string    `db:"status" json:"status"`
	StartedAt     time.Time `db:"started_at" json:"started_at"`
	EndedAt       time.Time `db:"ended_at" json:"ended_at"`
	EventCount    int       `db:"event_count" json:"event_count"`
}

type TraceEvent struct {
	SpanID       string    `db:"span_id" json:"span_id"`
	ParentSpanID string    `db:"parent_span_id" json:"parent_span_id,omitempty"`
	Source       string    `db:"source" json:"source"`
	EventType    string    `db:"event_type" json:"event_type"`
	EntityType   string    `db:"entity_type" json:"entity_type,omitempty"`
	EntityID     string    `db:"entity_id" json:"entity_id,omitempty"`
	Description  string    `db:"description" json:"description,omitempty"`
	Status       string    `db:"status" json:"status"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

type TraceNode struct {
	TraceEvent
	Children []*TraceNode `json:"children"`
}

// MarshalJSON ensures Children is emitted as [] (not null) so clients
// can iterate without a nil check.
func (n TraceNode) MarshalJSON() ([]byte, error) {
	type alias TraceNode
	if n.Children == nil {
		n.Children = []*TraceNode{}
	}
	return json.Marshal(alias(n))
}
