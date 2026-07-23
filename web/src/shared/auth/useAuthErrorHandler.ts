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

"use client";

import { signOut } from "next-auth/react";
import { useRouter } from "next/navigation";
import { useCallback } from "react";
import { ApiError } from "@/shared/api/client";

type ErrorBody = { code?: unknown };

function readCode(body: unknown): string | null {
  if (!body || typeof body !== "object") return null;
  const code = (body as ErrorBody).code;
  return typeof code === "string" ? code : null;
}

export function useAuthErrorHandler() {
  const router = useRouter();
  return useCallback(
    (error: unknown) => {
      if (!(error instanceof ApiError)) return;
      const code = readCode(error.body);

      // Sign the user out and let the sign-in page surface the reason as a
      // toast; a toast fired here would be lost across the sign-out redirect.
      if (error.status === 401 && (code === "invalid_token" || code === "missing_bearer")) {
        void signOut({ callbackUrl: "/sign-in?error=session_expired" });
        return;
      }
      // A signed-in user whose identity is not linked keeps their session and
      // goes to the request-access page, not signed out.
      if (error.status === 401 && code === "identity_not_linked") {
        router.push("/no-access?reason=identity_not_linked");
        return;
      }
      // 403 insufficient_privilege and other codes fall through to the
      // query's local ErrorState — no global takeover.
    },
    [router],
  );
}
