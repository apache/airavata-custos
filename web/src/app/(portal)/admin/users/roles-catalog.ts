import type { RoleRow } from "./types";

// Seed roles for the mock portal. Permissions are granted per left-nav
// section (see permissions.ts) — source once a real "list roles" call backs
// this page: GET /roles + GET /roles/{id}.
export const ROLE_SUPER_ADMIN: RoleRow = {
  id: "role-super-admin",
  name: "Super Admin",
  description: "Full administrative access across the portal.",
  is_system: true,
  permissions: [
    "allocations:read",
    "allocations:write",
    "projects:read",
    "projects:write",
    "tracing:read",
    "tracing:write",
    "amie:read",
    "amie:write",
    "users_permissions:read",
    "users_permissions:write",
  ],
};
export const ROLE_OPERATOR: RoleRow = {
  id: "role-operator",
  name: "Operator",
  description: "Day-to-day allocation, project, and AMIE operations (read + write).",
  is_system: false,
  permissions: [
    "allocations:read",
    "allocations:write",
    "projects:read",
    "projects:write",
    "amie:read",
    "amie:write",
  ],
};
export const ROLE_AUDITOR: RoleRow = {
  id: "role-auditor",
  name: "Auditor",
  description: "Read-only access across allocations, projects, tracing, and AMIE.",
  is_system: false,
  permissions: ["allocations:read", "projects:read", "tracing:read", "amie:read"],
};

export const INITIAL_ROLES: RoleRow[] = [ROLE_SUPER_ADMIN, ROLE_OPERATOR, ROLE_AUDITOR];
