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

// Guards MSW fixtures against the schemas they're rendered with.

import { describe, expect, it } from "vitest";
import { z } from "zod";
import allocationsFixture from "@/features/core/allocations/__fixtures__/allocations.json";
import { allocationStatusSchema } from "@/features/core/allocations/schemas";
import membersFixture from "@/features/core/projects/__fixtures__/members.json";
import projectsFixture from "@/features/core/projects/__fixtures__/projects.json";
import { projectMemberSchema, projectSchema } from "@/features/core/projects/schemas";

describe("MSW fixtures pass current Zod schemas", () => {
  it("project fixtures validate against projectSchema (zProjectStatus-backed)", () => {
    const result = z.array(projectSchema).safeParse(projectsFixture);
    expect(result.success, JSON.stringify(result.error?.issues, null, 2)).toBe(true);
  });

  it("member fixtures validate against projectMemberSchema (zUserType-backed)", () => {
    const allMembers = Object.values(
      membersFixture as Record<string, unknown[]>,
    ).flat();
    const result = z.array(projectMemberSchema).safeParse(allMembers);
    expect(result.success, JSON.stringify(result.error?.issues, null, 2)).toBe(true);
  });

  it("allocation status fixture values validate against zAllocationStatus-backed enum", () => {
    const statuses = (allocationsFixture as Array<{ status: string }>).map((a) => a.status);
    const result = z.array(allocationStatusSchema).safeParse(statuses);
    expect(result.success, JSON.stringify(result.error?.issues, null, 2)).toBe(true);
  });
});
