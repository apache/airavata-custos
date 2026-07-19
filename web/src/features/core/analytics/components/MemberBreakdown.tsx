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
import type { UsageMember } from "../schemas";
import { BudgetDonut, type BudgetSlice } from "./BudgetDonut";

const TOP_MEMBERS = 5;

// Arcs thinner than ~0.4% of the ring are unreadable and unhoverable, so the
// tail beyond the top members shares one arc; the legend still lists everyone.
const MIN_SLICE_FRACTION = 0.004;

export function MemberBreakdown({ members, total }: { members: UsageMember[]; total: number }) {
  const ranked = [...members].sort((a, b) => b.used - a.used);
  const used = ranked.reduce((a, m) => a + m.used, 0);
  const minSlice = MIN_SLICE_FRACTION * Math.max(total, used);

  const items: BudgetSlice[] = ranked.map((m, i) => {
    const folded = i >= TOP_MEMBERS || m.used < minSlice;
    return {
      key: m.user_id,
      label: m.name,
      value: m.used,
      color: folded ? CHART_OTHER_COLOR : (CHART_SERIES_COLORS[i] ?? CHART_OTHER_COLOR),
      folded,
    };
  });

  return (
    <div className="flex h-full flex-col rounded-lg border border-border bg-card p-4">
      <h2 className="text-sm font-semibold">Usage by member</h2>
      <p className="mb-4 text-xs text-muted-foreground">
        how much of the allocation each member has used
      </p>
      <BudgetDonut items={items} total={total} ariaContext="Usage by member" />
    </div>
  );
}
