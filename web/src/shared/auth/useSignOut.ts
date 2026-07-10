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

import { useCallback, useState } from "react";

export type SignOutOptions = { callbackUrl?: string };

export function useSignOut() {
  const [isPending, setIsPending] = useState(false);
  const signOut = useCallback((opts?: SignOutOptions) => {
    const callbackUrl = opts?.callbackUrl ?? "/sign-in";
    setIsPending(true);
    // The end-session route reads idToken from the session before clearing
    // the cookie, so it can build a Keycloak logout URL with id_token_hint
    // (otherwise Keycloak skips the SSO invalidation).
    window.location.assign(
      `/api/auth/end-session?callbackUrl=${encodeURIComponent(callbackUrl)}`,
    );
  }, []);
  return { signOut, isPending };
}
