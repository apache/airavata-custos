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
  getOrganizations,
  getOrganizationsById,
  postOrganizations,
} from "@/generated/core/sdk.gen";
import type { GetOrganizationsData } from "@/generated/core/types.gen";
import { paged } from "@/shared/api/paged";
import { skipToken, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

export const organizationKeys = {
  all: ["organizations"] as const,
  list: (query: GetOrganizationsData["query"]) => [...organizationKeys.all, "list", query] as const,
  detail: (id: string) => [...organizationKeys.all, "detail", id] as const,
};

export function useOrganizations(query: GetOrganizationsData["query"], enabled = true) {
  return useQuery({
    queryKey: organizationKeys.list(query),
    queryFn: () => getOrganizations({ query }),
    enabled,
    ...paged,
  });
}

export function useOrganization(id: string | undefined) {
  return useQuery({
    queryKey: organizationKeys.detail(id ?? ""),
    queryFn: id ? () => getOrganizationsById({ path: { id } }) : skipToken,
  });
}

export function useCreateOrganization() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: postOrganizations<true>,
    meta: {
      errors: [
        {
          status: 409,
          when: /originated_id/,
          copy: "An organization with this originated ID already exists.",
        },
      ],
    },
    onSettled: () => client.invalidateQueries({ queryKey: organizationKeys.all }),
  });
}
