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

import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { CardSkeleton } from "@/shared/ui/Loading";
import { SideDrawer } from "@/shared/ui/SideDrawer";
import { StatusBadge } from "@/shared/ui/StatusBadge";
import { useEffectiveRate, useResourceRates } from "../queries";
import { isRateActive, type Rate } from "../schemas";

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

export type ResourcesRatesDrawerProps = {
  resourceId: string | null;
  resourceName: string | null;
  onOpenChange: (open: boolean) => void;
};

export function ResourcesRatesDrawer({
  resourceId,
  resourceName,
  onOpenChange,
}: ResourcesRatesDrawerProps) {
  const rates = useResourceRates(resourceId ?? undefined);
  const effective = useEffectiveRate(resourceId ?? undefined);
  const history = rates.data ?? [];

  return (
    <SideDrawer
      open={Boolean(resourceId)}
      onOpenChange={onOpenChange}
      title={resourceName ? `Rates: ${resourceName}` : "Rates"}
      width="lg"
    >
      {effective.data ? (
        <div className="mb-6 space-y-2 rounded-lg border border-border bg-card p-4">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
              Effective rate
            </span>
            <StatusBadge variant="active" label="ACTIVE" />
          </div>
          <p className="font-display text-2xl font-bold text-foreground">{effective.data.rate}</p>
          <p className="text-sm text-muted-foreground">
            {formatDate(effective.data.start_time)} to {formatDate(effective.data.end_time)}
          </p>
        </div>
      ) : null}

      {rates.isLoading ? (
        <CardSkeleton />
      ) : rates.error ? (
        <ErrorState message={(rates.error as Error).message} onRetry={() => rates.refetch()} />
      ) : history.length === 0 ? (
        <EmptyState heading="No rates recorded for this resource." />
      ) : (
        <table className="w-full text-left text-sm">
          <thead className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
            <tr>
              <th className="py-2 pr-4 font-medium">Rate</th>
              <th className="py-2 pr-4 font-medium">From</th>
              <th className="py-2 pr-4 font-medium">To</th>
              <th className="py-2 font-medium">Status</th>
            </tr>
          </thead>
          <tbody>
            {history.map((rate: Rate) => (
              <tr key={rate.id} className="border-t border-border/60">
                <td className="py-2 pr-4 font-medium text-foreground">{rate.rate}</td>
                <td className="py-2 pr-4 text-muted-foreground">{formatDate(rate.start_time)}</td>
                <td className="py-2 pr-4 text-muted-foreground">{formatDate(rate.end_time)}</td>
                <td className="py-2">
                  {isRateActive(rate) ? (
                    <StatusBadge variant="active" label="ACTIVE" />
                  ) : (
                    <StatusBadge variant="expired" label="EXPIRED" />
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </SideDrawer>
  );
}
