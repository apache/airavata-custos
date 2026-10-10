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
	ID           string    `json:"id" db:"id" binding:"required"`
	EventType    string    `json:"event_type" db:"event_type" binding:"required"` // e.g., "COMPUTE_ALLOCATION_CREATED", "COMPUTE_ALLOCATION_UPDATED", "COMPUTE_ALLOCATION_DELETED", etc.
	EventTime    time.Time `json:"event_time" db:"event_time" binding:"required"`
	EntityID     string    `json:"entity_id" db:"entity_id" binding:"required"`     // ID of the entity the event is about.
	EntityType   string    `json:"entity_type" db:"entity_type" binding:"required"` // Kind of the entity ("user", "role", "compute_cluster_user", "packet", etc.).
	Details      string    `json:"details" db:"details" binding:"required"`         // Additional details about the event, stored as a JSON string or plain text.
	Source       string    `json:"source" db:"source" binding:"required"`           // Subsystem that produced the event (e.g., "amie", "comanage", "slurm", "core", etc.).
	TraceID      string    `json:"trace_id" db:"trace_id" binding:"required"`
	SpanID       string    `json:"span_id" db:"span_id" binding:"required"`
	ParentSpanID *string   `json:"parent_span_id,omitempty" db:"parent_span_id"`
}
