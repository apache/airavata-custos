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

// ExternalIdentity links a User to its identifier in an external system
// (ACCESS, NAIRR, CILogon, etc.). One user may have many external identities.
// Source-specific attributes (e.g. NSF status code, ACCESS org code) belong
// in Metadata as a JSON-encoded blob.
type ExternalIdentity struct {
	ID         string    `json:"id"                  db:"id"`
	UserID     string    `json:"user_id"             db:"user_id"`
	Source     string    `json:"source"              db:"source"`      // e.g. access, nairr, cilogon
	ExternalID string    `json:"external_id"         db:"external_id"` // the source's native identifier
	OIDCSub    string    `json:"oidc_sub,omitempty"  db:"oidc_sub"`    // OIDC subject when the source issues one
	Metadata   string    `json:"metadata,omitempty"  db:"metadata"`    // JSON-encoded source-specific fields
	CreatedAt  time.Time `json:"created_at"          db:"created_at"`
}

// UserDN binds an X.509 distinguished name (e.g. mTLS client cert subject) to
// a User. Append-only: DNs are credentials and are added or removed, never
// edited.
type UserDN struct {
	ID        string    `json:"id"         db:"id"`
	UserID    string    `json:"user_id"    db:"user_id"`
	DN        string    `json:"dn"         db:"dn"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
