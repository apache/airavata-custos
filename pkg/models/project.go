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

// UserStatus enumerates the lifecycle states a User may occupy.
type UserStatus string

const (
	UserActive    UserStatus = "ACTIVE"
	UserInactive  UserStatus = "INACTIVE"
	UserSuspended UserStatus = "SUSPENDED"
	UserMerged    UserStatus = "MERGED"
)

// ProjectStatus enumerates the lifecycle states a Project may occupy.
type ProjectStatus string

const (
	ProjectActive   ProjectStatus = "ACTIVE"
	ProjectInactive ProjectStatus = "INACTIVE"
	ProjectDeleted  ProjectStatus = "DELETED"
)

type Project struct {
	ID           string        `json:"id"            db:"id"`
	OriginatedID string        `json:"originated_id" db:"originated_id"` // The ID of the project in origination. For example: ACCESS Record ID.
	Title        string        `json:"title"         db:"title"`
	Origination  string        `json:"origination"   db:"origination"` // ACCESS, NAIRR, XRASS, etc.
	ProjectPIID  string        `json:"project_pi_id" db:"project_pi_id"`
	Status       ProjectStatus `json:"status"        db:"status"`
	CreatedTime  time.Time     `json:"created_time"  db:"created_time"`
}

type Organization struct {
	ID           string `json:"id"            db:"id"`
	OriginatedID string `json:"originated_id" db:"originated_id"` // The ID of the organization in origination. For example: ACCESS Record ID.
	Name         string `json:"name"          db:"name"`
}

// UserType enumerates the types of users.
type UserType string

const (
	UserTypeClusterLocal UserType = "CLUSTER_LOCAL"
	UserTypeVirtual      UserType = "VIRTUAL"
)

type User struct {
	ID             string     `json:"id"              db:"id"`
	OrganizationID string     `json:"organization_id" db:"organization_id"`
	FirstName      string     `json:"first_name"      db:"first_name"`
	LastName       string     `json:"last_name"       db:"last_name"`
	MiddleName     string     `json:"middle_name,omitempty" db:"middle_name"`
	Email          string     `json:"email"           db:"email"`
	Status         UserStatus `json:"status"          db:"status"`
	Type           UserType   `json:"type"            db:"type"`
}
