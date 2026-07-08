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

import type { z } from "zod";
import {
  zComputeAllocationResource,
  zComputeAllocationResourceRate,
} from "@/generated/core/zod.gen";

// The generated schema marks every field optional; the backend always returns
// these, so tighten them here where the UI reads them.
export const computeAllocationResourceSchema = zComputeAllocationResource.required({
  id: true,
  name: true,
  resource_type: true,
  resource_amount: true,
  compute_cluster_id: true,
});
export type ComputeAllocationResource = z.infer<typeof computeAllocationResourceSchema>;

export const rateSchema = zComputeAllocationResourceRate.required({
  id: true,
  compute_allocation_resource_id: true,
  rate: true,
  start_time: true,
  end_time: true,
});
export type Rate = z.infer<typeof rateSchema>;

// A rate is active when now falls within [start_time, end_time).
export function isRateActive(rate: Rate, at: Date = new Date()): boolean {
  const start = new Date(rate.start_time).getTime();
  const end = new Date(rate.end_time).getTime();
  const now = at.getTime();
  return now >= start && now < end;
}
