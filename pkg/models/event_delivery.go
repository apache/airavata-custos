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

// EventSubscription records that a subscriber listens to an event type.
type EventSubscription struct {
	Subscriber string    `json:"subscriber" db:"subscriber" binding:"required"`
	EventType  string    `json:"event_type" db:"event_type" binding:"required"`
	CreatedAt  time.Time `json:"created_at" db:"created_at" binding:"required"`
}

// Event is one published event. It is stored only when it has a subscriber.
type Event struct {
	ID        string          `json:"id" db:"id" binding:"required"`
	EventType string          `json:"event_type" db:"event_type" binding:"required"`
	Payload   json.RawMessage `json:"payload" db:"payload" swaggertype:"object" binding:"required"`
	Source    string          `json:"source" db:"source" binding:"required"` // Subsystem that published the event.
	TraceID   string          `json:"trace_id" db:"trace_id" binding:"required"`
	SpanID    string          `json:"span_id" db:"span_id" binding:"required"` // Publisher's span, the parent of its deliveries.
	CreatedAt time.Time       `json:"created_at" db:"created_at" binding:"required"`
}

type EventDeliveryStatus string

const (
	EventDeliveryPending   EventDeliveryStatus = "PENDING"
	EventDeliverySucceeded EventDeliveryStatus = "SUCCEEDED"
	EventDeliveryFailed    EventDeliveryStatus = "FAILED"
)

// EventDelivery is one event to be handled by the corresponding subscriber.
type EventDelivery struct {
	ID         string              `json:"id" db:"id" binding:"required"`
	EventID    string              `json:"event_id" db:"event_id" binding:"required"`
	Subscriber string              `json:"subscriber" db:"subscriber" binding:"required"`
	Status     EventDeliveryStatus `json:"status" db:"status" binding:"required"`
	Attempts   int                 `json:"attempts" db:"attempts" binding:"required"`
	NextRunAt  time.Time           `json:"next_run_at" db:"next_run_at" binding:"required"`
	LastError  *string             `json:"last_error,omitempty" db:"last_error"`
	CreatedAt  time.Time           `json:"created_at" db:"created_at" binding:"required"`
	FinishedAt *time.Time          `json:"finished_at,omitempty" db:"finished_at"`
}

// PendingDelivery is a delivery joined with its event.
type PendingDelivery struct {
	EventDelivery
	Event Event `json:"event" db:"event" binding:"required"`
}

// DeliveryHistory is event deliveries with its audit rows.
type DeliveryHistory struct {
	PendingDelivery
	History []AuditEvent `json:"history" extensions:"x-nullable"`
}
