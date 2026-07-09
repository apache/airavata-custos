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
import allocationResourcesFixture from "@/features/core/allocations/__fixtures__/resources.json";
import allocationDiffsFixture from "@/features/core/allocations/__fixtures__/diffs.json";
import allocationUsageFixture from "@/features/core/allocations/__fixtures__/usage.json";
import {
  allocationDiffSchema,
  attachedResourceSchema,
  allocationStatusSchema,
  allocationUsageSchema,
} from "@/features/core/allocations/schemas";
import clusterUsersFixture from "@/features/core/clusters/__fixtures__/cluster-users.json";
import clustersFixture from "@/features/core/clusters/__fixtures__/clusters.json";
import {
  computeClusterSchema,
  computeClusterUserSchema,
} from "@/features/core/clusters/schemas";
import organizationsFixture from "@/features/core/organizations/__fixtures__/organizations.json";
import { organizationSchema } from "@/features/core/organizations/schemas";
import ratesFixture from "@/features/core/resources/__fixtures__/rates.json";
import resourcesFixture from "@/features/core/resources/__fixtures__/resources.json";
import {
  computeAllocationResourceSchema,
  rateSchema,
} from "@/features/core/resources/schemas";
import membersFixture from "@/features/core/projects/__fixtures__/members.json";
import projectsFixture from "@/features/core/projects/__fixtures__/projects.json";
import { projectMemberSchema, projectSchema } from "@/features/core/projects/schemas";
import settingsFixture from "@/features/core/identity/__fixtures__/settings.json";
import {
  roleDetailResponseSchema,
  userIdentitySchema,
  userPrivilegeSchema,
  userRoleSchema,
  userSchema,
} from "@/features/core/identity/schemas";

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

  it("organization fixtures validate against organizationSchema", () => {
    const result = z.array(organizationSchema).safeParse(organizationsFixture);
    expect(result.success, JSON.stringify(result.error?.issues, null, 2)).toBe(true);
  });

  it("allocation status fixture values validate against zAllocationStatus-backed enum", () => {
    const statuses = (allocationsFixture as Array<{ status: string }>).map((a) => a.status);
    const result = z.array(allocationStatusSchema).safeParse(statuses);
    expect(result.success, JSON.stringify(result.error?.issues, null, 2)).toBe(true);
  });

  it("allocation resource fixtures validate against attachedResourceSchema", () => {
    const rows = Object.values(
      allocationResourcesFixture as Record<string, unknown[]>,
    ).flat();
    const result = z.array(attachedResourceSchema).safeParse(rows);
    expect(result.success, JSON.stringify(result.error?.issues, null, 2)).toBe(true);
  });

  it("allocation diff fixtures validate against allocationDiffSchema", () => {
    const rows = Object.values(allocationDiffsFixture as Record<string, unknown[]>).flat();
    const result = z.array(allocationDiffSchema).safeParse(rows);
    expect(result.success, JSON.stringify(result.error?.issues, null, 2)).toBe(true);
  });

  it("allocation usage fixtures validate against allocationUsageSchema", () => {
    const rows = Object.values(allocationUsageFixture as Record<string, unknown[]>).flat();
    const result = z.array(allocationUsageSchema).safeParse(rows);
    expect(result.success, JSON.stringify(result.error?.issues, null, 2)).toBe(true);
  });

  it("cluster fixtures validate against computeClusterSchema", () => {
    const result = z.array(computeClusterSchema).safeParse(clustersFixture);
    expect(result.success, JSON.stringify(result.error?.issues, null, 2)).toBe(true);
  });

  it("cluster user fixtures validate against computeClusterUserSchema", () => {
    const result = z.array(computeClusterUserSchema).safeParse(clusterUsersFixture);
    expect(result.success, JSON.stringify(result.error?.issues, null, 2)).toBe(true);
  });

  it("resource fixtures validate against computeAllocationResourceSchema", () => {
    const result = z.array(computeAllocationResourceSchema).safeParse(resourcesFixture);
    expect(result.success, JSON.stringify(result.error?.issues, null, 2)).toBe(true);
  });

  it("rate fixtures validate against rateSchema", () => {
    const result = z.array(rateSchema).safeParse(ratesFixture);
    expect(result.success, JSON.stringify(result.error?.issues, null, 2)).toBe(true);
  });

  it("settings /me user validates against userSchema", () => {
    const result = userSchema.safeParse(settingsFixture.user);
    expect(result.success, JSON.stringify(result.error?.issues, null, 2)).toBe(true);
  });

  it("settings identity fixtures validate against userIdentitySchema", () => {
    const result = z.array(userIdentitySchema).safeParse(settingsFixture.identities);
    expect(result.success, JSON.stringify(result.error?.issues, null, 2)).toBe(true);
  });

  it("settings role grants validate against userRoleSchema", () => {
    const result = z.array(userRoleSchema).safeParse(settingsFixture.roles);
    expect(result.success, JSON.stringify(result.error?.issues, null, 2)).toBe(true);
  });

  it("settings role details validate against roleDetailResponseSchema", () => {
    const result = z
      .array(roleDetailResponseSchema)
      .safeParse(Object.values(settingsFixture.roleDetails));
    expect(result.success, JSON.stringify(result.error?.issues, null, 2)).toBe(true);
  });

  it("settings direct grants validate against userPrivilegeSchema", () => {
    const result = z.array(userPrivilegeSchema).safeParse(settingsFixture.direct);
    expect(result.success, JSON.stringify(result.error?.issues, null, 2)).toBe(true);
  });

  it("settings role grant ids resolve via roleDetails (internally consistent)", () => {
    const detailIds = Object.keys(settingsFixture.roleDetails);
    for (const grant of settingsFixture.roles) {
      expect(detailIds).toContain(grant.role_id);
    }
  });
});
