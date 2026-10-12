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

import { renderHook } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("next-auth/react", () => ({ signOut: vi.fn() }));

import { signOut } from "next-auth/react";

// The sign-out guard is module state, so each test loads a fresh copy.
async function load() {
  vi.resetModules();
  const { ApiError } = await import("@/shared/api/client");
  const { useAuthErrorHandler } = await import("../useAuthErrorHandler");
  return { ApiError, handle: renderHook(() => useAuthErrorHandler()).result.current };
}

beforeEach(() => vi.mocked(signOut).mockClear());
afterEach(() => vi.unstubAllGlobals());

describe("useAuthErrorHandler", () => {
  it("signs a suspended user out of the IdP once, however many queries are refused", async () => {
    const assign = vi.fn();
    vi.stubGlobal("location", { ...window.location, assign });
    const { ApiError, handle } = await load();
    const disabled = () =>
      new ApiError(403, "/api/v1/me", {
        code: "user_disabled",
        error: "user account is not active",
      });
    handle(disabled());
    handle(disabled());
    handle(new ApiError(401, "/api/v1/me", { code: "invalid_token", error: "expired" }));
    expect(assign).toHaveBeenCalledOnce();
    expect(assign).toHaveBeenCalledWith(
      `/api/auth/end-session?callbackUrl=${encodeURIComponent("/sign-in?error=user_disabled")}`,
    );
    expect(signOut).not.toHaveBeenCalled();
  });

  it("returns an expired session to the page it was on after signing in again", async () => {
    vi.stubGlobal("location", {
      ...window.location,
      pathname: "/allocations/a-1",
      search: "?tab=usage",
    });
    const { ApiError, handle } = await load();
    handle(new ApiError(401, "/api/v1/me", { code: "invalid_token", error: "expired" }));
    expect(signOut).toHaveBeenCalledWith({
      callbackUrl: `/sign-in?error=session_expired&callbackUrl=${encodeURIComponent("/allocations/a-1?tab=usage")}`,
    });
  });
});
