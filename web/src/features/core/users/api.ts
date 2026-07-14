// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements. See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership. The ASF licenses this file
// to you under the Apache License, Version 2.0.

import { apiFetch } from "@/shared/api/client";
import {
  grantRoleResponseSchema,
  roleDetailResponseSchema,
  rolesResponseSchema,
  userIdentitiesResponseSchema,
  userListResponseSchema,
  userPrivilegesResponseSchema,
  userRolesResponseSchema,
} from "./schemas";
import type {
  Role,
  RoleDetail,
  UserIdentity,
  UserListResponse,
  UserPrivilege,
  UserRole,
} from "./schemas";
import type { UserListParams } from "./types";

export async function listUsers(params: UserListParams = {}): Promise<UserListResponse> {
  const search = new URLSearchParams();
  if (typeof params.limit === "number") search.set("limit", String(params.limit));
  if (typeof params.offset === "number") search.set("offset", String(params.offset));
  const query = search.toString();
  return userListResponseSchema.parse(await apiFetch(`/users${query ? `?${query}` : ""}`));
}

export async function listRolesCatalog(): Promise<Role[]> {
  return rolesResponseSchema.parse(await apiFetch("/roles"));
}

export async function listRolesForUser(userId: string): Promise<UserRole[]> {
  return userRolesResponseSchema.parse(await apiFetch(`/users/${userId}/roles`));
}

export async function listUserIdentities(userId: string): Promise<UserIdentity[]> {
  return userIdentitiesResponseSchema.parse(await apiFetch(`/users/${userId}/user-identities`));
}

export async function listDirectPrivileges(userId: string): Promise<UserPrivilege[]> {
  return userPrivilegesResponseSchema.parse(await apiFetch(`/users/${userId}/privileges`));
}

export async function getRoleDetail(roleId: string): Promise<RoleDetail> {
  return roleDetailResponseSchema.parse(await apiFetch(`/roles/${roleId}`));
}

export async function assignUserRole(userId: string, roleId: string): Promise<UserRole> {
  return grantRoleResponseSchema.parse(
    await apiFetch(`/users/${userId}/roles`, {
      method: "POST",
      body: { role_id: roleId },
    }),
  );
}

export async function removeUserRole(userId: string, roleId: string): Promise<void> {
  await apiFetch(`/users/${userId}/roles/${roleId}`, { method: "DELETE" });
}
