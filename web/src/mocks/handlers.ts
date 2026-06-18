import type { RequestHandler } from "msw";
import { allocationsHandlers } from "./handlers/allocations";
import { amieHandlers } from "./handlers/amie";
import { healthzHandlers } from "./handlers/healthz";
import { privilegesHandlers } from "./handlers/privileges";
import { projectsHandlers } from "./handlers/projects";
import { tracesHandlers } from "./handlers/traces";

export const handlers: RequestHandler[] = [
  ...healthzHandlers,
  ...privilegesHandlers,
  ...projectsHandlers,
  ...allocationsHandlers,
  ...tracesHandlers,
  ...amieHandlers,
];
