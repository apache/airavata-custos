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
	TraceID       string         `db:"trace_id" json:"trace_id"`
	RootOperation string         `db:"root_operation" json:"root_operation"`
	Source        string         `db:"source" json:"source"`
	Status        string         `db:"status" json:"status"`
	StartedAt     time.Time      `db:"started_at" json:"started_at"`
	EndedAt       time.Time      `db:"ended_at" json:"ended_at"`
	EventCount    int            `db:"event_count" json:"event_count"`
	Deliveries    DeliveryCounts `db:"-" json:"deliveries"`
}

// DeliveryCounts says how many of a trace's deliveries are pending, done and
// failed. Attempts is the most tries any pending one has made.
type DeliveryCounts struct {
	Pending   int `json:"pending"`
	Succeeded int `json:"succeeded"`
	Failed    int `json:"failed"`
	Attempts  int `json:"attempts"`
}

// TraceDelivery is a connector's delivery of an event the trace published.
// SpanID is the span of the step that published the event, so the delivery
// can be shown under that step.
type TraceDelivery struct {
	ID         string              `db:"id" json:"id"`
	TraceID    string              `db:"trace_id" json:"-"`
	EventType  string              `db:"event_type" json:"event_type"`
	Subscriber string              `db:"subscriber" json:"subscriber"`
	Status     EventDeliveryStatus `db:"status" json:"status"`
	Attempts   int                 `db:"attempts" json:"attempts"`
	NextRunAt  time.Time           `db:"next_run_at" json:"next_run_at"`
	LastError  *string             `db:"last_error" json:"last_error,omitempty"`
	SpanID     string              `db:"span_id" json:"span_id"`
}

type TraceEvent struct {
	ID           string    `db:"id" json:"id"`
	SpanID       string    `db:"span_id" json:"span_id"`
	ParentSpanID *string   `db:"parent_span_id" json:"parent_span_id,omitempty"`
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
