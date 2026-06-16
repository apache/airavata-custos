import { http, HttpResponse } from "msw";
import { DEV_LEVEL_PRIVILEGES } from "@/shared/auth/devLevels";

// Default to admin-grade so MSW-only browsing exercises the full UI; tests
// override per-case via server.use().
export const privilegesHandlers = [
  http.get("*/api/v1/user/privileges", () =>
    HttpResponse.json({ privileges: DEV_LEVEL_PRIVILEGES.admin }),
  ),
];
