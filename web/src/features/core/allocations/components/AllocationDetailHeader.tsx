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

import type { ComputeAllocation } from "@/generated/core/types.gen";
import { formatSU } from "@/shared/format";
import { MetaItem, MetaRow } from "@/shared/ui/MetaRow";
import { Calendar, Gauge, UserSquare } from "lucide-react";
import { AllocationStatusLabel, knownDay } from "./AllocationStatusLabel";

export type AllocationDetailHeaderProps = {
  allocation: ComputeAllocation;
  memberCount: number | undefined;
};

export function AllocationDetailHeader({ allocation, memberCount }: AllocationDetailHeaderProps) {
  return (
    <header className="space-y-4">
      <h1 className="font-display text-[28px] font-bold leading-tight text-foreground [overflow-wrap:anywhere]">
        {allocation.name}
      </h1>
      <div className="flex flex-wrap items-center gap-3">
        <AllocationStatusLabel
          status={allocation.status}
          startTime={allocation.start_time}
          endTime={allocation.end_time}
        />
        <MetaRow>
          <MetaItem
            icon={Gauge}
            label="Initial SUs"
            value={formatSU(allocation.initial_su_amount)}
          />
          <MetaItem icon={Calendar} label="End date" value={knownDay(allocation.end_time)} />
          <MetaItem icon={UserSquare} label="Members" value={memberCount ?? "—"} />
        </MetaRow>
      </div>
    </header>
  );
}
