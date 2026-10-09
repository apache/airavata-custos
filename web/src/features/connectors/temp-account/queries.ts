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

import {
  deleteConnectorsTempAccountRemoveByUserId,
  getConnectorsTempAccountMembershipByUserId,
  postConnectorsTempAccountAssignAllocation,
  postConnectorsTempAccountCreate,
  postConnectorsTempAccountUpdateAllocation,
} from "@/generated/temp-account/sdk.gen";
import { allocationKeys, membershipLists } from "@/features/core/allocations/queries";
import { userKeys } from "@/features/core/users/queries";
import { useInvalidating } from "@/shared/api/useInvalidating";
import { skipToken, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

const tempAccountKeys = {
  all: ["temp-account"] as const,
  memberships: (userId: string) => [...tempAccountKeys.all, "memberships", userId] as const,
};

// Creating or removing a temporary account creates or deletes its core user.
const userLists = [[...userKeys.all, "list"], [...userKeys.all, "pages"]];
// Temporary memberships are core allocation memberships.
const memberLists = [allocationKeys.all, ...membershipLists()];

export function useTempMemberships(userId: string | undefined) {
  return useQuery({
    queryKey: tempAccountKeys.memberships(userId ?? ""),
    queryFn: userId
      ? () => getConnectorsTempAccountMembershipByUserId({ path: { user_id: userId } })
      : skipToken,
    retry: false,
  });
}

export function useCreateTempAccount() {
  return useInvalidating(postConnectorsTempAccountCreate<true>, tempAccountKeys.all, ...userLists);
}

export function useAssignTempAllocation() {
  return useInvalidating(
    postConnectorsTempAccountAssignAllocation<true>,
    tempAccountKeys.all,
    ...memberLists,
  );
}

export function useUpdateTempAllocation() {
  return useInvalidating(
    postConnectorsTempAccountUpdateAllocation<true>,
    tempAccountKeys.all,
    ...memberLists,
  );
}

// Drops the removed account's memberships instead of refetching them into a 404.
export function useRemoveTempAccount() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: deleteConnectorsTempAccountRemoveByUserId<true>,
    onSuccess: (_data, { path }) => {
      client.removeQueries({ queryKey: tempAccountKeys.memberships(path.user_id) });
      return Promise.all(
        [...userLists, ...memberLists].map((queryKey) => client.invalidateQueries({ queryKey })),
      );
    },
  });
}
