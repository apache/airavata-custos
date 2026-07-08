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
import { createOrganization, listOrganizations } from "./api";
import type { CreateOrganizationPayload } from "./schemas";
import type { OrganizationListParams } from "./types";

export const organizationKeys = {
  all: ["organizations"] as const,
  list: (params: OrganizationListParams = {}) =>
    [...organizationKeys.all, "list", params] as const,
};

const DEFAULTS = {
  staleTime: 30_000,
  gcTime: 300_000,
  refetchOnWindowFocus: false,
} as const;

export function useOrganizations(
  params: OrganizationListParams = {},
  options: { enabled?: boolean } = {},
) {
  return useQuery({
    queryKey: organizationKeys.list(params),
    queryFn: () => listOrganizations(params),
    enabled: options.enabled ?? true,
    ...DEFAULTS,
  });
}

export function useCreateOrganization() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: (payload: CreateOrganizationPayload) => createOrganization(payload),
    onSuccess: () => client.invalidateQueries({ queryKey: organizationKeys.all }),
  });
}
