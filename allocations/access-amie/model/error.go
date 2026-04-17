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

// ProcessingError records a failure that occurred during packet or event
// processing. At least one of PacketID or EventID will be non-nil.
type ProcessingError struct {
	ID         int64     `db:"id" json:"id"`
	PacketID   *string   `db:"packet_id" json:"packet_id,omitempty"`
	EventID    *string   `db:"event_id" json:"event_id,omitempty"`
	OccurredAt time.Time `db:"occurred_at" json:"occurred_at"`
	Summary    string    `db:"summary" json:"summary"`
	Detail     *string   `db:"detail" json:"detail,omitempty"`
}
