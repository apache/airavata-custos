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

import { CHART_OTHER_COLOR, CHART_SERIES_COLORS } from "../lib";
import type { UsageSummary } from "../schemas";
import { BudgetDonut, type BudgetSlice } from "./BudgetDonut";

// Where the allocation's credit budget went, resource by resource: colored
// sectors are each resource's draw from the shared pool, the light track is
// what remains.
export function ResourceBreakdown({ summary }: { summary: UsageSummary }) {
  const ranked = [...summary.by_resource].sort((a, b) => b.used - a.used);
  const slices: BudgetSlice[] = ranked.map((r, i) => ({
    key: r.resource_id,
    label: r.name,
    value: r.used,
    color: CHART_SERIES_COLORS[i] ?? CHART_OTHER_COLOR,
  }));

  return (
    <div className="flex h-full flex-col rounded-lg border border-border bg-card p-4">
      <h2 className="text-sm font-semibold">Usage by resource</h2>
      <p className="mb-4 text-xs text-muted-foreground">
        how much of the allocation each resource has used
      </p>
      <BudgetDonut items={slices} total={summary.total} ariaContext="Usage by resource" />
    </div>
  );
}
