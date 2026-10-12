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

import type { Allocation, ProjectContext } from "@/generated/analytics/types.gen";
import {
  replaceShallowSearchParams,
  useCanonicalParams,
  useShallowSearchParams,
} from "@/shared/hooks/useShallowSearchParams";
import { normalizeQueryId } from "@/shared/layout/nav";
import * as React from "react";
import { pctRemaining } from "./lib";

// The lowest-remaining allocation across all contexts — the one most likely to
// need attention, used as the landing default when the URL says nothing.
function mostAtRisk(
  contexts: ProjectContext[],
): { project: ProjectContext; allocation: Allocation } | undefined {
  let best: { project: ProjectContext; allocation: Allocation; pct: number } | undefined;
  for (const c of contexts) {
    for (const a of c.allocations ?? []) {
      const pct = pctRemaining(a.initial_su_amount, a.used_su_amount);
      if (!best || pct < best.pct) best = { project: c, allocation: a, pct };
    }
  }
  return best;
}

// Resolves the selected project and allocation. An explicit allocation in the
// URL wins, then an explicit project (its first allocation), else the page lands
// on the most-at-risk allocation. Selection persists in the URL; once contexts load,
// a project or allocation outside the caller's contexts is dropped from it and
// `unavailable` says so, and the accepted ones are rewritten in normalised form.
export function useAnalyticsSelection(contexts: ProjectContext[], loaded: boolean) {
  const params = useShallowSearchParams();
  const projectParam = normalizeQueryId(params.get("project"));
  const allocationParam = normalizeQueryId(params.get("allocation"));
  const [unavailable, setUnavailable] = React.useState(false);
  const owner = contexts.find((c) => c.allocations?.some((a) => a.id === allocationParam));
  const knownProject = contexts.find((c) => c.project_id === projectParam);
  const missing = loaded && ((!!allocationParam && !owner) || (!!projectParam && !knownProject));

  React.useEffect(() => {
    if (missing) setUnavailable(true);
  }, [missing]);
  useCanonicalParams(
    loaded
      ? {
          project: projectParam && (owner ?? knownProject)?.project_id,
          allocation: owner ? allocationParam : undefined,
        }
      : {},
  );

  const { project, allocation } = React.useMemo<{
    project: ProjectContext | undefined;
    allocation: Allocation | undefined;
  }>(() => {
    const a = owner?.allocations?.find((x) => x.id === allocationParam);
    if (owner && a) return { project: owner, allocation: a };
    if (knownProject) return { project: knownProject, allocation: knownProject.allocations?.[0] };
    const best = mostAtRisk(contexts);
    if (best) return { project: best.project, allocation: best.allocation };
    return { project: contexts[0], allocation: contexts[0]?.allocations?.[0] };
  }, [contexts, owner, knownProject, allocationParam]);

  const select = React.useCallback((projectId: string, allocationId: string) => {
    setUnavailable(false);
    const next = new URLSearchParams(window.location.search);
    next.set("project", projectId);
    next.set("allocation", allocationId);
    replaceShallowSearchParams(next);
  }, []);

  return { project, allocation, select, unavailable };
}
