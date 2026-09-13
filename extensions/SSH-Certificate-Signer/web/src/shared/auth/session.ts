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
import { headers } from "next/headers";
import { getToken } from "next-auth/jwt";
import { serverEnv } from "@/lib/env";

export type SignerSession = {
  accessToken: string | null;
  idToken: string | null;
  privileges: string[];
  name: string | null;
  email: string | null;
};

type RequestWithHeaders = Request | { headers: Headers | Record<string, string> };

export async function decodeSharedSession(req: RequestWithHeaders): Promise<SignerSession | null> {
  const cookieName =
    serverEnv.NODE_ENV === "production" ? "__Secure-custos.session-token" : "custos.session-token";
  const token = await getToken({
    req,
    secret: serverEnv.NEXTAUTH_SECRET,
    salt: cookieName,
    cookieName,
    secureCookie: serverEnv.NODE_ENV === "production",
  });
  if (!token) return null;

  return {
    accessToken: typeof token.accessToken === "string" ? token.accessToken : null,
    idToken: typeof token.idToken === "string" ? token.idToken : null,
    privileges: Array.isArray(token.privileges)
      ? token.privileges.filter((value): value is string => typeof value === "string")
      : [],
    name: typeof token.name === "string" ? token.name : null,
    email: typeof token.email === "string" ? token.email : null,
  };
}

export async function getSharedSession(): Promise<SignerSession | null> {
  return decodeSharedSession({ headers: await headers() });
}

export function pickBackendBearer(session: SignerSession | null): string | null {
  if (!session) return null;
  if (looksLikeJwt(session.accessToken)) return session.accessToken;
  if (looksLikeJwt(session.idToken)) return session.idToken;
  return session.accessToken ?? session.idToken;
}

function looksLikeJwt(value: string | null): boolean {
  return typeof value === "string" && value.split(".").length === 3;
}
