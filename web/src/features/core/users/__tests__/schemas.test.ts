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
import {
  roleDetailResponseSchema,
  userListResponseSchema,
  userRolesResponseSchema,
} from "../schemas";

const user = {
  id: "user-1",
  email: "user@example.org",
  first_name: "Example",
  status: "ACTIVE",
};

describe("user schemas", () => {
  it("parses a complete paginated user response", () => {
    expect(userListResponseSchema.parse({ items: [user], total: 1 })).toEqual({
      items: [user],
      total: 1,
    });
  });

  it("rejects missing pagination totals", () => {
    expect(() => userListResponseSchema.parse({ items: [user] })).toThrow();
  });

  it("normalizes null role lists to an empty array", () => {
    expect(userRolesResponseSchema.parse(null)).toEqual([]);
  });

  it("preserves arbitrary privilege keys returned by role detail", () => {
    const result = roleDetailResponseSchema.parse({
      role: { id: "role-1", name: "Admin" },
      privileges: ["core:roles:manage", "connector:custom:inspect"],
    });
    expect(result.privileges).toEqual(["core:roles:manage", "connector:custom:inspect"]);
  });
});
