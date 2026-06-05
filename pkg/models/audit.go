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

import "time"

type AuditEvent struct {
	ID           string    `json:"id" db:"id"`
	EventType    string    `json:"event_type" db:"event_type"` // e.g., "COMPUTE_ALLOCATION_CREATED", "COMPUTE_ALLOCATION_UPDATED", "COMPUTE_ALLOCATION_DELETED", etc.
	EventTime    time.Time `json:"event_time" db:"event_time"`
	EntityID     string    `json:"entity_id" db:"entity_id"` // The ID of the entity associated with the event, e.g., the compute allocation ID.
	Details      string    `json:"details" db:"details"`     // Additional details about the event, stored as a JSON string or plain text.
	Source       string    `json:"source" db:"source"`       // The connector / subsystem that produced the event (e.g., "amie", "comanage", "slurm", "core").
	TraceID      []byte    `json:"-" db:"trace_id"`
	SpanID       []byte    `json:"-" db:"span_id"`
	ParentSpanID []byte    `json:"-" db:"parent_span_id"`
}
