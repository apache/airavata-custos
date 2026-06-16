import { type MongoAbility, createMongoAbility } from "@casl/ability";
import type { Privilege } from "@/features/core/identity/types";

export type AppAbility = MongoAbility;
export type Ability = AppAbility;

export type PrivilegeRule = { action: string; subject: string };

// Hand-maintained map from every PrivilegeKey the backend can return to one or
// more CASL (action, subject) pairs. Source: api/admin-privileges.openapi.yaml.
export const PRIVILEGE_ABILITY_MAP: Record<Privilege, PrivilegeRule[]> = {
  "amie:read": [{ action: "read", subject: "AMIE" }],
  "amie:write": [{ action: "manage", subject: "AMIE" }],
  "hpc:read": [
    { action: "read", subject: "Allocation" },
    { action: "read", subject: "Project" },
    { action: "read", subject: "Trace" },
    { action: "read", subject: "AuditEvent" },
  ],
  "hpc:write": [
    { action: "manage", subject: "Allocation" },
    { action: "manage", subject: "Project" },
  ],
  "signer:read": [{ action: "read", subject: "Signer" }],
  "signer:write": [{ action: "manage", subject: "Signer" }],
  "privileges:grant": [{ action: "manage", subject: "PrivilegeGrant" }],
  "roles:manage": [{ action: "manage", subject: "Role" }],
};

export function defineAbilitiesFor(privileges: Privilege[]): AppAbility {
  const rules = privileges.flatMap((p) => PRIVILEGE_ABILITY_MAP[p] ?? []);
  return createMongoAbility(rules);
}
