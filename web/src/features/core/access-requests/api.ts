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

import { ApiError, apiFetch } from "@/shared/api/client";
import {
  type AccessEventResolve,
  type AccessRequest,
  type UsernameCheck,
  accessEventResolveSchema,
  accessRequestSchema,
  accessRequestsSchema,
  usernameCheckSchema,
} from "./schemas";

// 404 means "no request yet" / "unknown code" — a state, not a failure.
async function orNullOn404<T>(fetchIt: () => Promise<T>): Promise<T | null> {
  try {
    return await fetchIt();
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) return null;
    throw err;
  }
}

export async function getMyAccessRequest(): Promise<AccessRequest | null> {
  return orNullOn404(async () => accessRequestSchema.parse(await apiFetch("/access-requests/me")));
}

export async function resolveAccessEvent(code: string): Promise<AccessEventResolve | null> {
  return orNullOn404(async () =>
    accessEventResolveSchema.parse(
      await apiFetch(`/access-requests/events/${encodeURIComponent(code)}`),
    ),
  );
}

export async function checkAccessRequestUsername(
  eventCode: string,
  username: string,
): Promise<UsernameCheck> {
  const search = new URLSearchParams({ event_code: eventCode });
  if (username) search.set("username", username);
  return usernameCheckSchema.parse(await apiFetch(`/access-requests/username?${search.toString()}`));
}

export type CreateAccessRequestBody = {
  institution: string;
  event_code: string;
  desired_username?: string;
  reason?: string;
};

export async function createAccessRequest(body: CreateAccessRequestBody): Promise<AccessRequest> {
  const raw = await apiFetch("/access-requests", { method: "POST", body });
  return accessRequestSchema.parse(raw);
}

export type AccessRequestListFilter = {
  status?: string;
  event?: string;
  limit?: number;
};

export async function listAccessRequests(
  filter: AccessRequestListFilter = {},
): Promise<AccessRequest[]> {
  const search = new URLSearchParams();
  if (filter.status) search.set("status", filter.status);
  if (filter.event) search.set("event", filter.event);
  if (typeof filter.limit === "number") search.set("limit", String(filter.limit));
  const qs = search.toString();
  const raw = await apiFetch(`/access-requests${qs ? `?${qs}` : ""}`);
  return accessRequestsSchema.parse(raw);
}

export type DecideAccessRequestBody = {
  status: "APPROVED" | "DENIED";
  expires_at?: string;
  deny_reason?: string;
};

export async function decideAccessRequest(
  id: string,
  body: DecideAccessRequestBody,
): Promise<AccessRequest> {
  const raw = await apiFetch(`/access-requests/${encodeURIComponent(id)}`, {
    method: "PUT",
    body,
  });
  return accessRequestSchema.parse(raw);
}
