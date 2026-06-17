import { http, HttpResponse } from "msw";
import type { Privilege } from "@/features/core/identity/types";

// Default to admin-grade so MSW-only browsing exercises the full UI; tests
// override per-case via server.use().
const adminPrivileges: Privilege[] = [
  "amie:read",
  "amie:write",
  "hpc:read",
  "hpc:write",
  "signer:read",
  "signer:write",
  "privileges:grant",
  "roles:manage",
];

export const privilegesHandlers = [
  http.get("*/api/v1/user/privileges", () =>
    HttpResponse.json({ privileges: adminPrivileges }),
  ),
];
