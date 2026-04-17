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

package model

import "time"

// ProcessingStatus represents the lifecycle state of a processing event.
type ProcessingStatus string

const (
	ProcessingStatusNew               ProcessingStatus = "NEW"
	ProcessingStatusRunning           ProcessingStatus = "RUNNING"
	ProcessingStatusSucceeded         ProcessingStatus = "SUCCEEDED"
	ProcessingStatusFailed            ProcessingStatus = "FAILED"
	ProcessingStatusRetryScheduled    ProcessingStatus = "RETRY_SCHEDULED"
	ProcessingStatusPermanentlyFailed ProcessingStatus = "PERMANENTLY_FAILED"
)

// ProcessingEventType identifies the kind of work a processing event performs.
type ProcessingEventType string

const (
	EventTypeDecodePacket ProcessingEventType = "DECODE_PACKET"
)

// ProcessingEvent tracks a single unit of asynchronous work against a packet.
type ProcessingEvent struct {
	ID          string              `db:"id" json:"id"`
	PacketID    string              `db:"packet_id" json:"packet_id"`
	Type        ProcessingEventType `db:"type" json:"type"`
	Status      ProcessingStatus    `db:"status" json:"status"`
	Attempts    int                 `db:"attempts" json:"attempts"`
	Payload     []byte              `db:"payload" json:"-"`
	CreatedAt   time.Time           `db:"created_at" json:"created_at"`
	StartedAt   *time.Time          `db:"started_at" json:"started_at,omitempty"`
	FinishedAt  *time.Time          `db:"finished_at" json:"finished_at,omitempty"`
	LastError   *string             `db:"last_error" json:"last_error,omitempty"`
	NextRetryAt *time.Time          `db:"next_retry_at" json:"next_retry_at,omitempty"`
}

// EventWithPacket is a read-only projection that joins a processing event with
// selected columns from its associated packet. It is used by the worker to
// avoid a second query when claiming work.
type EventWithPacket struct {
	ProcessingEvent
	PacketAmieID  int64  `db:"packet_amie_id"`
	PacketType    string `db:"packet_type"`
	PacketRawJSON string `db:"packet_raw_json"`
}
