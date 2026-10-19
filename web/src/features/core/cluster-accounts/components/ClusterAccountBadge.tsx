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

import { HardDrive } from "lucide-react";
import type { ClusterAccountResponse } from "@/generated/core/zod.gen";
import { Button } from "@/shared/ui/button";
import { StatusIcon } from "@/shared/ui/StatusIcon";
import type { StatusBadgeVariant } from "@/shared/ui/StatusBadge";

export function clusterAccountStatus(
  row: Pick<ClusterAccountResponse, "approval_status" | "provisioned_at">,
): {
  variant: StatusBadgeVariant;
  label: string;
} {
  if (row.approval_status === "DENIED") return { variant: "rejected", label: "Denied" };
  if (row.approval_status === "APPROVED") {
    return row.provisioned_at
      ? { variant: "active", label: "Active" }
      : { variant: "pending", label: "Approved · provisioning" };
  }
  return { variant: "warning", label: "Pending approval" };
}

// The review note, such as why an account was denied, is shown nowhere else.
export const clusterAccountLabel = (account: ClusterAccountResponse | undefined) =>
  account
    ? [clusterAccountStatus(account).label, account.review_note].filter(Boolean).join(": ")
    : "No account";

export function ClusterAccountBadge({
  name,
  account,
}: {
  name: string;
  account: ClusterAccountResponse | undefined;
}) {
  const active = account ? clusterAccountStatus(account).variant === "active" : false;
  return (
    <span className="inline-flex items-center gap-1.5 text-sm">
      <StatusIcon icon={HardDrive} active={active} label={clusterAccountLabel(account)} />
      {name}
    </span>
  );
}

// A pending account can be approved or denied, a denied one approved later.
export function ClusterAccountDecisions({
  account,
  onApprove,
  onDeny,
}: {
  account: ClusterAccountResponse;
  onApprove: ((account: ClusterAccountResponse) => void) | undefined;
  onDeny: ((account: ClusterAccountResponse) => void) | undefined;
}) {
  return (
    <span className="inline-flex gap-1">
      {onDeny && account.approval_status === "PENDING" ? (
        <Button size="xs" variant="destructive" onClick={() => onDeny(account)}>
          Deny
        </Button>
      ) : null}
      {onApprove && account.approval_status !== "APPROVED" ? (
        <Button size="xs" variant="success" onClick={() => onApprove(account)}>
          Approve
        </Button>
      ) : null}
    </span>
  );
}
