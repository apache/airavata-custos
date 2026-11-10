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
import type { App, AppConnector, NewApp } from "@/features/core/apps/api";

const connectors: AppConnector[] = [
  {
    name: "signer",
    needs_user_grant: true,
    actions: ["signer:app:generate-certificate", "signer:app:revoke-certificate"],
  },
  { name: "temp-account", needs_user_grant: false, actions: ["temp-account:app:create-user"] },
];

const apps: App[] = [
  {
    id: "app-1",
    key: "notebook-hub",
    name: "Notebook Hub",
    description: "Creates workshop accounts.",
    connector: "temp-account",
    status: "ACTIVE",
    secret_hint: "9F3A",
    needs_user_authorization: false,
    actions: ["temp-account:app:create-user"],
    created_by: "user-admin",
    created_at: "2026-11-01T12:00:00Z",
    updated_at: "2026-11-01T12:00:00Z",
  },
];

function setStatus(id: unknown, status: App["status"]) {
  const app = apps.find((a) => a.id === id);
  if (!app) return HttpResponse.json({ error: "not found" }, { status: 404 });
  app.status = status;
  return HttpResponse.json(app);
}

export const appsHandlers = [
  http.get("*/api/v1/apps", () => HttpResponse.json(apps)),
  http.get("*/api/v1/apps/connectors", () => HttpResponse.json(connectors)),
  http.get("*/api/v1/apps/names", () =>
    HttpResponse.json(apps.map(({ key, name }) => ({ key, name }))),
  ),
  http.put("*/api/v1/apps/:id", async ({ params: { id }, request }) => {
    const app = apps.find((a) => a.id === id);
    if (!app) return HttpResponse.json({ error: "not found" }, { status: 404 });
    Object.assign(app, await request.json());
    return HttpResponse.json(app);
  }),
  http.post("*/api/v1/apps", async ({ request }) => {
    const body = (await request.json()) as NewApp;
    const connector = body.actions[0]?.split(":")[0] ?? "";
    const app: App = {
      ...body,
      id: `app-${apps.length + 1}`,
      connector,
      status: "ACTIVE",
      secret_hint: "ABCD",
      needs_user_authorization: connector === "signer",
      created_by: "user-admin",
      created_at: "2026-11-10T12:00:00Z",
      updated_at: "2026-11-10T12:00:00Z",
    };
    apps.push(app);
    return HttpResponse.json({ ...app, token: `cak_${app.key}.ABCDEFGH` }, { status: 201 });
  }),
  http.post("*/api/v1/apps/:id/disable", ({ params: { id } }) => setStatus(id, "DISABLED")),
  http.post("*/api/v1/apps/:id/enable", ({ params: { id } }) => setStatus(id, "ACTIVE")),
  http.post("*/api/v1/apps/:id/secret", ({ params: { id } }) => {
    const app = apps.find((a) => a.id === id);
    if (!app) return HttpResponse.json({ error: "not found" }, { status: 404 });
    return HttpResponse.json({ ...app, token: `cak_${app.key}.NEWSECRET` });
  }),
];
