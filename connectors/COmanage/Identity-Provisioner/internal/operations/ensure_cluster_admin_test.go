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
	"strings"
	"testing"

	"github.com/apache/airavata-custos/connectors/COmanage/Identity-Provisioner/internal/client"
	"github.com/apache/airavata-custos/pkg/models"
)

// adminCore is a fake core with a user whose CoPerson is already stored
func adminCore() *fakeCore {
	return &fakeCore{
		user: &models.User{ID: "user-1", Email: "admin@example.edu"},
		identities: []models.UserIdentity{
			{UserID: "user-1", Source: "comanage", ExternalID: "Person100099"},
		},
	}
}

func hasAudit(core *fakeCore, eventType string) bool {
	for _, e := range core.auditEvents {
		if e.EventType == eventType {
			return true
		}
	}
	return false
}

// Make sure all the members are added as plain members without making them owners.
func TestEnsureClusterAdminMembership_JoinsAsPlainMember(t *testing.T) {
	reg := &mockRegistry{adminGroup: adminGroupID}
	srv := mockComanageServer(t, reg)
	defer srv.Close()
	core := adminCore()

	cu := &models.ComputeClusterUser{ID: "ccu-1", UserID: "user-1", AccessLevel: models.ClusterAccessAdmin}
	if err := newOrchestratorForTest(t, srv, core, adminGroupName).EnsureClusterAdminMembership(context.Background(), cu); err != nil {
		t.Fatalf("EnsureClusterAdminMembership: %v", err)
	}

	want := client.CoGroupMemberCreateOne{
		Version:   "1.0",
		Person:    client.IdentifierParent{Type: "CO", Id: 42},
		CoGroupId: adminGroupID,
		Member:    true,
		Owner:     false,
	}
	if len(reg.posted) != 1 || reg.posted[0] != want {
		t.Fatalf("membership posted = %+v, want exactly %+v", reg.posted, want)
	}
	if !hasAudit(core, "ComanageClusterAdminGroupJoined") {
		t.Error("no ComanageClusterAdminGroupJoined audit event")
	}
}

// Make sure a second run does not add a duplicate membership.
func TestEnsureClusterAdminMembership_SecondRunAddsNothing(t *testing.T) {
	reg := &mockRegistry{adminGroup: adminGroupID, adminMember: true}
	srv := mockComanageServer(t, reg)
	defer srv.Close()

	cu := &models.ComputeClusterUser{ID: "ccu-1", UserID: "user-1", AccessLevel: models.ClusterAccessAdmin}
	if err := newOrchestratorForTest(t, srv, adminCore(), adminGroupName).EnsureClusterAdminMembership(context.Background(), cu); err != nil {
		t.Fatalf("EnsureClusterAdminMembership: %v", err)
	}
	if len(reg.posted) != 0 {
		t.Fatalf("posted %d memberships for an existing member, want 0", len(reg.posted))
	}
}

// Make sure nothing is written when the user has no stored CoPerson.
func TestEnsureClusterAdminMembership_NoStoredPersonWritesNothing(t *testing.T) {
	reg := &mockRegistry{adminGroup: adminGroupID}
	srv := mockComanageServer(t, reg)
	defer srv.Close()
	core := &fakeCore{user: &models.User{ID: "user-1", Email: "admin@example.edu"}}

	cu := &models.ComputeClusterUser{ID: "ccu-1", UserID: "user-1", AccessLevel: models.ClusterAccessAdmin}
	err := newOrchestratorForTest(t, srv, core, adminGroupName).EnsureClusterAdminMembership(context.Background(), cu)
	if err == nil || !strings.Contains(err.Error(), "no stored CoPerson") {
		t.Fatalf("err = %v, want no stored CoPerson", err)
	}
	if len(reg.posted) != 0 {
		t.Errorf("posted %d memberships, want 0", len(reg.posted))
	}
	if !hasAudit(core, "ComanageProvisioningFailed") {
		t.Error("no ComanageProvisioningFailed audit event")
	}
}

// Make sure nothing is written when the admin group is not configured or not in the registry.
func TestEnsureClusterAdminMembership_MissingGroupWritesNothing(t *testing.T) {
	tests := []struct {
		name      string
		groupName string
		registry  mockRegistry
		wantErr   string
	}{
		{"not configured", "", mockRegistry{adminGroup: adminGroupID}, "not configured"},
		{"not in registry", adminGroupName, mockRegistry{}, "not in the registry"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := tt.registry
			srv := mockComanageServer(t, &reg)
			defer srv.Close()
			core := adminCore()

			cu := &models.ComputeClusterUser{ID: "ccu-1", UserID: "user-1", AccessLevel: models.ClusterAccessAdmin}
			err := newOrchestratorForTest(t, srv, core, tt.groupName).EnsureClusterAdminMembership(context.Background(), cu)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("err = %v, want it to contain %q", err, tt.wantErr)
			}
			if len(reg.posted) != 0 {
				t.Errorf("posted %d memberships, want 0", len(reg.posted))
			}
			if !hasAudit(core, "ComanageProvisioningFailed") {
				t.Error("no ComanageProvisioningFailed audit event")
			}
		})
	}
}
