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

import type { AllocationStatus as Status } from "@/generated/core/types.gen";
import { formatDay } from "@/shared/format";
import { statusLabel } from "@/shared/ui/FormDialog";
import { StatusBadge, statusBadgeVariantFromAllocationStatus } from "@/shared/ui/StatusBadge";

// Records the server created without a window carry Go's zero time, which reads as no time.
export const knownTime = (iso?: string) =>
  iso && new Date(iso).getUTCFullYear() >= 1000 ? iso : undefined;
export const knownDay = (iso?: string) => formatDay(knownTime(iso)) || "—";
// The calendar day after a date input's value, the earliest end of a window starting on it.
export const dayAfter = (day: string) =>
  day
    ? new Date(Date.parse(`${day}T00:00:00Z`) + 86_400_000).toISOString().slice(0, 10)
    : undefined;

export const membershipStatusText = (status: Status) => statusLabel(status);
export const membershipStatusVariant = statusBadgeVariantFromAllocationStatus;

// The stored status stays authoritative; a window not yet begun or already over is noted beside it.
export function AllocationStatusLabel({
  status,
  startTime,
  endTime,
}: {
  status: Status;
  startTime: string;
  endTime: string;
}) {
  const now = Date.now();
  const start = knownTime(startTime);
  const end = knownTime(endTime);
  const hint =
    start && Date.parse(start) > now
      ? `starts ${formatDay(start)}`
      : end && Date.parse(end) < now
        ? `ended ${formatDay(end)}`
        : null;
  return (
    <span className="inline-flex items-center gap-2">
      <StatusBadge variant={membershipStatusVariant(status)} label={membershipStatusText(status)} />
      {hint ? <span className="text-xs text-muted-foreground">{hint}</span> : null}
    </span>
  );
}
