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

import { useQuery } from "@tanstack/react-query";
import { getAccessStatus, getAllocationAccessMembers } from "./api";

export const accessStatusKeys = {
  all: ["access-status"] as const,
  status: (allocationId: string) => [...accessStatusKeys.all, allocationId] as const,
  members: (allocationId: string) => [...accessStatusKeys.all, allocationId, "members"] as const,
};

export function useAccessStatus(allocationId: string | undefined) {
  return useQuery({
    queryKey: allocationId
      ? accessStatusKeys.status(allocationId)
      : [...accessStatusKeys.all, "none"],
    queryFn: () => getAccessStatus(allocationId as string),
    enabled: Boolean(allocationId),
    refetchInterval: 30_000,
    refetchOnWindowFocus: false,
    staleTime: 15_000,
    gcTime: 300_000,
  });
}

export function useAllocationAccessMembers(allocationId: string | undefined, enabled: boolean) {
  return useQuery({
    queryKey: allocationId
      ? accessStatusKeys.members(allocationId)
      : [...accessStatusKeys.all, "members", "none"],
    queryFn: () => getAllocationAccessMembers(allocationId as string),
    enabled: Boolean(allocationId) && enabled,
    refetchInterval: 30_000,
    refetchOnWindowFocus: false,
    staleTime: 15_000,
    gcTime: 300_000,
  });
}
