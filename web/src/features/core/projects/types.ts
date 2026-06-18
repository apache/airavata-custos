import type { ProjectStatus } from "./schemas";

export type ProjectListParams = {
  limit?: number;
  offset?: number;
  pi_id?: string;
  status?: ProjectStatus;
  q?: string;
};
