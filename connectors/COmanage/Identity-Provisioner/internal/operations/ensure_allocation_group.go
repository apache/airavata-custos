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
	"fmt"
	"strconv"

	"github.com/apache/airavata-custos/pkg/models"
)

// allocationGIDBase offsets a CoGroup id into a gid above the personal
// groups, which mirror uidnumbers.
const allocationGIDBase = 3_000_000

// EnsureAllocationGroup gives an active allocation its CoGroup and makes it
// hold exactly its active members with a provisioned cluster account.
func (o *Orchestrator) EnsureAllocationGroup(ctx context.Context, a *models.ComputeAllocation) error {
	if a.PosixGroup == nil {
		return nil
	}
	groupID, err := o.c.FindCoGroupByName(*a.PosixGroup)
	if err != nil {
		return err
	}
	if groupID == 0 {
		if a.Status != models.ACTIVE {
			return nil
		}
		if groupID, err = o.c.CreateCoGroup(*a.PosixGroup, "Custos allocation "+a.ID); err != nil {
			return err
		}
		o.auditAllocation(ctx, a, "ComanageAllocationGroupCreated", fmt.Sprintf("group=%s co_group_id=%d", *a.PosixGroup, groupID))
	}

	binding, err := o.c.FindUnixClusterGroup(groupID)
	if err == nil && binding == 0 {
		_, err = o.c.CreateUnixClusterGroup(groupID)
	}
	if err != nil {
		return err
	}
	for typ, value := range map[string]string{"uid": *a.PosixGroup, "gidnumber": strconv.Itoa(allocationGIDBase + groupID)} {
		existing, err := o.c.FindIdentifierOnGroup(groupID, typ)
		if err == nil && existing == 0 {
			_, err = o.c.CreateIdentifierOnGroup(value, typ, groupID)
		}
		if err != nil {
			return err
		}
	}

	members, err := o.core.ListMembersForAllocation(ctx, a.ID)
	if err != nil {
		return err
	}
	want := map[int]bool{}
	for _, m := range members {
		if m.MembershipStatus != models.ACTIVE || m.ProvisionedAt == nil {
			continue
		}
		_, composite, err := o.findCoPerson(ctx, &models.User{ID: m.UserID})
		if err != nil {
			return err
		}
		if id, _ := extractCoPersonID(composite); id != 0 {
			want[id] = true
		}
	}
	rows, err := o.c.ListCoGroupMembers(groupID)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if row.Member && want[row.Person.Id] {
			delete(want, row.Person.Id)
			continue
		}
		if err := ignoreNotFound(o.c.DeleteCoGroupMember(row.Id)); err != nil {
			return err
		}
		o.auditAllocation(ctx, a, "ComanageAllocationGroupMemberRemoved", fmt.Sprintf("group=%s co_person_id=%d", *a.PosixGroup, row.Person.Id))
	}
	for id := range want {
		if _, err := o.c.CreateCoGroupMember(id, groupID); err != nil {
			return err
		}
		o.auditAllocation(ctx, a, "ComanageAllocationGroupMemberAdded", fmt.Sprintf("group=%s co_person_id=%d", *a.PosixGroup, id))
	}
	return nil
}

func (o *Orchestrator) auditAllocation(ctx context.Context, a *models.ComputeAllocation, eventType, details string) {
	_, _ = o.core.CreateAuditEvent(ctx, &models.AuditEvent{EventType: eventType, EntityID: a.ID, EntityType: "compute_allocation", Details: details})
}
