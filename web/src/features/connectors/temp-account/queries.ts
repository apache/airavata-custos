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

import { allocationKeys, membershipLists } from "@/features/core/allocations/queries";
import { userKeys } from "@/features/core/users/queries";
import {
  deleteConnectorsTempAccountRemoveByUserId,
  getConnectorsTempAccountMembershipByUserId,
  postConnectorsTempAccountAssignAllocation,
  postConnectorsTempAccountCreate,
  postConnectorsTempAccountUpdateAllocation,
} from "@/generated/core/sdk.gen";
import type { ComputeAllocationMembership, UserInput } from "@/generated/core/types.gen";
import type { KnownError } from "@/shared/api/client";
import { isGone } from "@/shared/api/useOnGone";
import { skipToken, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

const tempAccountKeys = {
  all: ["temp-account"] as const,
  memberships: (userId: string) => [...tempAccountKeys.all, "memberships", userId] as const,
};

// Creating or removing a temporary account creates or deletes its core user.
const userLists = [
  [...userKeys.all, "list"],
  [...userKeys.all, "pages"],
];
// Temporary memberships are core allocation memberships.
const memberLists = [allocationKeys.all, ...membershipLists()];

const ASSIGN_ERRORS: KnownError[] = [
  { status: 404, copy: "This user no longer exists." },
  { status: 400, when: /must be VIRTUAL/, copy: "This user is not a temporary account." },
  { status: 400, when: /compute allocation "[^"]*" not found/, copy: "No allocation has that ID." },
  { status: 409, copy: "This account already has a membership on that allocation." },
];

function useMembershipWrite<O, R>(
  mutationFn: (options: O) => Promise<R>,
  errors: readonly KnownError[],
  userOf?: (options: O) => string,
) {
  const client = useQueryClient();
  return useMutation({
    mutationFn,
    meta: { errors },
    // An assign's 404 means the user is gone; dropping its record lets the page say so.
    onError: (error, options) =>
      userOf && isGone(error)
        ? client.resetQueries({ queryKey: [...userKeys.all, "detail", userOf(options)] })
        : undefined,
    onSettled: () =>
      Promise.all(
        [tempAccountKeys.all, ...memberLists].map((queryKey) =>
          client.invalidateQueries({ queryKey }),
        ),
      ),
  });
}

// The connector answers 404 for a user without memberships and 400 for one that is not VIRTUAL.
export function useTempMemberships(userId: string | undefined) {
  return useQuery({
    queryKey: tempAccountKeys.memberships(userId ?? ""),
    queryFn: userId
      ? () => getConnectorsTempAccountMembershipByUserId({ path: { user_id: userId } })
      : skipToken,
    retry: false,
  });
}

// The connector decodes a core user and refuses one whose type is not VIRTUAL. A new account has
// no memberships, which the connector would answer with a 404.
export function useCreateTempAccount() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: (body: UserInput) =>
      postConnectorsTempAccountCreate({ body: { ...body, type: "VIRTUAL" } }),
    meta: {
      errors: [
        { status: 409, when: /email/, copy: "A user with this email already exists." },
        { status: 400, when: /organization/, copy: "No organization has that ID." },
      ],
    },
    onSuccess: (user) =>
      client.setQueryData<ComputeAllocationMembership[]>(tempAccountKeys.memberships(user.id), []),
    onSettled: () =>
      Promise.all(userLists.map((queryKey) => client.invalidateQueries({ queryKey }))),
  });
}

export const useAssignTempAllocation = () =>
  useMembershipWrite(
    postConnectorsTempAccountAssignAllocation<true>,
    ASSIGN_ERRORS,
    ({ body }) => body.user_id ?? "",
  );

// The connector requires the user and allocation beside the id, so an update sends the whole
// membership.
export const useUpdateTempAllocation = () =>
  useMembershipWrite(
    (membership: ComputeAllocationMembership) =>
      postConnectorsTempAccountUpdateAllocation({ body: membership }),
    [{ status: 404, copy: "This membership no longer exists." }],
  );

// Drops the removed account's memberships instead of refetching them into a 404.
export function useRemoveTempAccount() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: deleteConnectorsTempAccountRemoveByUserId<true>,
    meta: { errors: [{ status: 404, copy: "This user no longer exists." }] },
    onError: (error, { path }) =>
      isGone(error)
        ? Promise.all(
            [
              tempAccountKeys.memberships(path.user_id),
              [...userKeys.all, "detail", path.user_id],
            ].map((queryKey) => client.resetQueries({ queryKey })),
          )
        : undefined,
    onSuccess: (_data, { path }) => {
      client.removeQueries({ queryKey: tempAccountKeys.memberships(path.user_id) });
      return Promise.all(
        [...userLists, ...memberLists].map((queryKey) => client.invalidateQueries({ queryKey })),
      );
    },
  });
}
