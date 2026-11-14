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

import (
	"strconv"
	"time"
)

// StorageType says who a storage is for: an account, an allocation group, or a member of one.
type StorageType string

const (
	StorageUser          StorageType = "user"
	StorageProject       StorageType = "project"
	StorageProjectUser   StorageType = "project_user"
	StorageProjectShared StorageType = "project_shared"
)

var storageTypes = map[StorageType]struct{ perUser, perAllocation bool }{
	StorageUser:          {perUser: true},
	StorageProject:       {perAllocation: true},
	StorageProjectUser:   {perUser: true, perAllocation: true},
	StorageProjectShared: {perAllocation: true},
}

// Takes reports whether a storage of the type is for a user, an allocation, or both; ok is false for an unknown type.
func (t StorageType) Takes() (user, allocation, ok bool) {
	fit, ok := storageTypes[t]
	return fit.perUser, fit.perAllocation, ok
}

// Storage is one owner's storage on a cluster, as a storage connector makes it. Username and the allocation fields say
// who it is for; the allocation fields are a copy, so they outlive the allocation. A zero quota means none of its own.
type Storage struct {
	ID               string      `json:"id"                 db:"id"`
	ComputeClusterID string      `json:"cluster_id"         db:"compute_cluster_id"`
	PolicyID         *string     `json:"policy_id"          db:"policy_id"` // The storage policy it was last made from; nil once that is deleted.
	Type             StorageType `json:"type"               db:"type"`
	Path             string      `json:"path"               db:"path"`
	Username         string      `json:"username"           db:"username"`
	AllocationID     *string     `json:"allocation_id"      db:"allocation_id"`
	AllocationName   *string     `json:"allocation_name"    db:"allocation_name"`
	StartTime        *time.Time  `json:"start_time"         db:"start_time"`
	EndTime          *time.Time  `json:"end_time"           db:"end_time"`
	Owner            string      `json:"owner"              db:"owner"`
	Group            string      `json:"group"              db:"group"`
	Mode             string      `json:"mode"               db:"mode"`
	QuotaBytes       int64       `json:"quota_bytes"        db:"quota_bytes"`
	QuotaInodes      int64       `json:"quota_inodes"       db:"quota_inodes"`
	QuotaUntil       *time.Time  `json:"quota_until"        db:"quota_until"` // Set while an administrator's quota stands in for the policy's.
	UsedBytes        int64       `json:"used_bytes"         db:"used_bytes"`
	Status           string      `json:"status"             db:"status"`
	Inherited        bool        `json:"inherited"          db:"inherited"` // It counts against the quota of a storage above it.
}

// StoragePolicy says how a cluster's storages of one type, such as home or scratch, are made. PathPattern, Owner and
// Group are templates over {user}, the account's username, and {allocation}, the allocation's group. A zero quota means none.
type StoragePolicy struct {
	ID               string      `json:"id"                 db:"id"`
	ComputeClusterID string      `json:"cluster_id"         db:"compute_cluster_id"`
	Type             StorageType `json:"type"               db:"type"`
	PathPattern      string      `json:"path_pattern"       db:"path_pattern"`
	Owner            string      `json:"owner"              db:"owner"`
	Group            string      `json:"group"              db:"group"`
	Mode             string      `json:"mode"               db:"mode"` // Octal permission bits, such as "0755"; see ModeBits.
	QuotaBytes       int64       `json:"quota_bytes"        db:"quota_bytes"`
	QuotaInodes      int64       `json:"quota_inodes"       db:"quota_inodes"`
}

// ModeBits parses a storage's octal mode.
func ModeBits(mode string) (int, error) {
	bits, err := strconv.ParseUint(mode, 8, 12)
	return int(bits), err
}
