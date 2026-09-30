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
import { useAbility } from "@/shared/casl/AbilityProvider";
import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { type UserActivityParams, getUserActivity, getUserActivityAnalytics } from "./api";

function useCanReadActivity() {
  const { status } = useSession();
  const ability = useAbility();
  return status === "authenticated" && ability.can("read", "UserActivity");
}
/** The filter identity of a request, ignoring which page of it was asked for. */
function sameSelection(a: UserActivityParams, b: UserActivityParams) {
  return (
    a.window === b.window &&
    a.status === b.status &&
    a.query === b.query &&
    a.sort === b.sort &&
    a.direction === b.direction
  );
}

export function useUserActivity(params: UserActivityParams) {
  const enabled = useCanReadActivity();
  return useQuery({
    queryKey: ["user-activity", "list", params],
    queryFn: () => getUserActivity(params),
    enabled,
    staleTime: 60_000,
    refetchInterval: 60_000,
    // Paging keeps the same columns, so holding the previous page avoids
    // collapsing the table to a skeleton and jolting the charts above it.
    // A new selection must not reuse it: rows from another window or status
    // would sit under the new label until the request lands.
    placeholderData: (previous, previousQuery) => {
      const previousParams = previousQuery?.queryKey?.[2] as UserActivityParams | undefined;
      if (!previous || !previousParams) return undefined;
      return sameSelection(previousParams, params) ? previous : undefined;
    },
  });
}
export function useUserActivityAnalytics(windowDays: number, userID?: string, open = true) {
  const enabled = useCanReadActivity();
  return useQuery({
    queryKey: ["user-activity", "analytics", userID ?? "all", windowDays],
    queryFn: () => getUserActivityAnalytics(windowDays, userID),
    enabled: enabled && open,
    staleTime: 60_000,
    refetchInterval: 60_000,
  });
}
