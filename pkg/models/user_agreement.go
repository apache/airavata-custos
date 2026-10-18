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

type UserAgreementStatus string

const (
	UserAgreementDraft     UserAgreementStatus = "DRAFT"
	UserAgreementPublished UserAgreementStatus = "PUBLISHED"
	UserAgreementRemoved   UserAgreementStatus = "REMOVED"
)

// UserAgreement is one version of the agreement users must accept at sign-in.
// The current version is the PUBLISHED one with the newest PublishedAt.
type UserAgreement struct {
	ID           string              `json:"id"                     db:"id"`
	Title        string              `json:"title"                  db:"title"`
	VersionLabel string              `json:"version_label"          db:"version_label"`
	Summary      string              `json:"summary"                db:"summary"`
	BodyMD       string              `json:"body_md"                db:"body_md"`
	BodyHash     string              `json:"body_hash"              db:"body_hash"`
	Status       UserAgreementStatus `json:"status"                 db:"status"`
	ChangeNote   string              `json:"change_note"            db:"change_note"`
	CreatedBy    string              `json:"created_by"             db:"created_by"`
	CreatedAt    time.Time           `json:"created_at"             db:"created_at"`
	PublishedBy  *string             `json:"published_by,omitempty" db:"published_by"`
	PublishedAt  *time.Time          `json:"published_at,omitempty" db:"published_at"`
	UpdatedAt    time.Time           `json:"updated_at"             db:"updated_at"`
	RemovedBy    *string             `json:"removed_by,omitempty"   db:"removed_by"`
	RemovedAt    *time.Time          `json:"removed_at,omitempty"   db:"removed_at"`
}

// UserAgreementAcceptance records that a user accepted one agreement version.
// BodyHash is the hash of the text the user saw.
type UserAgreementAcceptance struct {
	ID          string    `json:"id"           db:"id"`
	UserID      string    `json:"user_id"      db:"user_id"`
	AgreementID string    `json:"agreement_id" db:"agreement_id"`
	AcceptedAt  time.Time `json:"accepted_at"  db:"accepted_at"`
	BodyHash    string    `json:"body_hash"    db:"body_hash"`
}
