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
  callerPrivilegesSchema,
  privilegeKeySchema,
  privilegesResponseSchema,
} from "../schemas";

describe("identity schemas", () => {
  it("re-exports the generated PrivilegeKey enum", () => {
    expect(() => privilegeKeySchema.parse("hpc:read")).not.toThrow();
    expect(() => privilegeKeySchema.parse("not-a-privilege")).toThrow();
  });

  it("accepts the OpenAPI CallerPrivileges shape", () => {
    const parsed = callerPrivilegesSchema.parse({ privileges: ["amie:read"] });
    expect(parsed.privileges).toEqual(["amie:read"]);
  });

  it("unwraps the response transform to a Privilege[] (default [])", () => {
    expect(privilegesResponseSchema.parse({ privileges: ["hpc:read"] })).toEqual(["hpc:read"]);
    expect(privilegesResponseSchema.parse({})).toEqual([]);
  });
});
