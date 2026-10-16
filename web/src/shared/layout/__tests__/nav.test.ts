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
import { normalizeQueryId, normalizeRouteId, pageMetadata } from "../nav";

describe("pageMetadata", () => {
  it("titles administration pages '<Page> · Admin' and the caller's own pages by brand", () => {
    expect(pageMetadata("/admin/amie").title).toBe("AMIE · Admin");
    expect(pageMetadata("/admin/users/roles").title).toBe("Role management · Admin");
    expect(pageMetadata("/change-requests").title).toBe("Change requests · Admin");
    expect(pageMetadata("/allocations").title).toBe("Allocations · Custos");
  });
});

describe("normalizeRouteId", () => {
  // Make sure a pasted id opens its record: decoded once, trimmed, hex ids lowercased, blank ignored.
  it("normalises pasted ids", () => {
    expect(normalizeRouteId("%20A1B2C3D4-0000-4000-8000-00000000ABCD%20")).toBe(
      "a1b2c3d4-0000-4000-8000-00000000abcd",
    );
    expect(normalizeRouteId("a%2Fb")).toBe("a/b");
    expect(normalizeRouteId("alloc-004")).toBe("alloc-004");
    expect(normalizeRouteId("  ")).toBeUndefined();
    expect(normalizeRouteId(null)).toBeUndefined();
    // A query value is already decoded, so it is not decoded again.
    expect(normalizeQueryId("%2Fx")).toBe("%2Fx");
    expect(normalizeQueryId(" A1B2C3D4-0000-4000-8000-00000000ABCD ")).toBe(
      "a1b2c3d4-0000-4000-8000-00000000abcd",
    );
    expect(normalizeRouteId("4BF92F3577B34DA6A3CE929D0E0E4736")).toBe(
      "4bf92f3577b34da6a3ce929d0e0e4736",
    );
  });
});
