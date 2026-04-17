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

// PacketStatus represents the lifecycle state of an AMIE packet.
type PacketStatus string

const (
	PacketStatusNew       PacketStatus = "NEW"
	PacketStatusDecoded   PacketStatus = "DECODED"
	PacketStatusProcessed PacketStatus = "PROCESSED"
	PacketStatusFailed    PacketStatus = "FAILED"
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
}
