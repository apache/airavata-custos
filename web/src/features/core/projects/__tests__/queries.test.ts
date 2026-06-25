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
import { projectKeys } from "../queries";

describe("projectKeys", () => {
  it("namespaces under 'projects'", () => {
    expect(projectKeys.all).toEqual(["projects"]);
  });

  it("list key carries the params", () => {
    const params = { limit: 20, status: "ACTIVE" as const };
    expect(projectKeys.list(params)).toEqual(["projects", "list", params]);
  });

  it("detail key carries the id", () => {
    expect(projectKeys.detail("p-1")).toEqual(["projects", "detail", "p-1"]);
  });

  it("members key carries the project id", () => {
    expect(projectKeys.members("p-1")).toEqual(["projects", "members", "p-1"]);
  });
});
