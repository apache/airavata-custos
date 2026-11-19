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

// The issuer's OIDC discovery document, for the endpoints Auth.js does not expose.
import "server-only";
import { serverEnv } from "@/lib/env";
import { z } from "zod";

const discoverySchema = z.object({
  token_endpoint: z.string().optional(),
  end_session_endpoint: z.string().optional(),
});

let cached: z.infer<typeof discoverySchema> | undefined;

// Null when the issuer is unreachable or its document is malformed; only a success is cached.
export async function discoverOidcEndpoints() {
  if (cached) return cached;
  try {
    const issuer = serverEnv.OIDC_ISSUER_URL.replace(/\/$/, "");
    const res = await fetch(`${issuer}/.well-known/openid-configuration`, { cache: "no-store" });
    if (!res.ok) return null;
    const doc = discoverySchema.safeParse(await res.json());
    if (!doc.success) return null;
    cached = doc.data;
    return cached;
  } catch {
    return null;
  }
}
