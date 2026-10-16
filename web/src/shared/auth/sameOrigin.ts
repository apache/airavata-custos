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

import { serverEnv } from "@/lib/env";

// Same-origin only: "//host", "/\host" and tab-split paths resolve off-site, so they fall back to
// the portal root. An absolute URL on the portal origin is reduced to its path, query and hash.
// Behind a reverse proxy the request origin is the internal address, so NEXTAUTH_URL wins.
export function sameOriginUrl(callbackUrl: string | null | undefined, requestOrigin: string): URL {
  const base = new URL(serverEnv.NEXTAUTH_URL ?? requestOrigin);
  const target = callbackUrl ?? "";
  const url = target.startsWith("/") || URL.canParse(target) ? new URL(target, base) : base;
  return url.origin === base.origin ? new URL(url.pathname + url.search + url.hash, base) : base;
}
