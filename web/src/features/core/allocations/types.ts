// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

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
