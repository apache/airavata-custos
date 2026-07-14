// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements. See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership. The ASF licenses this file
// to you under the Apache License, Version 2.0.

import { z } from "zod";
import { zRole, zUser, zUserIdentity, zUserPrivilege, zUserRole } from "@/generated/core/zod.gen";

export const privilegeKeySchema = z.string().min(1);

export const userSchema = zUser.required({ id: true, email: true });
export const roleSchema = zRole.required({ id: true, name: true });
export const userIdentitySchema = zUserIdentity.required({
  id: true,
  source: true,
  user_id: true,
});
export const userRoleSchema = zUserRole
  .extend({
    granted_by: z.string().nullish(),
    reason: z.string().nullish(),
  })
  .required({ role_id: true, user_id: true });
export const userPrivilegeSchema = zUserPrivilege
  .extend({
    privilege: privilegeKeySchema,
    granted_by: z.string().nullish(),
    reason: z.string().nullish(),
  })
  .required({ user_id: true });

export const userListResponseSchema = z.object({
  items: z.array(userSchema),
  total: z.number().int().nonnegative(),
});
export const rolesResponseSchema = z
  .array(roleSchema)
  .nullish()
  .transform((value) => value ?? []);
export const userRolesResponseSchema = z
  .array(userRoleSchema)
  .nullish()
  .transform((value) => value ?? []);
export const userIdentitiesResponseSchema = z
  .array(userIdentitySchema)
  .nullish()
  .transform((value) => value ?? []);
export const userPrivilegesResponseSchema = z
  .array(userPrivilegeSchema)
  .nullish()
  .transform((value) => value ?? []);
export const roleDetailResponseSchema = z.object({
  role: roleSchema.optional(),
  privileges: z
    .array(privilegeKeySchema)
    .nullish()
    .transform((value) => value ?? []),
});

export const grantRoleResponseSchema = userRoleSchema;

export type User = z.infer<typeof userSchema>;
export type Role = z.infer<typeof roleSchema>;
export type UserIdentity = z.infer<typeof userIdentitySchema>;
export type UserRole = z.infer<typeof userRoleSchema>;
export type UserPrivilege = z.infer<typeof userPrivilegeSchema>;
export type UserListResponse = z.infer<typeof userListResponseSchema>;
export type RoleDetail = z.infer<typeof roleDetailResponseSchema>;
