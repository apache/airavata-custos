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
import { allocationKeys } from "../queries";

describe("allocationKeys", () => {
  it("namespaces under 'allocations'", () => {
    expect(allocationKeys.all).toEqual(["allocations"]);
  });

  it("list key carries the params", () => {
    const params = { limit: 20, status: "ACTIVE" as const };
    expect(allocationKeys.list(params)).toEqual(["allocations", "list", params]);
  });

  it("detail key carries the id", () => {
    expect(allocationKeys.detail("alloc-1")).toEqual(["allocations", "detail", "alloc-1"]);
  });

  it("resources key extends detail", () => {
    expect(allocationKeys.resources("alloc-1")).toEqual([
      "allocations",
      "detail",
      "alloc-1",
      "resources",
    ]);
  });

  it("members key extends detail", () => {
    expect(allocationKeys.members("alloc-1")).toEqual([
      "allocations",
      "detail",
      "alloc-1",
      "members",
    ]);
  });

  it("change-requests list key carries params", () => {
    expect(allocationKeys.changeRequests({ status: "PENDING" })).toEqual([
      "allocations",
      "change-requests",
      "list",
      { status: "PENDING" },
    ]);
  });

  it("change-requests detail key carries the id", () => {
    expect(allocationKeys.changeRequestDetail("cr-1")).toEqual([
      "allocations",
      "change-requests",
      "detail",
      "cr-1",
    ]);
  });

  it("change-requests events key carries the id", () => {
    expect(allocationKeys.changeRequestEvents("cr-1")).toEqual([
      "allocations",
      "change-requests",
      "events",
      "cr-1",
    ]);
  });
});
