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

import { describe, expect, it, vi } from "vitest";

vi.mock("@/shared/auth/auth", () => ({ auth: async () => null }));
vi.mock("next/headers", () => ({
  headers: async () => new Headers({ "x-portal-path": "/allocations/a-1?tab=usage" }),
}));
vi.mock("next/navigation", () => ({
  redirect: vi.fn((url: string) => {
    throw new Error(`redirect ${url}`);
  }),
}));

import PortalRoutesLayout from "../layout";

describe("portal layout", () => {
  it("sends a signed-out deep link to sign-in with the path as its callback", async () => {
    await expect(PortalRoutesLayout({ children: null })).rejects.toThrow(
      `redirect /sign-in?callbackUrl=${encodeURIComponent("/allocations/a-1?tab=usage")}`,
    );
  });
});
