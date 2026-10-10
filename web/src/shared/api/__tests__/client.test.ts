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

import { getProjectsById } from "@/generated/core/sdk.gen";
import type { ProjectResponse } from "@/generated/core/types.gen";
import { ApiError } from "@/shared/api/client";
import { http, HttpResponse } from "msw";
import { setupServer } from "msw/node";
import { afterAll, afterEach, beforeAll, describe, expect, it } from "vitest";

const project: ProjectResponse = {
  id: "p1",
  originated_id: "",
  title: "P",
  origination: "internal",
  project_pi_id: "u1",
  status: "ACTIVE",
  created_time: "2026-01-01T00:00:00Z",
};

const server = setupServer(
  http.get("*/api/v1/projects/traced", () =>
    HttpResponse.json(project),
  ),
  http.get("*/api/v1/projects/locked", () =>
    HttpResponse.json({ error: "project is managed externally" }, { status: 409 }),
  ),
);

beforeAll(() => server.listen({ onUnhandledRequest: "error" }));
afterAll(() => server.close());
afterEach(() => server.resetHandlers());

describe("generated SDK client", () => {
  it("prefixes /api/v1, returns the body", async () => {
    expect(await getProjectsById({ path: { id: "traced" } })).toEqual(project);
  });

  it("throws ApiError carrying the body's error text", async () => {
    const error = await getProjectsById({ path: { id: "locked" } }).catch((e) => e);
    expect(error).toBeInstanceOf(ApiError);
    expect(error).toMatchObject({ status: 409, message: "project is managed externally" });
  });
});
