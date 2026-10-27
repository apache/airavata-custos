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

import (
	"errors"
	"time"
)

// ErrInvalidPacket marks a packet that can never succeed as sent, so it fails without retries.
var ErrInvalidPacket = errors.New("invalid packet")

// PacketStatus represents the lifecycle state of an AMIE packet.
type PacketStatus string

const (
	PacketStatusNew       PacketStatus = "NEW"
	PacketStatusDecoded   PacketStatus = "DECODED"
	PacketStatusProcessed PacketStatus = "PROCESSED"
	PacketStatusFailed    PacketStatus = "FAILED"
	// PacketStatusWaitingApproval means the reply is held until an approval on
	// Custos side, for e.g., a cluster account approval.
	PacketStatusWaitingApproval PacketStatus = "WAITING_APPROVAL"
	// PacketStatusRefused means Custos declined the request and ACCESS was sent a
	// failure, for e.g., when a cluster account is denied.
	PacketStatusRefused PacketStatus = "REFUSED"
)

// Packet stores a raw AMIE packet and its processing state.
type Packet struct {
	ID          string       `db:"id" json:"id"`
	AmieID      int64        `db:"amie_id" json:"amie_id"`
	Type        string       `db:"type" json:"type"`
	Status      PacketStatus `db:"status" json:"status"`
	RawJSON     string       `db:"raw_json" json:"raw_json"`
	ReceivedAt  time.Time    `db:"received_at" json:"received_at"`
	DecodedAt   *time.Time   `db:"decoded_at" json:"decoded_at,omitempty"`
	ProcessedAt *time.Time   `db:"processed_at" json:"processed_at,omitempty"`
	Retries     int          `db:"retries" json:"retries"`
	LastError   *string      `db:"last_error" json:"last_error,omitempty"`
	HeldReply   *string      `db:"held_reply" json:"held_reply,omitempty"`
	// HeldFor names what HeldReply waits on, for example, a compute cluster user.
	HeldFor *string `db:"held_for" json:"held_for,omitempty"`
}
