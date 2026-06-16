import type { Privilege } from "@/features/core/identity/types";

export type DevLevel = "viewer" | "manager" | "admin";

// Dev-mode privilege bundles. Levels are coarser than the spec roles so the
// dropdown stays simple; admin covers the full PrivilegeKey enum.
export const DEV_LEVEL_PRIVILEGES: Record<DevLevel, Privilege[]> = {
  viewer: ["hpc:read"],
  manager: ["hpc:read", "hpc:write", "amie:read"],
  admin: [
    "amie:read",
    "amie:write",
    "hpc:read",
    "hpc:write",
    "signer:read",
    "signer:write",
    "privileges:grant",
    "roles:manage",
  ],
};

export const DEV_LEVEL_NAMES: Record<DevLevel, string> = {
  viewer: "Dev Viewer",
  manager: "Dev Manager",
  admin: "Dev Admin",
};
