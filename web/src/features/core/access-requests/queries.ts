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

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import * as React from "react";
import {
  type AccessRequestListFilter,
  createAccessRequest,
  decideAccessRequest,
  type DecideAccessRequestBody,
  getMyAccessRequest,
  listAccessRequests,
  resolveAccessEvent,
} from "./api";

export const accessRequestKeys = {
  all: ["access-requests"] as const,
  mine: () => [...accessRequestKeys.all, "me"] as const,
  event: (code: string) => [...accessRequestKeys.all, "event", code] as const,
  list: (filter: AccessRequestListFilter) => [...accessRequestKeys.all, "list", filter] as const,
};

export function useMyAccessRequest() {
  const { status } = useSession();
  return useQuery({
    queryKey: accessRequestKeys.mine(),
    queryFn: getMyAccessRequest,
    enabled: status === "authenticated",
    retry: false,
    refetchOnWindowFocus: false,
  });
}

export function useAccessEvent(code: string) {
  return useQuery({
    queryKey: accessRequestKeys.event(code),
    queryFn: () => resolveAccessEvent(code),
    enabled: code.length > 0,
    retry: false,
    staleTime: 60_000,
    refetchOnWindowFocus: false,
  });
}

export function useCreateAccessRequest() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createAccessRequest,
    onSuccess: (created) => {
      queryClient.setQueryData(accessRequestKeys.mine(), created);
    },
  });
}

export function useAccessRequests(
  filter: AccessRequestListFilter = {},
  options: { refetchInterval?: number } = {},
) {
  return useQuery({
    queryKey: accessRequestKeys.list(filter),
    queryFn: () => listAccessRequests(filter),
    refetchOnWindowFocus: false,
    ...options,
  });
}

// Trial users are the ones an approved access request created; the map keys
// their user id to the trial end date. Admin-gated pages only.
export function useTrialExpiryByUserId(): Map<string, string> {
  const { data } = useAccessRequests({ status: "APPROVED" });
  return React.useMemo(() => {
    const map = new Map<string, string>();
    for (const r of data ?? []) {
      if (r.created_user_id && r.expires_at) map.set(r.created_user_id, r.expires_at);
    }
    return map;
  }, [data]);
}

export function useDecideAccessRequest() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, body }: { id: string; body: DecideAccessRequestBody }) =>
      decideAccessRequest(id, body),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: accessRequestKeys.all });
    },
  });
}
