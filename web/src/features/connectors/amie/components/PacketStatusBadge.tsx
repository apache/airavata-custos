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
import type { PacketStatus, ReplyStatus } from "../types";

const packetStyles: Record<PacketStatus, string> = {
  NEW: "bg-[color:var(--custos-blue-50)] text-[color:var(--custos-blue-700)]",
  DECODED: "bg-[color:var(--custos-amber-50)] text-[color:var(--custos-amber-700)]",
  PROCESSED: "bg-[color:var(--custos-green-50)] text-[color:var(--custos-green-700)]",
  FAILED: "bg-[color:var(--custos-red-50)] text-[color:var(--custos-red-700)]",
};

const replyStyles: Record<ReplyStatus, string> = {
  PENDING: "bg-[color:var(--custos-amber-50)] text-[color:var(--custos-amber-700)]",
  SENT: "bg-[color:var(--custos-blue-50)] text-[color:var(--custos-blue-700)]",
  ACKED: "bg-[color:var(--custos-green-50)] text-[color:var(--custos-green-700)]",
  FAILED: "bg-[color:var(--custos-red-50)] text-[color:var(--custos-red-700)]",
};

export function PacketStatusBadge({
  status,
  ageHours,
  className,
}: {
  status: PacketStatus;
  ageHours?: number;
  className?: string;
}) {
  // Aged-FAILED packets stay louder via weight + trailing "!" — no ring needed.
  const loud = status === "FAILED" && ageHours !== undefined && ageHours > 24;
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium",
        packetStyles[status],
        loud && "font-semibold",
        className,
      )}
    >
      {status}
      {loud ? <span aria-hidden="true">!</span> : null}
    </span>
  );
}

export function ReplyStatusBadge({
  status,
  className,
}: {
  status: ReplyStatus;
  className?: string;
}) {
  return (
    <span
      className={cn(
        "inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium",
        replyStyles[status],
        className,
      )}
    >
      {status}
    </span>
  );
}
