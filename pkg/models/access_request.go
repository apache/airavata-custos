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

// AccessRequestStatus enumerates the lifecycle states of an AccessRequest.
type AccessRequestStatus string

const (
	AccessRequestPending  AccessRequestStatus = "PENDING"
	AccessRequestApproved AccessRequestStatus = "APPROVED"
	AccessRequestDenied   AccessRequestStatus = "DENIED"
)

const (
	AccessRequestEventCreated  = "CREATED"
	AccessRequestEventApproved = "APPROVED"
	AccessRequestEventDenied   = "DENIED"
	// AccessRequestEventFailed records an approval attempt whose provisioning
	// failed; the request itself stays PENDING so it can be re-approved.
	AccessRequestEventFailed = "FAILED"
)

// AccessEvent maps an event code to the allocation and organization an
// approved requester joins.
type AccessEvent struct {
	Code                string `json:"code"                  db:"code"`
	ComputeAllocationID string `json:"compute_allocation_id" db:"compute_allocation_id"`
	OrganizationID      string `json:"organization_id"       db:"organization_id"`
}

// AccessRequest is a self-service request for a trial account, identified by
// the verified token subject rather than an existing user.
type AccessRequest struct {
	ID            string              `json:"id"                        db:"id"`
	OIDCSub       string              `json:"oidc_sub"                  db:"oidc_sub"`
	Email         string              `json:"email"                     db:"email"`
	Name          string              `json:"name"                      db:"name"`
	Institution   string              `json:"institution"               db:"institution"`
	EventCode     string              `json:"event_code"                db:"event_code"`
	Reason        string              `json:"reason,omitempty"          db:"reason"`
	Status        AccessRequestStatus `json:"status"                    db:"status"`
	ApproverID    string              `json:"approver_id,omitempty"     db:"approver_id"`
	DenyReason    string              `json:"deny_reason,omitempty"     db:"deny_reason"`
	ExpiresAt     *time.Time          `json:"expires_at,omitempty"      db:"expires_at"`
	CreatedUserID string              `json:"created_user_id,omitempty" db:"created_user_id"`
	Timestamp     time.Time           `json:"timestamp"                 db:"timestamp"`
}

// AccessRequestEvent is an append-only audit trail entry for state
// transitions applied to an AccessRequest.
type AccessRequestEvent struct {
	ID              string    `json:"id"                    db:"id"`
	AccessRequestID string    `json:"access_request_id"     db:"access_request_id"`
	EventType       string    `json:"event_type"            db:"event_type"` // "CREATED", "APPROVED", "DENIED"
	Description     string    `json:"description,omitempty" db:"description"`
	Timestamp       time.Time `json:"timestamp"             db:"timestamp"`
}
