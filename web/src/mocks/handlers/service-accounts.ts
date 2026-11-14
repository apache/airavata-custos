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

import { HttpResponse, http } from "msw";
import type { NewServiceAccount, ServiceAccount } from "@/features/core/service-accounts/api";

const serviceAccountPrivileges = ["temp-account:users:read", "temp-account:users:create"];

const accounts: ServiceAccount[] = [
  {
    id: "service-account-1",
    key: "notebook-hub",
    name: "Notebook Hub",
    description: "Creates workshop accounts.",
    privileges: ["temp-account:users:create"],
    status: "ACTIVE",
    secret_hint: "9F3A",
    created_by: "user-admin",
    created_at: "2026-11-01T12:00:00Z",
    updated_at: "2026-11-01T12:00:00Z",
  },
];

function setStatus(id: unknown, status: ServiceAccount["status"]) {
  const account = accounts.find((a) => a.id === id);
  if (!account) return HttpResponse.json({ error: "not found" }, { status: 404 });
  account.status = status;
  return HttpResponse.json(account);
}

export const serviceAccountsHandlers = [
  http.get("*/api/v1/service-accounts", () =>
    HttpResponse.json(accounts.filter((a) => a.status !== "DELETED")),
  ),
  http.get("*/api/v1/service-accounts/privileges", () =>
    HttpResponse.json(serviceAccountPrivileges),
  ),
  http.delete("*/api/v1/service-accounts/:id", ({ params: { id } }) => {
    const account = accounts.find((a) => a.id === id && a.status !== "DELETED");
    if (!account) return HttpResponse.json({ error: "not found" }, { status: 404 });
    if (account.status !== "DISABLED") {
      return HttpResponse.json(
        { error: "disable the service account before deleting it" },
        { status: 409 },
      );
    }
    account.status = "DELETED";
    return new HttpResponse(null, { status: 204 });
  }),
  http.put("*/api/v1/service-accounts/:id", async ({ params: { id }, request }) => {
    const account = accounts.find((a) => a.id === id);
    if (!account) return HttpResponse.json({ error: "not found" }, { status: 404 });
    Object.assign(account, await request.json());
    return HttpResponse.json(account);
  }),
  http.post("*/api/v1/service-accounts", async ({ request }) => {
    const body = (await request.json()) as NewServiceAccount;
    const account: ServiceAccount = {
      ...body,
      id: `service-account-${accounts.length + 1}`,
      status: "ACTIVE",
      secret_hint: "ABCD",
      created_by: "user-admin",
      created_at: "2026-11-10T12:00:00Z",
      updated_at: "2026-11-10T12:00:00Z",
    };
    accounts.push(account);
    return HttpResponse.json({ ...account, token: `csa_${account.key}.ABCDEFGH` }, { status: 201 });
  }),
  http.post("*/api/v1/service-accounts/:id/disable", ({ params: { id } }) =>
    setStatus(id, "DISABLED"),
  ),
  http.post("*/api/v1/service-accounts/:id/enable", ({ params: { id } }) =>
    setStatus(id, "ACTIVE"),
  ),
  http.post("*/api/v1/service-accounts/:id/secret", ({ params: { id } }) => {
    const account = accounts.find((a) => a.id === id);
    if (!account) return HttpResponse.json({ error: "not found" }, { status: 404 });
    return HttpResponse.json({ ...account, token: `csa_${account.key}.NEWSECRET` });
  }),
];
