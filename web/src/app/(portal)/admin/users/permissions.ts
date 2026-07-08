// Permission sections mirror the portal's left-nav items (see
// src/shared/layout/nav.ts) one-to-one, so each section can be granted
// read/write independently.
export const PERMISSION_SECTIONS = [
  "allocations",
  "projects",
  "tracing",
  "amie",
  "users_permissions",
] as const;
export type PermissionSection = (typeof PERMISSION_SECTIONS)[number];
export type PermissionKey = `${PermissionSection}:read` | `${PermissionSection}:write`;

export const PERMISSION_SECTION_LABELS: Record<PermissionSection, string> = {
  allocations: "Allocations",
  projects: "Projects",
  tracing: "Tracing",
  amie: "AMIE",
  users_permissions: "Users & Permissions",
};

export function rwStateFor(permissions: PermissionKey[]) {
  const held = new Set(permissions);
  return PERMISSION_SECTIONS.map((section) => ({
    section,
    label: PERMISSION_SECTION_LABELS[section],
    read: held.has(`${section}:read` as PermissionKey),
    write: held.has(`${section}:write` as PermissionKey),
  }));
}

export function togglePermission(permissions: PermissionKey[], key: PermissionKey): PermissionKey[] {
  return permissions.includes(key) ? permissions.filter((p) => p !== key) : [...permissions, key];
}

// A user's effective permissions are the union of everything granted by
// each role they hold.
export function permissionsFromRoles(roles: Array<{ permissions: PermissionKey[] }>): PermissionKey[] {
  return Array.from(new Set(roles.flatMap((r) => r.permissions)));
}
