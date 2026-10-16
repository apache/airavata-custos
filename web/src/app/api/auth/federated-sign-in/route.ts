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

import { signIn } from "@/shared/auth/auth";
import { sameOriginUrl } from "@/shared/auth/sameOrigin";
import type { NextRequest } from "next/server";

export const runtime = "nodejs";

// Starts the OIDC flow directly so a plain link can skip the sign-in card.
export async function GET(request: NextRequest) {
  const { pathname, search, hash } = sameOriginUrl(
    request.nextUrl.searchParams.get("callbackUrl"),
    request.nextUrl.origin,
  );
  await signIn("oidc", { redirectTo: pathname + search + hash });
}
