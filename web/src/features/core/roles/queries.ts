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
import {
  createRole,
  listPrivilegeCatalog,
  listRoleRows,
  reconcileRolePrivileges,
  updateRole,
} from "./api";
import type { RoleInput, RoleRow } from "./schemas";

export const roleKeys = {
  all: ["roles"] as const,
  rows: () => [...roleKeys.all, "rows"] as const,
  privileges: () => [...roleKeys.all, "privileges"] as const,
};

const DEFAULTS = {
  staleTime: 30_000,
  gcTime: 300_000,
  refetchOnWindowFocus: false,
} as const;

export function useRoleRows() {
  return useQuery({
    queryKey: roleKeys.rows(),
    queryFn: listRoleRows,
    ...DEFAULTS,
  });
}

export function usePrivilegeCatalog() {
  return useQuery({
    queryKey: roleKeys.privileges(),
    queryFn: listPrivilegeCatalog,
    ...DEFAULTS,
  });
}

export function useCreateRole() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: async (input: RoleInput) => {
      const role = await createRole(input);
      if (role.id) await reconcileRolePrivileges(role.id, [], input.privileges);
      return role;
    },
    onSuccess: () => client.invalidateQueries({ queryKey: roleKeys.all }),
  });
}

export function useUpdateRole() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: async ({ role, input }: { role: RoleRow; input: RoleInput }) => {
      if (!role.id) throw new Error("Role id is required");
      const updated = await updateRole(role.id, input);
      await reconcileRolePrivileges(role.id, role.privileges, input.privileges);
      return updated;
    },
    onSuccess: () => client.invalidateQueries({ queryKey: roleKeys.all }),
  });
}
