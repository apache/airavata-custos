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

const session = vi.hoisted(() => ({ current: {} as { custosUserId?: string } }));
vi.mock("@/shared/auth/auth", () => ({
  auth: async () => ({ user: { name: "Ada" }, ...session.current }),
}));
vi.mock("@/lib/env", () => ({ serverEnv: {} }));
vi.mock("next/headers", () => ({ headers: async () => new Headers({ host: "portal.test" }) }));
const redirect = vi.hoisted(() =>
  vi.fn(() => {
    throw new Error("NEXT_REDIRECT");
  }),
);
vi.mock("next/navigation", () => ({ redirect }));

import SignInPage from "../page";

describe("sign-in page", () => {
  it("sends a signed-in user to the same-origin callback, not the IdP", async () => {
    session.current = { custosUserId: "u-1" };
    redirect.mockClear();
    const visit = (callbackUrl: string) =>
      expect(SignInPage({ searchParams: Promise.resolve({ callbackUrl }) })).rejects.toThrow();
    await visit("https://evil.test/x");
    await visit("/projects?q=1");
    expect(redirect.mock.calls).toEqual([["/"], ["/projects?q=1"]]);
  });

  it("sends a sign-in without a Custos user back to the IdP", async () => {
    session.current = {};
    redirect.mockClear();
    await expect(SignInPage({ searchParams: Promise.resolve({}) })).rejects.toThrow();
    expect(redirect).toHaveBeenCalledWith("/api/auth/federated-sign-in");
  });
});
