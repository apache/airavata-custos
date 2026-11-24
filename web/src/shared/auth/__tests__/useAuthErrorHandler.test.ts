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
import { beforeEach, describe, expect, it, vi } from "vitest";
import { ApiError } from "@/shared/api/client";

const update = vi.fn((_data?: unknown) => Promise.resolve(null));

vi.mock("next-auth/react", () => ({
  signOut: vi.fn(),
  useSession: () => ({ update }),
}));
vi.mock("next/navigation", () => ({ useRouter: () => ({ refresh: vi.fn() }) }));

import { useAuthErrorHandler } from "../useAuthErrorHandler";

beforeEach(() => update.mockClear());

describe("useAuthErrorHandler", () => {
  // Checks that a consent_required 403 reloads the session flags so the policy page comes up.
  // update() needs an argument, or next-auth only re-reads the cookie.
  it("reloads the session on a consent_required 403", () => {
    const { result } = renderHook(() => useAuthErrorHandler());
    result.current(new ApiError(403, "/allocations", { code: "consent_required" }));
    expect(update).toHaveBeenCalledWith({});
  });
});
