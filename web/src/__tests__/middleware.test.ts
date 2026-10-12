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

import { NextRequest } from "next/server";
import { describe, expect, it } from "vitest";
import { middleware } from "../middleware";

function request(cookie?: string, path = "/") {
  const headers = cookie ? { cookie } : undefined;
  return new NextRequest(`http://localhost:3000${path}`, { headers });
}

describe("root middleware", () => {
  it("rewrites signed-out users to the landing page", () => {
    const res = middleware(request());
    expect(res.headers.get("x-middleware-rewrite")).toContain("/landing/index.html");
  });

  it("passes signed-in users through to the portal", () => {
    const res = middleware(request("custos.session-token=abc"));
    expect(res.headers.get("x-middleware-rewrite")).toBeNull();
  });

  it("recognizes the production cookie name", () => {
    const res = middleware(request("__Secure-custos.session-token=abc"));
    expect(res.headers.get("x-middleware-rewrite")).toBeNull();
  });

  it("recognizes chunked session cookies", () => {
    const res = middleware(request("__Secure-custos.session-token.0=abc"));
    expect(res.headers.get("x-middleware-rewrite")).toBeNull();
  });

  it("redirects a padded or uppercase record id to its canonical path", () => {
    const res = middleware(
      request(
        "custos.session-token=abc",
        "/projects/%20AAAAAAAA-0000-4000-8000-000000000001?tab=members",
      ),
    );
    expect(res.status).toBe(308);
    expect(res.headers.get("location")).toBe(
      "http://localhost:3000/projects/aaaaaaaa-0000-4000-8000-000000000001?tab=members",
    );
  });
});
