//go:build integration

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

package service

import (
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
)

// Make sure a requested deletion leaves the allocation until approved, a denial keeps it, and an approval deletes it.
func TestDeletionWaitsForApproval(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	admin := seedUser(t, database, fmt.Sprintf("admin-%s@example.edu", uuid.NewString()))
	_, allocID := newClusterAllocation(t, svc, admin)

	pending := func() bool {
		rows, err := svc.ListDeletionRequests(ctx())
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		for _, r := range rows {
			if r.EntityID == allocID && r.Name != "" && r.ClusterName != "" {
				return true
			}
		}
		return false
	}
	exists := func() bool {
		_, err := svc.GetComputeAllocation(ctx(), allocID)
		return !errors.Is(err, ErrNotFound)
	}

	if err := svc.RequestDeletion(ctx(), DeletionOfAllocation, allocID, admin); err != nil || !pending() || !exists() {
		t.Fatalf("request: err=%v pending=%v exists=%v", err, pending(), exists())
	}
	if err := svc.DenyDeletion(ctx(), DeletionOfAllocation, allocID, admin); err != nil || pending() || !exists() {
		t.Fatalf("deny: err=%v pending=%v exists=%v", err, pending(), exists())
	}
	if err := svc.ApproveDeletion(ctx(), DeletionOfAllocation, allocID, admin); !errors.Is(err, ErrNotFound) {
		t.Fatalf("approve without a request: got %v, want ErrNotFound", err)
	}
	if err := svc.RequestDeletion(ctx(), DeletionOfAllocation, allocID, admin); err != nil {
		t.Fatalf("request again: %v", err)
	}
	if err := svc.ApproveDeletion(ctx(), DeletionOfAllocation, allocID, admin); err != nil || pending() || exists() {
		t.Fatalf("approve: err=%v pending=%v exists=%v", err, pending(), exists())
	}
}
