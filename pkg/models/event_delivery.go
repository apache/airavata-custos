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
	Subscriber string    `json:"subscriber" db:"subscriber"`
	EventType  string    `json:"event_type" db:"event_type"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// Event is one published event. It is stored only when it has a subscriber.
type Event struct {
	ID        string          `json:"id" db:"id"`
	EventType string          `json:"event_type" db:"event_type"`
	Payload   json.RawMessage `json:"payload" db:"payload" swaggertype:"object"`
	Source    string          `json:"source" db:"source"` // Subsystem that published the event.
	TraceID   string          `json:"trace_id" db:"trace_id"`
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
}

type EventDeliveryStatus string

const (
	EventDeliveryPending   EventDeliveryStatus = "PENDING"
	EventDeliverySucceeded EventDeliveryStatus = "SUCCEEDED"
	EventDeliveryFailed    EventDeliveryStatus = "FAILED"
)

// EventDelivery is one event to be handled by the corresponding subscriber.
type EventDelivery struct {
	ID         string              `json:"id" db:"id"`
	EventID    string              `json:"event_id" db:"event_id"`
	Subscriber string              `json:"subscriber" db:"subscriber"`
	Status     EventDeliveryStatus `json:"status" db:"status"`
	Attempts   int                 `json:"attempts" db:"attempts"`
	NextRunAt  time.Time           `json:"next_run_at" db:"next_run_at"`
	LastError  *string             `json:"last_error,omitempty" db:"last_error"`
	CreatedAt  time.Time           `json:"created_at" db:"created_at"`
	FinishedAt *time.Time          `json:"finished_at,omitempty" db:"finished_at"`
}

// PendingDelivery is a delivery joined with its event.
type PendingDelivery struct {
	EventDelivery
	Event Event `json:"event" db:"event"`
}

// DeliveryHistory is event deliveries with its audit rows.
type DeliveryHistory struct {
	PendingDelivery
	History []AuditEvent `json:"history"`
}
