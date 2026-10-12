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

import { ApiError, errorCode } from "@/shared/api/client";
import { signOut } from "next-auth/react";
import { useCallback } from "react";
import { endSession, leaveOnce } from "./useSignOut";

export function useAuthErrorHandler() {
  return useCallback((error: unknown) => {
    if (!(error instanceof ApiError)) return;
    const code = errorCode(error);

    // Sign the user out and let the sign-in page surface the reason as a
    // toast; a toast fired here would be lost across the sign-out redirect.
    if (error.status === 401 && (code === "invalid_token" || code === "missing_bearer")) {
      const here = encodeURIComponent(window.location.pathname + window.location.search);
      leaveOnce(() => {
        void signOut({ callbackUrl: `/sign-in?error=session_expired&callbackUrl=${here}` });
      });
      return;
    }
    // These end the IdP session too, or SSO would sign straight back into the same refusal.
    if (
      (error.status === 401 && code === "identity_not_linked") ||
      (error.status === 403 && code === "user_disabled")
    ) {
      endSession(`/sign-in?error=${code}`);
      return;
    }
    // 403 insufficient_privilege and other codes fall through to the
    // query's local ErrorState — no global takeover.
  }, []);
}
