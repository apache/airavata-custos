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
import { computeAllocationResourceSchema, isRateActive, rateSchema, type Rate } from "../schemas";

describe("computeAllocationResourceSchema", () => {
  it("accepts a fully populated resource", () => {
    const parsed = computeAllocationResourceSchema.parse({
      id: "res-1",
      name: "ClusterA CPU",
      resource_type: "CPU",
      resource_amount: 1000,
      compute_cluster_id: "cluster-a",
    });
    expect(parsed.name).toBe("ClusterA CPU");
  });

  it("rejects a resource missing its cluster id", () => {
    const result = computeAllocationResourceSchema.safeParse({
      id: "res-1",
      name: "ClusterA CPU",
      resource_type: "CPU",
      resource_amount: 1000,
    });
    expect(result.success).toBe(false);
  });
});

describe("isRateActive", () => {
  const at = new Date("2026-06-01T00:00:00Z");
  const rate = (start: string, end: string): Rate =>
    rateSchema.parse({
      id: "r",
      compute_allocation_resource_id: "res-1",
      rate: 0.05,
      start_time: start,
      end_time: end,
    });

  it("is active when the reference time is within the window", () => {
    expect(isRateActive(rate("2020-01-01T00:00:00Z", "2035-01-01T00:00:00Z"), at)).toBe(true);
  });

  it("is inactive when the window has ended", () => {
    expect(isRateActive(rate("2018-01-01T00:00:00Z", "2020-01-01T00:00:00Z"), at)).toBe(false);
  });

  it("is inactive at the exact end (half-open interval)", () => {
    const end = new Date("2020-01-01T00:00:00Z");
    expect(isRateActive(rate("2018-01-01T00:00:00Z", "2020-01-01T00:00:00Z"), end)).toBe(false);
  });
});
