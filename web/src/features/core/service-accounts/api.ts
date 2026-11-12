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
import { zServiceAccount, zServiceAccountWithTokenView } from "@/generated/core/zod.gen";
import { apiFetch } from "@/shared/api/client";

const privilegesSchema = z.array(privilegeKeySchema).nullable();
const serviceAccountSchema = zServiceAccount.extend({ privileges: privilegesSchema });
const serviceAccountWithTokenSchema = zServiceAccountWithTokenView.extend({
  privileges: privilegesSchema,
});

export type ServiceAccount = z.infer<typeof serviceAccountSchema>;
export type ServiceAccountWithToken = z.infer<typeof serviceAccountWithTokenSchema>;

export type NewServiceAccount = {
  key: string;
  name: string;
  description: string;
  privileges: string[];
};

const serviceAccountPath = (id: string) => `/service-accounts/${encodeURIComponent(id)}`;

export async function listServiceAccounts(): Promise<ServiceAccount[]> {
  return z.array(serviceAccountSchema).parse(await apiFetch("/service-accounts"));
}

export async function listServiceAccountPrivileges(): Promise<string[]> {
  return z.array(privilegeKeySchema).parse(await apiFetch("/service-accounts/privileges"));
}

export async function createServiceAccount(
  body: NewServiceAccount,
): Promise<ServiceAccountWithToken> {
  return serviceAccountWithTokenSchema.parse(
    await apiFetch("/service-accounts", { method: "POST", body }),
  );
}

export async function updateServiceAccount({
  id,
  ...body
}: {
  id: string;
  name: string;
  description: string;
}): Promise<ServiceAccount> {
  return serviceAccountSchema.parse(
    await apiFetch(serviceAccountPath(id), { method: "PUT", body }),
  );
}

export async function disableServiceAccount(id: string): Promise<ServiceAccount> {
  return serviceAccountSchema.parse(
    await apiFetch(`${serviceAccountPath(id)}/disable`, { method: "POST" }),
  );
}

export async function enableServiceAccount(id: string): Promise<ServiceAccount> {
  return serviceAccountSchema.parse(
    await apiFetch(`${serviceAccountPath(id)}/enable`, { method: "POST" }),
  );
}

export async function deleteServiceAccount(id: string): Promise<void> {
  await apiFetch(serviceAccountPath(id), { method: "DELETE" });
}

export async function issueServiceAccountSecret(id: string): Promise<ServiceAccountWithToken> {
  return serviceAccountWithTokenSchema.parse(
    await apiFetch(`${serviceAccountPath(id)}/secret`, { method: "POST" }),
  );
}
