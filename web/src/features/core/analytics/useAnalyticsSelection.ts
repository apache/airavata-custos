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

"use client";

import * as React from "react";
import {
  replaceShallowSearchParams,
  useShallowSearchParams,
} from "@/shared/hooks/useShallowSearchParams";
import { pctRemaining } from "./lib";
import type { AnalyticsAllocation, AnalyticsContext } from "./schemas";

// The lowest-remaining allocation across all contexts — the one most likely to
// need attention, used as the landing default when the URL says nothing.
function mostAtRisk(
  contexts: AnalyticsContext[],
): { project: AnalyticsContext; allocation: AnalyticsAllocation } | undefined {
  let best: { project: AnalyticsContext; allocation: AnalyticsAllocation; pct: number } | undefined;
  for (const c of contexts) {
    for (const a of c.allocations) {
      const pct = pctRemaining(a.initial_su_amount, a.used_su_amount);
      if (!best || pct < best.pct) best = { project: c, allocation: a, pct };
    }
  }
  return best;
}

// Resolves the selected project and allocation. An explicit allocation in the
// URL wins, then an explicit project (its first allocation), else the page lands
// on the most-at-risk allocation. Selection persists in the URL.
export function useAnalyticsSelection(contexts: AnalyticsContext[]) {
  const params = useShallowSearchParams();
  const projectParam = params.get("project");
  const allocationParam = params.get("allocation");

  const { project, allocation } = React.useMemo<{
    project: AnalyticsContext | undefined;
    allocation: AnalyticsAllocation | undefined;
  }>(() => {
    if (allocationParam) {
      for (const c of contexts) {
        const a = c.allocations.find((x) => x.id === allocationParam);
        if (a) return { project: c, allocation: a };
      }
    }
    if (projectParam) {
      const c = contexts.find((x) => x.project_id === projectParam);
      if (c) return { project: c, allocation: c.allocations[0] };
    }
    const best = mostAtRisk(contexts);
    if (best) return { project: best.project, allocation: best.allocation };
    return { project: contexts[0], allocation: contexts[0]?.allocations[0] };
  }, [contexts, projectParam, allocationParam]);

  const select = React.useCallback((projectId: string, allocationId: string) => {
    const next = new URLSearchParams(window.location.search);
    next.set("project", projectId);
    next.set("allocation", allocationId);
    replaceShallowSearchParams(next);
  }, []);

  return { project, allocation, select };
}
