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

// StorageType names a target in the config, such as "user_home". Its path's placeholders say whose storage it is.
type StorageType string

// StorageTarget says how a cluster's storage of one type is made; see docs/storage.md. A zero HardLimitSize means no
// quota of its own, so the storage counts against the one above it.
type StorageTarget struct {
	Path            string `yaml:"path"`
	Owner           string `yaml:"owner"`
	Group           string `yaml:"group"`
	Mode            string `yaml:"mode"` // Octal permission bits, such as "0700".
	HardLimitSize   int64  `yaml:"hard_limit_size"`
	HardLimitInodes int64  `yaml:"hard_limit_inodes"`
}

// Storage is one owner's storage on a cluster, as a storage connector makes it. ComputeClusterUserID, AllocationID and
// UserID say who it is for; a deactivated storage keeps none once its holder is deleted. A zero quota means none of its own.
type Storage struct {
	ID                   string      `json:"id"                      db:"id"`
	ComputeClusterID     string      `json:"compute_cluster_id"      db:"compute_cluster_id"`
	ComputeClusterUserID *string     `json:"compute_cluster_user_id" db:"compute_cluster_user_id"`
	AllocationID         *string     `json:"allocation_id"           db:"allocation_id"`
	UserID               *string     `json:"user_id"                 db:"user_id"`
	Type                 StorageType `json:"type"                    db:"type"`
	Path                 string      `json:"path"                    db:"path"`
	Owner                string      `json:"owner"                   db:"owner"`
	Group                string      `json:"group"                   db:"group"`
	Mode                 string      `json:"mode"                    db:"mode"`
	QuotaBytes           int64       `json:"quota_bytes"             db:"quota_bytes"`
	QuotaInodes          int64       `json:"quota_inodes"            db:"quota_inodes"`
	QuotaUntil           *time.Time  `json:"quota_until"             db:"quota_until"` // Set while an administrator's quota stands in for the target's.
	Status               string      `json:"status"                  db:"status"`
	// Read from the holder when listed.
	Username       *string    `json:"username,omitempty"        db:"username"`
	AllocationName *string    `json:"allocation_name,omitempty" db:"allocation_name"`
	StartTime      *time.Time `json:"start_time,omitempty"      db:"start_time"`
	EndTime        *time.Time `json:"end_time,omitempty"        db:"end_time"`
	UsedBytes      *int64     `json:"used_bytes,omitempty"      db:"-"` // Read from the storage connector.
}
