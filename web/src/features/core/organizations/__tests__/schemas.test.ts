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
  createOrganizationPayloadSchema,
  organizationListEnvelopeSchema,
  organizationSchema,
} from "../schemas";

const validOrg = {
  id: "org-gatech",
  name: "Georgia Institute of Technology",
  originated_id: "GATECH",
};

describe("organizationSchema", () => {
  it("accepts a backend-shaped organization", () => {
    expect(organizationSchema.safeParse(validOrg).success).toBe(true);
  });
  it("rejects a missing name", () => {
    expect(organizationSchema.safeParse({ id: "org-1", originated_id: "X" }).success).toBe(false);
  });
});

describe("organizationListEnvelopeSchema", () => {
  it("accepts a valid envelope", () => {
    expect(
      organizationListEnvelopeSchema.safeParse({ items: [validOrg], total: 1 }).success,
    ).toBe(true);
  });
  it("rejects a negative total", () => {
    expect(organizationListEnvelopeSchema.safeParse({ items: [], total: -1 }).success).toBe(false);
  });
});

describe("createOrganizationPayloadSchema", () => {
  it("requires name 1+ chars", () => {
    expect(createOrganizationPayloadSchema.safeParse({ name: "" }).success).toBe(false);
  });
  it("accepts a name-only payload", () => {
    expect(createOrganizationPayloadSchema.safeParse({ name: "GT" }).success).toBe(true);
  });
  it("accepts an optional originated_id", () => {
    expect(
      createOrganizationPayloadSchema.safeParse({ name: "GT", originated_id: "GATECH" }).success,
    ).toBe(true);
  });
});
