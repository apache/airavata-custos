import { z } from "zod";
import {
  zPrivilegeKey,
  zRole,
  zUserPrivilege,
  zUserRole,
} from "@/generated/core/zod.gen";

export const privilegeKeySchema = zPrivilegeKey;
export const userPrivilegeSchema = zUserPrivilege;
export const userRoleSchema = zUserRole;
export const roleSchema = zRole;

export const roleWithPrivilegesSchema = z.object({
  role: zRole,
  privileges: z.array(zPrivilegeKey),
});

export const callerPrivilegesSchema = z.object({
  privileges: z.array(zPrivilegeKey).optional(),
});

export const privilegesResponseSchema = callerPrivilegesSchema.transform(
  (value) => value.privileges ?? [],
);

export type PrivilegeKey = z.infer<typeof privilegeKeySchema>;
