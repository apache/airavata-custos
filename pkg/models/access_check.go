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

// AccessCheckType identifies which part of a member's cluster access a
// check covers.
type AccessCheckType string

const (
	AccessCheckSignIn        AccessCheckType = "SIGN_IN"
	AccessCheckJobSubmission AccessCheckType = "JOB_SUBMISSION"
)

// AccessCheckStatus is the raw probe outcome. UI states (setting up,
// retrying, stuck) derive from this plus the failing_since timestamps.
type AccessCheckStatus string

const (
	AccessCheckPending AccessCheckStatus = "PENDING"
	AccessCheckOK      AccessCheckStatus = "OK"
	AccessCheckFailing AccessCheckStatus = "FAILING"
)

// Access check event types. Milestones only, never individual probe ticks.
const (
	AccessCheckEventStarted   = "STARTED"
	AccessCheckEventOnline    = "ONLINE"
	AccessCheckEventFailed    = "FAILED"
	AccessCheckEventRecovered = "RECOVERED"
	AccessCheckEventStuck     = "STUCK"
)

// AccessCheck is the current probe state for one (allocation, user, check)
// target.
type AccessCheck struct {
	ID                  string            `json:"id"                    db:"id"`
	ComputeAllocationID string            `json:"compute_allocation_id" db:"compute_allocation_id"`
	UserID              string            `json:"user_id"               db:"user_id"`
	CheckType           AccessCheckType   `json:"check_type"            db:"check_type"`
	Status              AccessCheckStatus `json:"status"                db:"status"`
	Detail              string            `json:"detail,omitempty"      db:"detail"`
	LastCheckedAt       time.Time         `json:"last_checked_at"       db:"last_checked_at"`
	LastOKAt            *time.Time        `json:"last_ok_at"            db:"last_ok_at"`
	FailingSince        *time.Time        `json:"failing_since"         db:"failing_since"`
}

// AccessCheckEvent is one milestone in a check's history (state changes
// only).
type AccessCheckEvent struct {
	ID            string    `json:"id"              db:"id"`
	AccessCheckID string    `json:"access_check_id" db:"access_check_id"`
	EventType     string    `json:"event_type"      db:"event_type"`
	OccurredAt    time.Time `json:"occurred_at"     db:"occurred_at"`
}
