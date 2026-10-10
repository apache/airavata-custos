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

import Link from "next/link";
import type * as React from "react";
import { AllocationName } from "@/features/core/allocations/components/AllocationName";
import { useClusterName } from "@/features/core/clusters/queries";
import {
  useUserChangeRequests,
  useUserClusterAccounts,
  useUserMemberships,
  useUserUsages,
} from "@/features/core/users/queries";
import { formatDate, formatNumber } from "@/shared/format";
import { DrawerSection } from "./DrawerSection";

const linkClass = "text-brand underline-offset-4 hover:underline";

// Read-only views of what a user holds elsewhere; each links to the page that edits it.
export function UserActivitySections({
  userId,
  canReadClusters,
  canReadAllocations,
}: {
  userId: string;
  canReadClusters: boolean;
  canReadAllocations: boolean;
}) {
  return (
    <>
      {canReadClusters ? <ClusterAccountsSection userId={userId} /> : null}
      {canReadAllocations ? (
        <>
          <MembershipsSection userId={userId} />
          <UsageSection userId={userId} />
          <ChangeRequestsSection userId={userId} />
        </>
      ) : null}
    </>
  );
}

function Section({
  title,
  query,
  empty,
  children,
}: {
  title: string;
  query: { isLoading: boolean; isError: boolean; data?: unknown[] };
  empty: string;
  children: React.ReactNode;
}) {
  return (
    <DrawerSection
      title={title}
      isLoading={query.isLoading}
      isError={query.isError}
      empty={(query.data ?? []).length === 0 && empty}
    >
      <ul className="space-y-1.5 text-sm">{children}</ul>
    </DrawerSection>
  );
}

function ClusterAccountsSection({ userId }: { userId: string }) {
  const query = useUserClusterAccounts(userId);
  const clusterName = useClusterName();
  return (
    <Section title="Cluster Accounts" query={query} empty="No cluster accounts.">
      {(query.data ?? []).map((account) => (
        <li key={account.id} className="flex items-center justify-between gap-2">
          <span>
            <span className="font-mono">{account.local_username}</span>
            <span className="text-muted-foreground">
              {" "}
              on {clusterName(account.compute_cluster_id)}
            </span>
          </span>
          <span className="text-xs text-muted-foreground">
            {account.access_level} · {account.approval_status}
          </span>
        </li>
      ))}
    </Section>
  );
}

function MembershipsSection({ userId }: { userId: string }) {
  const query = useUserMemberships(userId);
  return (
    <Section title="Allocation Memberships" query={query} empty="No allocation memberships.">
      {(query.data ?? []).map((m) => (
        <li key={m.id} className="flex items-center justify-between gap-2">
          <AllocationName id={m.compute_allocation_id} className={linkClass} />
          <span className="text-xs text-muted-foreground">{m.membership_status}</span>
        </li>
      ))}
    </Section>
  );
}

function UsageSection({ userId }: { userId: string }) {
  const query = useUserUsages(userId);
  const totals = new Map<string, number>();
  for (const usage of query.data ?? []) {
    const id = usage.compute_allocation_id ?? "";
    totals.set(id, (totals.get(id) ?? 0) + (usage.used_su_amount ?? 0));
  }
  return (
    <Section title="Usage" query={query} empty="No recorded usage.">
      {[...totals].map(([allocationId, su]) => (
        <li key={allocationId} className="flex items-center justify-between gap-2">
          <AllocationName id={allocationId} className={linkClass} />
          <span className="tabular-nums text-muted-foreground">{formatNumber(su)} SU</span>
        </li>
      ))}
    </Section>
  );
}

function ChangeRequestsSection({ userId }: { userId: string }) {
  const query = useUserChangeRequests(userId);
  return (
    <Section title="Submitted Change Requests" query={query} empty="No change requests.">
      {(query.data ?? []).map((cr) => (
        <li key={cr.id} className="flex items-center justify-between gap-2">
          <Link href={`/change-requests/${cr.id}`} className={linkClass}>
            {cr.requested_su_amount !== undefined
              ? `${formatNumber(cr.requested_su_amount)} SU`
              : (cr.requested_status ?? "Change")}
          </Link>
          <span className="text-xs text-muted-foreground">
            {cr.change_status} · {formatDate(cr.timestamp)}
          </span>
        </li>
      ))}
    </Section>
  );
}
