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
import { describe, expect, it, vi } from "vitest";

vi.mock("@/shared/auth/auth", () => ({
  signIn: vi.fn(),
}));

import { signIn } from "@/shared/auth/auth";
import { GET } from "../route";

describe("federated-sign-in route", () => {
  it("starts the oidc flow with the requested callback", async () => {
    await GET(
      new NextRequest("http://localhost:3000/api/auth/federated-sign-in?callbackUrl=/projects"),
    );
    expect(signIn).toHaveBeenCalledWith("oidc", { redirectTo: "/projects" });
  });

  it("defaults the callback to the portal root", async () => {
    await GET(new NextRequest("http://localhost:3000/api/auth/federated-sign-in"));
    expect(signIn).toHaveBeenCalledWith("oidc", { redirectTo: "/" });
  });
});
