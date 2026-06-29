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
import { auth } from "./auth";

export type PortalSession = {
  accessToken?: string | null;
  idToken?: string | null;
  userId?: string | null;
} | null;

export async function getPortalSession(): Promise<PortalSession> {
  const session = await auth();
  if (!session) return null;
  return {
    accessToken: session.accessToken ?? null,
    idToken: session.idToken ?? null,
    userId: session.user?.id ?? null,
  };
}

// Some IdPs (CILogon) hand out opaque access tokens while their id_tokens are
// JWTs; the backend only verifies JWTs, so prefer whichever bearer actually
// has the JWT shape. Keycloak access_tokens are JWTs so they win there.
export function pickBackendBearer(session: PortalSession): string | null {
  if (!session) return null;
  if (looksLikeJwt(session.accessToken)) return session.accessToken ?? null;
  if (looksLikeJwt(session.idToken)) return session.idToken ?? null;
  return session.accessToken ?? session.idToken ?? null;
}

function looksLikeJwt(token: string | null | undefined): boolean {
  return typeof token === "string" && token.split(".").length === 3;
}
