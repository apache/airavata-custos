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

import { cn } from "@/lib/utils";

export type BurnDownBarProps = {
  used: number;
  projected: number;
  capacity: number;
  className?: string;
  ariaLabel?: string;
};

// Visualizes the sam-queries pace pattern: solid past-burn segment and a hatched
// projected-future segment so the viewer can compare against capacity at a glance.
export function BurnDownBar({ used, projected, capacity, className, ariaLabel }: BurnDownBarProps) {
  const safeCapacity = capacity <= 0 ? 1 : capacity;
  const usedPct = Math.min(100, Math.max(0, (used / safeCapacity) * 100));
  const projectedPct = Math.min(
    100 - usedPct,
    Math.max(0, ((projected - used) / safeCapacity) * 100),
  );
  return (
    <div
      className={cn("w-full", className)}
      role="img"
      aria-label={ariaLabel ?? `Burn-down: ${used} used of ${capacity}, projected ${projected}`}
    >
      <div className="flex h-3 w-full overflow-hidden rounded-full bg-muted">
        <div className="bg-brand" style={{ width: `${usedPct}%` }} />
        {/* Lighter brand-blue hatch for projected-future segment — no semantic token
            exists for "brand tint that reads as darker than --brand-tint". */}
        <div
          className="bg-[color:var(--custos-blue-300)] bg-[image:repeating-linear-gradient(45deg,transparent_0_4px,rgba(255,255,255,0.45)_4px_8px)]"
          style={{ width: `${projectedPct}%` }}
        />
      </div>
    </div>
  );
}
