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

const MESSAGES: Record<string, string> = {
  AccessDenied: "Your account isn't permitted to sign in.",
  Configuration: "Sign-in is misconfigured. Contact your administrator.",
  OAuthCallback: "Sign-in didn't complete. Try again.",
  OAuthSignin: "Sign-in didn't complete. Try again.",
  Verification: "Sign-in didn't complete. Try again.",
  identity_not_linked: "Your identity isn't linked to a portal user yet. Ask an administrator to link it.",
  session_expired: "Your session expired. Sign in again to continue.",
};

const DEFAULT_MESSAGE = "Sign-in failed. Try again.";

export function AuthErrorAlert({ code }: { code: string | null }) {
  if (!code) return null;
  const text = MESSAGES[code] ?? DEFAULT_MESSAGE;
  return (
    <div
      role="alert"
      className="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive"
    >
      {text}
    </div>
  );
}
