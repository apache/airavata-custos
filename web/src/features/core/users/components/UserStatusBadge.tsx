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

import { StatusBadge, type StatusBadgeVariant } from "@/shared/ui/StatusBadge";

const USER_STATUSES: Record<string, { variant: StatusBadgeVariant; label: string }> = {
  ACTIVE: { variant: "active", label: "Active" },
  PENDING: { variant: "pending", label: "Awaiting sign-in" },
  SUSPENDED: { variant: "warning", label: "Suspended" },
  INACTIVE: { variant: "inactive", label: "Inactive" },
  MERGED: { variant: "inactive", label: "Merged" },
};

export const userStatusLabel = (status: string) => USER_STATUSES[status]?.label ?? status;

// PENDING means the account exists but nobody has signed in as it yet. The
// status flips on first sign-in, not when cluster provisioning finishes.
export function UserStatusBadge({ status }: { status: string | undefined }) {
  const known = status ? USER_STATUSES[status] : undefined;
  if (!known) return <span className="text-muted-foreground">—</span>;
  return <StatusBadge variant={known.variant} label={known.label} />;
}
