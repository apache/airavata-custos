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

type Person struct {
	ID             string    `db:"id" json:"id"`
	AccessGlobalID string    `db:"access_global_id" json:"access_global_id"`
	FirstName      string    `db:"first_name" json:"first_name"`
	LastName       string    `db:"last_name" json:"last_name"`
	Email          string    `db:"email" json:"email"`
	Organization   *string   `db:"organization" json:"organization,omitempty"`
	OrgCode        *string   `db:"org_code" json:"org_code,omitempty"`
	NsfStatusCode  *string   `db:"nsf_status_code" json:"nsf_status_code,omitempty"`
	IsActive       bool      `db:"is_active" json:"is_active"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}

type PersonGlobalID struct {
	ID       int64  `db:"id" json:"id"`
	PersonID string `db:"person_id" json:"person_id"`
	GlobalID string `db:"global_id" json:"global_id"`
}

type PersonDN struct {
	ID       int64  `db:"id" json:"id"`
	PersonID string `db:"person_id" json:"person_id"`
	DN       string `db:"dn" json:"dn"`
}
