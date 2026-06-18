import type { AllocationStatus, ChangeRequestStatus } from "./schemas";

export type AllocationListParams = {
  limit?: number;
  offset?: number;
  project_id?: string;
  status?: AllocationStatus;
  q?: string;
};

export type ChangeRequestListParams = {
  status?: ChangeRequestStatus;
  allocation_id?: string;
  requester_id?: string;
};

// Drawer/form-side discriminated union. The backend only models
// requested_su_amount + requested_status + reason; the change "type" is
// derived from which field the user filled in (see drawer submit handler).
export type ChangeRequestType = "INCREASE_CREDITS" | "EXTEND_END_DATE" | "OTHER";

export type ChangeRequestFormValues =
  | { requested_change_type: "INCREASE_CREDITS"; requested_amount: number; reason: string }
  | { requested_change_type: "EXTEND_END_DATE"; requested_end_date: string; reason: string }
  | { requested_change_type: "OTHER"; reason: string };
