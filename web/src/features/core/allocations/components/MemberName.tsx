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

"use client";

import { UserName } from "@/features/core/users/components/UserPicker";
import type { AllocationMembershipResponse } from "@/generated/core/types.gen";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { useAllocationMembers } from "../queries";

// A member as every allocation view names them: their name, else their email.
export const memberName = (m: AllocationMembershipResponse) =>
  m.display_name || m.email || "Unknown user";

// Without the users read privilege, a person is named from the allocation's members, which every
// allocation reader may list; anyone else is known to such a caller only as not a member.
function usePerson(id: string, allocationId: string) {
  const canReadUsers = useAbility().can("read", "User");
  const members = useAllocationMembers(canReadUsers ? undefined : allocationId);
  const member = members.data?.find((m) => m.user_id === id);
  return { canReadUsers, member, outsider: !canReadUsers && Boolean(members.data) && !member };
}

// False when this caller can never name the person, so a field naming them is left out.
export const useNameable = (id: string, allocationId: string) =>
  !usePerson(id, allocationId).outsider;

export function MemberName({ id, allocationId }: { id: string; allocationId: string }) {
  const { canReadUsers, member, outsider } = usePerson(id, allocationId);
  if (canReadUsers) return <UserName id={id} />;
  return <>{member ? memberName(member) : outsider ? "Not a member" : "—"}</>;
}
