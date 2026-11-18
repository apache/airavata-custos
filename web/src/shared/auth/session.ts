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

// Server-side reads of the session cookie; its tokens never reach browser JS.
import "server-only";
import { serverEnv } from "@/lib/env";
import { type JWT, getToken } from "next-auth/jwt";
import { backendBearerOf, refreshIfDue, sessionCookieName } from "./auth";

export function readSessionCookie(request: Request): Promise<JWT | null> {
  return getToken({
    req: request,
    secret: serverEnv.NEXTAUTH_SECRET,
    cookieName: sessionCookieName,
  });
}

export async function freshBackendBearer(request: Request): Promise<string | undefined> {
  const token = await readSessionCookie(request);
  return token ? backendBearerOf(await refreshIfDue(token)) : undefined;
}
