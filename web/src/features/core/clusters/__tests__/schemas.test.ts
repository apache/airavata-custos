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
import { computeClusterSchema, computeClusterUserSchema } from "../schemas";

describe("computeClusterSchema", () => {
  it("accepts a cluster with id and name", () => {
    expect(computeClusterSchema.parse({ id: "cluster-a", name: "ClusterA" }).name).toBe("ClusterA");
  });

  it("rejects a cluster missing its name", () => {
    expect(computeClusterSchema.safeParse({ id: "cluster-a" }).success).toBe(false);
  });
});

describe("computeClusterUserSchema", () => {
  it("accepts a fully populated cluster user", () => {
    const parsed = computeClusterUserSchema.parse({
      id: "ccu-1",
      compute_cluster_id: "cluster-a",
      user_id: "user-ada-lovelace",
      local_username: "alovelace",
    });
    expect(parsed.local_username).toBe("alovelace");
  });

  it("rejects a cluster user missing its local username", () => {
    const result = computeClusterUserSchema.safeParse({
      id: "ccu-1",
      compute_cluster_id: "cluster-a",
      user_id: "user-ada-lovelace",
    });
    expect(result.success).toBe(false);
  });
});
