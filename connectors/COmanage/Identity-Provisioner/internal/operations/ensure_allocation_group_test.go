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
	"testing"
	"time"

	"github.com/apache/airavata-custos/connectors/COmanage/Identity-Provisioner/internal/client"
	"github.com/apache/airavata-custos/internal/store"
	"github.com/apache/airavata-custos/pkg/models"
)

func TestEnsureAllocationGroup_RemovesExtraAndDuplicateMemberRows(t *testing.T) {
	name := "proj-grant"
	alloc := &models.ComputeAllocation{ID: "alloc-1", Status: models.ACTIVE, PosixGroup: &name}
	row := func(id, person int) client.CoGroupMemberListOne {
		return client.CoGroupMemberListOne{Id: id, CoGroupId: 55, Person: client.IdentifierParent{Type: "CO", Id: person}, Member: true}
	}
	reg := mockRegistry{members: []client.CoGroupMemberListOne{row(9, 42), row(10, 42), row(11, 7)}}
	srv := mockComanageServer(t, &reg)
	defer srv.Close()
	core := adminCore()
	provisioned := time.Now()
	core.members = []store.MembershipWithUser{
		{ComputeAllocationMembership: models.ComputeAllocationMembership{UserID: "user-1", MembershipStatus: models.ACTIVE}, ProvisionedAt: &provisioned},
	}
	if err := newOrchestratorForTest(t, srv, core, "").EnsureAllocationGroup(context.Background(), alloc); err != nil {
		t.Fatalf("sync: %v", err)
	}
	if !slices.Equal(reg.deleted, []string{"/co_group_members/10.json", "/co_group_members/11.json"}) || len(reg.posted) != 0 {
		t.Fatalf("deleted %v, posted %v; want rows 10 and 11 deleted, nothing posted", reg.deleted, reg.posted)
	}
	var removed []string
	for _, e := range core.auditEvents {
		if e.EventType == "ComanageAllocationGroupMemberRemoved" && e.EntityID == "alloc-1" {
			removed = append(removed, e.Details)
		}
	}
	if !slices.Equal(removed, []string{"group=proj-grant co_person_id=42", "group=proj-grant co_person_id=7"}) {
		t.Fatalf("removal audits %v, want one per deleted row", removed)
	}
}
