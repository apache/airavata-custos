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

import { Calendar, Gauge, Server, UserSquare } from "lucide-react";
import { MetaItem, MetaRow } from "@/shared/ui/MetaRow";
import type { ComputeAllocation } from "../schemas";

export type AllocationDetailHeaderProps = {
  allocation: ComputeAllocation;
  memberCount: number;
};

function formatDate(iso: string): string {
  try {
    return new Date(iso).toLocaleDateString(undefined, {
      year: "numeric",
      month: "short",
      day: "numeric",
    });
  } catch {
    return iso;
  }
}

function formatNumber(n: number): string {
  return new Intl.NumberFormat().format(n);
}

function isExpired(endTime: string, now = new Date()): boolean {
  const t = new Date(endTime).getTime();
  return Number.isFinite(t) && t < now.getTime();
}

export function AllocationDetailHeader({ allocation, memberCount }: AllocationDetailHeaderProps) {
  const expired = allocation.status === "ACTIVE" && isExpired(allocation.end_time);
  const tone: "success" | "warning" | "danger" =
    allocation.status === "ACTIVE" && !expired
      ? "success"
      : allocation.status === "ACTIVE" && expired
        ? "warning"
        : "danger";
  const label = expired ? "EXPIRED" : allocation.status;

  return (
    <header className="space-y-4">
      <h1 className="font-display text-[28px] font-bold leading-tight text-foreground">
        {allocation.name}
      </h1>
      <MetaItem
        className="hidden"
        icon={Server}
        label="Cluster"
        value={allocation.compute_cluster_id}
      />
      <div className="flex flex-wrap items-center gap-3">
        <MetaItem variant="status" tone={tone} value={label} className="py-1.5" />
        <MetaRow>
          <MetaItem
            icon={Gauge}
            label="Initial SUs"
            value={formatNumber(allocation.initial_su_amount)}
          />
          <MetaItem icon={Calendar} label="End date" value={formatDate(allocation.end_time)} />
          <MetaItem icon={UserSquare} label="Members" value={memberCount} />
        </MetaRow>
      </div>
    </header>
  );
}
