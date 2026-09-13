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

// @vitest-environment node

import { encode } from "next-auth/jwt";
import { beforeEach, describe, expect, it, vi } from "vitest";

const env = vi.hoisted(() => ({
  NODE_ENV: "test" as "test" | "production",
  NEXTAUTH_SECRET: "shared-session-secret-for-tests-32chars",
}));

vi.mock("@/lib/env", () => ({ serverEnv: env }));

import { decodeSharedSession, pickBackendBearer } from "./session";

async function encoded(cookieName: string, overrides: Record<string, unknown> = {}) {
  return encode({
    salt: cookieName,
    secret: env.NEXTAUTH_SECRET,
    token: {
      name: "Signer Admin",
      email: "admin@example.org",
      accessToken: "opaque-access",
      idToken: "one.two.three",
      privileges: ["signer:certificates:read", "signer:certificates:write"],
      ...overrides,
    },
  });
}

beforeEach(() => {
  env.NODE_ENV = "test";
});

describe("shared portal sessions", () => {
  it("decodes the development cookie and selects a JWT-shaped bearer", async () => {
    const value = await encoded("custos.session-token");
    const session = await decodeSharedSession({
      headers: new Headers({ cookie: `custos.session-token=${value}` }),
    });
    expect(session?.privileges).toContain("signer:certificates:write");
    expect(pickBackendBearer(session)).toBe("one.two.three");
  });

  it("decodes a chunked production cookie", async () => {
    env.NODE_ENV = "production";
    const name = "__Secure-custos.session-token";
    const value = await encoded(name);
    const midpoint = Math.floor(value.length / 2);
    const cookie = `${name}.0=${value.slice(0, midpoint)}; ${name}.1=${value.slice(midpoint)}`;
    const session = await decodeSharedSession({ headers: new Headers({ cookie }) });
    expect(session?.email).toBe("admin@example.org");
  });

  it("returns null for an invalid cookie", async () => {
    await expect(
      decodeSharedSession({ headers: new Headers({ cookie: "custos.session-token=invalid" }) }),
    ).resolves.toBeNull();
  });
});
