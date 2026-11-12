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

type AppStatus string

const (
	AppActive   AppStatus = "ACTIVE"
	AppDisabled AppStatus = "DISABLED"
	// AppDeleted soft delete the record.
	AppDeleted AppStatus = "DELETED"
)

// App is a tool that calls Custos as itself with its own token, never as a user.
type App struct {
	ID          string         `json:"id"          db:"id"`
	Key         string         `json:"key"         db:"key"`
	Name        string         `json:"name"        db:"name"`
	Description string         `json:"description" db:"description"`
	Status      AppStatus      `json:"status"      db:"status"`
	SecretHash  string         `json:"-"           db:"secret_hash"`
	SecretHint  string         `json:"secret_hint" db:"secret_hint"`
	CreatedBy   string         `json:"created_by"  db:"created_by"`
	CreatedAt   time.Time      `json:"created_at"  db:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"  db:"updated_at"`
	Privileges  []PrivilegeKey `json:"privileges"  db:"-"`
}
