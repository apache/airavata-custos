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
import type { App, NewApp } from "@/features/core/apps/api";

const appPrivileges = ["temp-account:users:create"];

const apps: App[] = [
  {
    id: "app-1",
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

function setStatus(id: unknown, status: App["status"]) {
  const app = apps.find((a) => a.id === id);
  if (!app) return HttpResponse.json({ error: "not found" }, { status: 404 });
  app.status = status;
  return HttpResponse.json(app);
}

export const appsHandlers = [
  http.get("*/api/v1/apps", () => HttpResponse.json(apps.filter((a) => a.status !== "DELETED"))),
  http.get("*/api/v1/apps/privileges", () => HttpResponse.json(appPrivileges)),
  http.delete("*/api/v1/apps/:id", ({ params: { id } }) => {
    const app = apps.find((a) => a.id === id && a.status !== "DELETED");
    if (!app) return HttpResponse.json({ error: "not found" }, { status: 404 });
    if (app.status !== "DISABLED") {
      return HttpResponse.json({ error: "disable the app before deleting it" }, { status: 409 });
    }
    app.status = "DELETED";
    return new HttpResponse(null, { status: 204 });
  }),
  http.put("*/api/v1/apps/:id", async ({ params: { id }, request }) => {
    const app = apps.find((a) => a.id === id);
    if (!app) return HttpResponse.json({ error: "not found" }, { status: 404 });
    Object.assign(app, await request.json());
    return HttpResponse.json(app);
  }),
  http.post("*/api/v1/apps", async ({ request }) => {
    const body = (await request.json()) as NewApp;
    const app: App = {
      ...body,
      id: `app-${apps.length + 1}`,
      status: "ACTIVE",
      secret_hint: "ABCD",
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
