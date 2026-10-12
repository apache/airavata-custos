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
import { enumLabel } from "@/shared/format";

const packetStyles: Record<string, string> = {
  NEW: "bg-[color:var(--tone-info-bg)] text-[color:var(--tone-info-fg)]",
  DECODED: "bg-[color:var(--tone-warn-bg)] text-[color:var(--tone-warn-fg)]",
  PROCESSED: "bg-[color:var(--tone-ok-bg)] text-[color:var(--tone-ok-fg)]",
  FAILED: "bg-[color:var(--tone-error-bg)] text-[color:var(--tone-error-fg)]",
  WAITING_APPROVAL: "bg-[color:var(--tone-accent-bg)] text-[color:var(--tone-accent-fg)]",
  // A refusal is an admin's decision, not an error, so it stays neutral.
  REFUSED: "bg-muted text-muted-foreground",
};

export function PacketStatusBadge({
  status = "",
  ageHours,
}: { status: string | undefined; ageHours: number }) {
  // Aged-FAILED packets stay louder via weight + trailing "!" — no ring needed.
  const loud = status === "FAILED" && ageHours > 24;
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium",
        packetStyles[status] ?? "bg-muted text-muted-foreground",
        loud && "font-semibold",
      )}
    >
      {enumLabel(status)}
      {loud ? <span aria-hidden="true">!</span> : null}
    </span>
  );
}
