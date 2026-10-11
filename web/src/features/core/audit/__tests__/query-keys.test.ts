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

import { describe, expect, it } from "vitest";
import { traceKeys } from "@/features/core/audit/queries";

describe("traceKeys shape", () => {
  it("exposes a stable hierarchy", () => {
    expect(traceKeys.all).toEqual(["traces"]);
    expect(traceKeys.lists()).toEqual(["traces", "list"]);
    expect(traceKeys.details()).toEqual(["traces", "detail"]);
  });

  it("nests list filters under the list root", () => {
    const filters = { source: ["amie"], status: [1], limit: 50 };
    expect(traceKeys.list(filters)).toEqual(["traces", "list", filters]);
  });

  it("nests detail ids under the detail root", () => {
    expect(traceKeys.detail("a3b1c92d3f4e5a6b7c8d9e0f12345678")).toEqual([
      "traces",
      "detail",
      "a3b1c92d3f4e5a6b7c8d9e0f12345678",
    ]);
  });

  it("namespaces sources and audit lookups distinctly", () => {
    expect(traceKeys.sources()).toEqual(["traces", "sources"]);
    expect(traceKeys.audit("a3b1c92d3f4e5a6b7c8d9e0f12345678", "1000000000000006")).toEqual([
      "traces",
      "audit",
      "a3b1c92d3f4e5a6b7c8d9e0f12345678",
      "1000000000000006",
    ]);
    expect(traceKeys.audit("a3b1c92d3f4e5a6b7c8d9e0f12345678")).toEqual([
      "traces",
      "audit",
      "a3b1c92d3f4e5a6b7c8d9e0f12345678",
      null,
    ]);
  });
});
