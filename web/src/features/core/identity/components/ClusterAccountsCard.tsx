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

import { clusterAccountStatus } from "@/features/core/cluster-accounts/components/ClusterAccountBadge";
import type { CallerClusterAccount } from "@/generated/core/zod.gen";
import { StatusIcon } from "@/shared/ui/StatusIcon";
import { Card, CardDescription, CardHeader, CardTitle } from "@/shared/ui/card";
import { HardDrive } from "lucide-react";

export function ClusterAccountsCard({ accounts }: { accounts: CallerClusterAccount[] }) {
  return (
    <Card className="gap-0 py-0">
      <CardHeader className="px-6 pt-5 pb-4">
        <CardTitle>Cluster accounts</CardTitle>
        <CardDescription>
          The clusters you can sign in to, and your username on each.
        </CardDescription>
      </CardHeader>
      {accounts.length === 0 ? (
        <p className="border-t border-border px-6 py-6 text-sm text-muted-foreground">
          No cluster accounts yet.
        </p>
      ) : (
        <ul className="divide-y divide-border border-t border-border">
          {accounts.map((account) => {
            const status = clusterAccountStatus(account);
            const active = status.variant === "active";
            return (
              <li key={account.id} className="flex flex-wrap items-center gap-3 px-6 py-3.5">
                <StatusIcon icon={HardDrive} active={active} label={status.label} />
                <span className="min-w-[88px] font-semibold">{account.cluster_name}</span>
                <span className="flex-1 font-mono text-[13px]">{account.local_username}</span>
                {active ? null : (
                  <span className="ml-auto text-xs text-muted-foreground">{status.label}</span>
                )}
              </li>
            );
          })}
        </ul>
      )}
    </Card>
  );
}
