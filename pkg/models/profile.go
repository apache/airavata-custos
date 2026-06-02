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

// Profile is the persistent researcher identity that can outlive any single
// institutional account or external identity provider login.
type Profile struct {
	ID                 string    `json:"id" db:"id"`
	DisplayName        string    `json:"display_name" db:"display_name"`
	Emails             []string  `json:"emails,omitempty"`
	ResearchDomain     string    `json:"research_domain,omitempty" db:"research_domain"`
	Department         string    `json:"department,omitempty" db:"department"`
	Institution        string    `json:"institution,omitempty" db:"institution"`
	ProjectMemberships []string  `json:"project_memberships,omitempty"`
	GroupMemberships   []string  `json:"group_memberships,omitempty"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}
