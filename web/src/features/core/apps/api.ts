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

import { z } from "zod";
import { privilegeKeySchema } from "@/features/core/identity/schemas";
import {
  type AppAuthorizationView,
  type AppMode,
  zApp,
  zAppAuthorizationView,
  zAppWithTokenView,
} from "@/generated/core/zod.gen";
import { apiFetch } from "@/shared/api/client";

const privilegesSchema = z.array(privilegeKeySchema).nullable();
const appSchema = zApp.extend({ privileges: privilegesSchema });
const appWithTokenSchema = zAppWithTokenView.extend({ privileges: privilegesSchema });

export type { AppMode };
export type App = z.infer<typeof appSchema>;
export type AppWithToken = z.infer<typeof appWithTokenSchema>;
export type AppAuthorization = AppAuthorizationView;

export type NewApp = {
  key: string;
  name: string;
  description: string;
  mode: AppMode;
  privileges: string[];
};

const appPath = (id: string) => `/apps/${encodeURIComponent(id)}`;

export async function listApps(): Promise<App[]> {
  return z.array(appSchema).parse(await apiFetch("/apps"));
}

export async function listAppPrivileges(): Promise<string[]> {
  return z.array(privilegeKeySchema).parse(await apiFetch("/apps/privileges"));
}

export async function createApp(body: NewApp): Promise<AppWithToken> {
  return appWithTokenSchema.parse(await apiFetch("/apps", { method: "POST", body }));
}

export async function updateApp({
  id,
  ...body
}: {
  id: string;
  name: string;
  description: string;
}): Promise<App> {
  return appSchema.parse(await apiFetch(appPath(id), { method: "PUT", body }));
}

export async function disableApp(id: string): Promise<App> {
  return appSchema.parse(await apiFetch(`${appPath(id)}/disable`, { method: "POST" }));
}

export async function enableApp(id: string): Promise<App> {
  return appSchema.parse(await apiFetch(`${appPath(id)}/enable`, { method: "POST" }));
}

// How many users would lose the app if it were deleted.
export async function countAppAuthorizations(id: string): Promise<number> {
  const app = z.object({ authorized_users: z.number().int() }).parse(await apiFetch(appPath(id)));
  return app.authorized_users;
}

export async function deleteApp(id: string): Promise<void> {
  await apiFetch(appPath(id), { method: "DELETE" });
}

export async function issueAppSecret(id: string): Promise<AppWithToken> {
  return appWithTokenSchema.parse(await apiFetch(`${appPath(id)}/secret`, { method: "POST" }));
}

const authorizationPath = (appId: string) => `/me/app-authorizations/${encodeURIComponent(appId)}`;

export async function listAppAuthorizations(): Promise<AppAuthorization[]> {
  return z.array(zAppAuthorizationView).parse(await apiFetch("/me/app-authorizations"));
}

export async function authorizeApp(appId: string): Promise<void> {
  await apiFetch(authorizationPath(appId), { method: "POST" });
}

export async function removeAppAuthorization(appId: string): Promise<void> {
  await apiFetch(authorizationPath(appId), { method: "DELETE" });
}
