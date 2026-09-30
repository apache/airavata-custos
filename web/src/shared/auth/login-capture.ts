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

import "server-only";

interface CaptureOptions {
  bearer: string;
  coreApiBaseUrl: string;
  fetcher?: typeof fetch;
  sleeper?: (milliseconds: number) => Promise<void>;
}

interface InitialOIDCAccount {
  access_token?: string;
  id_token?: string;
}

interface InitialLoginOptions {
  account?: InitialOIDCAccount | null;
  coreApiBaseUrl: string;
  fetcher?: typeof fetch;
  capture?: typeof captureLoginEvent;
}

export function loginCaptureBearer(account: InitialOIDCAccount): string | undefined {
  if (account.access_token?.split(".").length === 3) return account.access_token;
  return account.id_token ?? account.access_token;
}

export async function captureInitialOIDCLogin({
  account,
  coreApiBaseUrl,
  fetcher = fetch,
  capture = captureLoginEvent,
}: InitialLoginOptions): Promise<boolean | undefined> {
  if (!account?.access_token) return undefined;
  const bearer = loginCaptureBearer(account);
  if (!bearer) return false;
  return capture({ bearer, coreApiBaseUrl, fetcher });
}

export async function captureLoginEvent({
  bearer,
  coreApiBaseUrl,
  fetcher = fetch,
  sleeper = (milliseconds) => new Promise((resolve) => setTimeout(resolve, milliseconds)),
}: CaptureOptions): Promise<boolean> {
  const payload = JSON.stringify({});
  const endpoint = new URL("/me/login-events", coreApiBaseUrl).toString();

  for (let attempt = 0; attempt < 3; attempt += 1) {
    try {
      const response = await fetcher(endpoint, {
        method: "POST",
        headers: {
          authorization: `Bearer ${bearer}`,
          "content-type": "application/json",
        },
        body: payload,
        cache: "no-store",
        signal: AbortSignal.timeout(2000),
      });
      if (response.ok) return true;
      if (response.status >= 400 && response.status < 500) return false;
    } catch {
      // Network failures are transient and retry with the same stable payload.
    }
    if (attempt < 2) await sleeper(50 * 2 ** attempt);
  }
  return false;
}
