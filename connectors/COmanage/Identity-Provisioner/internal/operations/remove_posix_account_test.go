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

package operations

import (
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/apache/airavata-custos/pkg/models"
)

// Make sure removing an account takes it off the person and deletes its primary group, binding first.
func TestRemovePOSIXAccount_RemovesTheAccountAndItsPrimaryGroup(t *testing.T) {
	reg := &mockRegistry{account: "jdoe", userGroup: 55}
	srv := mockComanageServer(t, reg)
	defer srv.Close()
	core := adminCore()

	cu := &models.ComputeClusterUser{ID: "ccu-1", UserID: "user-1", LocalUsername: "jdoe"}
	if err := newOrchestratorForTest(t, srv, core, adminGroupName).RemovePOSIXAccount(context.Background(), cu); err != nil {
		t.Fatalf("RemovePOSIXAccount: %v", err)
	}

	if len(reg.put) != 1 || strings.Contains(string(reg.put[0]), "jdoe") {
		t.Errorf("expected one person update without jdoe, got %d: %s", len(reg.put), reg.put)
	}
	want := []string{"/unix_cluster/unix_cluster_groups/3.json", "/co_groups/55.json"}
	if !slices.Equal(reg.deleted, want) {
		t.Errorf("deleted = %v, want %v", reg.deleted, want)
	}
	if !hasAudit(core, "ComanageClusterAccountRemoved") {
		t.Error("no ComanageClusterAccountRemoved audit event")
	}
}

// Make sure removing an admin's account also takes the person out of the admin group.
func TestRemovePOSIXAccount_TakesAnAdminOutOfTheAdminGroup(t *testing.T) {
	reg := &mockRegistry{account: "jdoe", adminGroup: adminGroupID, adminMember: true}
	srv := mockComanageServer(t, reg)
	defer srv.Close()

	cu := &models.ComputeClusterUser{ID: "ccu-1", UserID: "user-1", LocalUsername: "jdoe", AccessLevel: models.ClusterAccessAdmin}
	if err := newOrchestratorForTest(t, srv, adminCore(), adminGroupName).RemovePOSIXAccount(context.Background(), cu); err != nil {
		t.Fatalf("RemovePOSIXAccount: %v", err)
	}

	if !slices.Contains(reg.deleted, "/co_group_members/9.json") {
		t.Errorf("admin membership not deleted, deleted = %v", reg.deleted)
	}
}

// Make sure a second run, with nothing left in the registry, writes nothing and does not fail.
func TestRemovePOSIXAccount_SecondRunChangesNothing(t *testing.T) {
	reg := &mockRegistry{}
	srv := mockComanageServer(t, reg)
	defer srv.Close()

	cu := &models.ComputeClusterUser{ID: "ccu-1", UserID: "user-1", LocalUsername: "jdoe"}
	if err := newOrchestratorForTest(t, srv, adminCore(), adminGroupName).RemovePOSIXAccount(context.Background(), cu); err != nil {
		t.Fatalf("RemovePOSIXAccount: %v", err)
	}

	if len(reg.put) != 0 || len(reg.deleted) != 0 {
		t.Errorf("expected no writes, got %d updates and deletes %v", len(reg.put), reg.deleted)
	}
}
