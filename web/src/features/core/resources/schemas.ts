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

import type { ComputeAllocationResourceRate } from "@/generated/core/zod.gen";
import { z } from "zod";

// A rate is active when now falls within [start_time, end_time).
export function isRateActive(rate: ComputeAllocationResourceRate, at: Date = new Date()): boolean {
  const start = new Date(rate.start_time).getTime();
  const end = new Date(rate.end_time).getTime();
  const now = at.getTime();
  return now >= start && now < end;
}

export type RateStatus = "ACTIVE" | "SCHEDULED" | "SUPERSEDED" | "EXPIRED";

// Effective = the containing window with the latest start_time, mirroring the
// backend ORDER BY start_time DESC. Overlaps resolve to SUPERSEDED, not errors.
export function classifyRate(
  rate: ComputeAllocationResourceRate,
  all: ComputeAllocationResourceRate[],
  now: number = Date.now(),
): RateStatus {
  const start = Date.parse(rate.start_time);
  const end = Date.parse(rate.end_time);
  if (start > now) return "SCHEDULED";
  if (end <= now) return "EXPIRED";
  const effectiveStart = Math.max(
    ...all
      .filter((r) => Date.parse(r.start_time) <= now && Date.parse(r.end_time) > now)
      .map((r) => Date.parse(r.start_time)),
  );
  return start === effectiveStart ? "ACTIVE" : "SUPERSEDED";
}

export const createRateSchema = z
  .object({
    rate: z.number({ message: "Rate is required" }).min(0, "Rate must be zero or greater"),
    start_date: z.string().min(1, "Start date is required"),
    end_date: z.string().min(1, "End date is required"),
  })
  .refine((d) => d.end_date > d.start_date, {
    path: ["end_date"],
    message: "End date must be after the start date",
  });
export type CreateRateForm = z.infer<typeof createRateSchema>;
