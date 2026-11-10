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
import {
  type AppConnector,
  type AppView,
  type AppWithTokenView,
  zAppConnector,
  zAppView,
  zAppWithTokenView,
} from "@/generated/core/zod.gen";
import { apiFetch } from "@/shared/api/client";

export type App = AppView;
export type AppWithToken = AppWithTokenView;
export type { AppConnector };

export type NewApp = {
  key: string;
  name: string;
  description: string;
  actions: string[];
};

const appPath = (id: string) => `/apps/${encodeURIComponent(id)}`;

export async function listApps(): Promise<App[]> {
  return z.array(zAppView).parse(await apiFetch("/apps"));
}

export async function listAppConnectors(): Promise<AppConnector[]> {
  return z.array(zAppConnector).parse(await apiFetch("/apps/connectors"));
}

export async function createApp(body: NewApp): Promise<AppWithToken> {
  return zAppWithTokenView.parse(await apiFetch("/apps", { method: "POST", body }));
}

export async function updateApp({
  id,
  ...body
}: {
  id: string;
  name: string;
  description: string;
}): Promise<App> {
  return zAppView.parse(await apiFetch(appPath(id), { method: "PUT", body }));
}

export async function disableApp(id: string): Promise<App> {
  return zAppView.parse(await apiFetch(`${appPath(id)}/disable`, { method: "POST" }));
}

export async function enableApp(id: string): Promise<App> {
  return zAppView.parse(await apiFetch(`${appPath(id)}/enable`, { method: "POST" }));
}

export async function issueAppSecret(id: string): Promise<AppWithToken> {
  return zAppWithTokenView.parse(await apiFetch(`${appPath(id)}/secret`, { method: "POST" }));
}
