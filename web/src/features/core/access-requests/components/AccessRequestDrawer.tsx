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

import type * as React from "react";
import { useAllocation } from "@/features/core/allocations/queries";
import { formatCreditsFull } from "@/features/core/analytics/lib";
import { SideDrawer } from "@/shared/ui/SideDrawer";
import type { AccessRequest } from "../schemas";
import { formatDate, statusBadgeFor } from "./AccessRequestsQueue";

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="grid grid-cols-[140px_1fr] gap-2 text-sm">
      <dt className="text-muted-foreground">{label}</dt>
      <dd className="text-foreground">{children}</dd>
    </div>
  );
}

function GrantedAllocation({ allocationId }: { allocationId: string }) {
  const { data: allocation } = useAllocation(allocationId);
  if (!allocation) return <span className="text-muted-foreground">—</span>;
  return (
    <span>
      {allocation.name}
      <span className="text-muted-foreground">
        {" "}
        · {formatCreditsFull(allocation.initial_su_amount)} credits
      </span>
    </span>
  );
}

export function AccessRequestDrawer({
  request,
  onClose,
}: {
  request: AccessRequest | null;
  onClose: () => void;
}) {
  const decided = request !== null && request.status !== "PENDING";
  return (
    <SideDrawer
      open={request !== null}
      onOpenChange={(open) => {
        if (!open) onClose();
      }}
      title={request?.name}
      description={request?.email}
      width="sm"
      modal={false}
      disablePointerDismissal
    >
      {request && (
        <div className="space-y-5">
          <section>
            <h3 className="mb-3 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
              Request
            </h3>
            <dl className="space-y-2">
              <Field label="Institution">{request.institution}</Field>
              <Field label="Event">
                <span className="font-mono text-xs">{request.event_code}</span>
              </Field>
              <Field label="Reason">{request.reason || "—"}</Field>
              <Field label="Submitted">{formatDate(request.timestamp)}</Field>
            </dl>
          </section>

          {decided ? (
            <>
              <div className="border-t border-border" />
              <section>
                <h3 className="mb-3 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                  Decision
                </h3>
                <dl className="space-y-2">
                  <Field label="Status">{statusBadgeFor(request.status)}</Field>
                  <Field label="Decided by">
                    <span className="font-mono text-xs" title={request.approver_id}>
                      {request.approver_id ? `${request.approver_id.slice(0, 8)}…` : "—"}
                    </span>
                  </Field>
                  <Field label="Decided at">
                    {request.decided_at ? formatDate(request.decided_at) : "—"}
                  </Field>
                  {request.status === "APPROVED" ? (
                    <>
                      <Field label="Granted allocation">
                        {request.allocation_id ? (
                          <GrantedAllocation allocationId={request.allocation_id} />
                        ) : (
                          "—"
                        )}
                      </Field>
                      <Field label="Allocation ends">
                        {request.expires_at ? formatDate(request.expires_at) : "—"}
                      </Field>
                    </>
                  ) : (
                    <Field label="Deny reason">{request.deny_reason || "—"}</Field>
                  )}
                </dl>
              </section>
            </>
          ) : null}
        </div>
      )}
    </SideDrawer>
  );
}
