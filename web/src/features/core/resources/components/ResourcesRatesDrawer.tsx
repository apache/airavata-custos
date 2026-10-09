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

import { zodResolver } from "@hookform/resolvers/zod";
import * as React from "react";
import { useForm } from "react-hook-form";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { formatDay } from "@/shared/format";
import { Button } from "@/shared/ui/button";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { Input } from "@/shared/ui/input";
import { Label } from "@/shared/ui/label";
import { CardSkeleton } from "@/shared/ui/Loading";
import { SideDrawer } from "@/shared/ui/SideDrawer";
import { StatusBadge, type StatusBadgeVariant } from "@/shared/ui/StatusBadge";
import { toastOnSuccess } from "@/shared/ui/sonner";
import type { ComputeAllocationResourceRate } from "@/generated/core/types.gen";
import { z } from "zod";
import {
  useCreateResourceRate,
  useEffectiveRate,
  useResourceRate,
  useResourceRates,
} from "../queries";

type RateStatus = "ACTIVE" | "SCHEDULED" | "SUPERSEDED" | "EXPIRED";

// Effective = the containing window with the latest start_time, mirroring the
// backend ORDER BY start_time DESC. Overlaps resolve to SUPERSEDED, not errors.
function classifyRate(
  rate: ComputeAllocationResourceRate,
  all: ComputeAllocationResourceRate[],
): RateStatus {
  const now = Date.now();
  const start = Date.parse(rate.start_time ?? "");
  const end = Date.parse(rate.end_time ?? "");
  if (start > now) return "SCHEDULED";
  if (end <= now) return "EXPIRED";
  const effectiveStart = Math.max(
    ...all
      .filter((r) => Date.parse(r.start_time ?? "") <= now && Date.parse(r.end_time ?? "") > now)
      .map((r) => Date.parse(r.start_time ?? "")),
  );
  return start === effectiveStart ? "ACTIVE" : "SUPERSEDED";
}

const createRateSchema = z
  .object({
    rate: z.number({ message: "Rate is required" }).min(0, "Rate must be zero or greater"),
    start_date: z.string().min(1, "Start date is required"),
    end_date: z.string().min(1, "End date is required"),
  })
  .refine((d) => d.end_date > d.start_date, {
    path: ["end_date"],
    message: "End date must be after the start date",
  });
type CreateRateForm = z.infer<typeof createRateSchema>;

const STATUS_VARIANT: Record<RateStatus, StatusBadgeVariant> = {
  ACTIVE: "active",
  SCHEDULED: "pending",
  SUPERSEDED: "inactive",
  EXPIRED: "expired",
};

function todayUtc(): string {
  return new Date().toISOString().slice(0, 10);
}

function defaultForm(): CreateRateForm {
  const start = todayUtc();
  const end = new Date(start);
  end.setUTCFullYear(end.getUTCFullYear() + 1);
  return {
    // NaN leaves the number input empty until the user types.
    rate: Number.NaN,
    start_date: start,
    end_date: end.toISOString().slice(0, 10),
  };
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
  const ability = useAbility();
  const canManage = ability.can("write", "Allocation");

  const rates = useResourceRates(resourceId ?? undefined);
  const effective = useEffectiveRate(resourceId ?? undefined);
  const createMutation = useCreateResourceRate(resourceId ?? "");

  const history = React.useMemo(
    () =>
      [...(rates.data ?? [])].sort(
        (a, b) => Date.parse(b.start_time ?? "") - Date.parse(a.start_time ?? ""),
      ),
    [rates.data],
  );

  const [selectedRateId, setSelectedRateId] = React.useState<string>();
  const selectedRate = useResourceRate(selectedRateId);
  const [formOpen, setFormOpen] = React.useState(false);
  const form = useForm<CreateRateForm>({
    resolver: zodResolver(createRateSchema),
    defaultValues: defaultForm(),
    mode: "onBlur",
  });

  function openForm() {
    form.reset(defaultForm());
    setFormOpen(true);
  }

  const startDate = form.watch("start_date");
  const endDate = form.watch("end_date");
  const isImmediate = startDate === todayUtc();
  const currentRate = effective.data?.rate;
  // Backend forces an end date, so warn when nothing starts at or after it.
  const leavesGap =
    Boolean(endDate) &&
    !history.some((r) => Date.parse(r.start_time ?? "") >= Date.parse(`${endDate}T00:00:00Z`));

  const onSubmit = form.handleSubmit((values) => {
    if (!resourceId) return;
    const start_time =
      values.start_date === todayUtc()
        ? new Date().toISOString()
        : `${values.start_date}T00:00:00Z`;
    createMutation.mutate(
      {
        body: {
          compute_allocation_resource_id: resourceId,
          rate: values.rate,
          start_time,
          end_time: `${values.end_date}T00:00:00Z`,
        },
      },
      toastOnSuccess("Rate added", () => setFormOpen(false)),
    );
  });

  const errors = form.formState.errors;

  return (
    <SideDrawer
      open={Boolean(resourceId)}
      onOpenChange={(open) => {
        // The drawer stays mounted across resources, so closing drops the selection and the form.
        if (!open) {
          setSelectedRateId(undefined);
          setFormOpen(false);
        }
        onOpenChange(open);
      }}
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
            {formatDay(effective.data.start_time)} to {formatDay(effective.data.end_time)}
          </p>
        </div>
      ) : null}

      {canManage && history.length > 0 ? (
        <div className="mb-4 flex items-start justify-between gap-4">
          <p className="text-xs text-muted-foreground">
            Rates can&apos;t be edited. Adding a rate supersedes the current one from its effective
            date.
          </p>
          {!formOpen ? (
            <Button variant="outline" size="sm" className="shrink-0" onClick={openForm}>
              + Add rate
            </Button>
          ) : null}
        </div>
      ) : null}

      {canManage && formOpen ? (
        <form
          onSubmit={onSubmit}
          aria-label="Add rate form"
          className="mb-6 space-y-4 rounded-lg border border-border bg-card p-4"
        >
          <div className="space-y-2">
            <Label htmlFor="rate-value">Rate</Label>
            <Input
              id="rate-value"
              type="number"
              step="0.01"
              placeholder="e.g. 0.05"
              {...form.register("rate", { valueAsNumber: true })}
            />
            {errors.rate?.message ? (
              <p className="text-xs text-destructive">{errors.rate.message}</p>
            ) : null}
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="rate-start">Effective date</Label>
              <Input id="rate-start" type="date" {...form.register("start_date")} />
              {errors.start_date?.message ? (
                <p className="text-xs text-destructive">{errors.start_date.message}</p>
              ) : null}
            </div>
            <div className="space-y-2">
              <Label htmlFor="rate-end">Ends</Label>
              <Input id="rate-end" type="date" {...form.register("end_date")} />
              {errors.end_date?.message ? (
                <p className="text-xs text-destructive">{errors.end_date.message}</p>
              ) : null}
            </div>
          </div>

          <p className="text-xs text-muted-foreground">
            {isImmediate
              ? `This becomes the active rate immediately. Current rate (${currentRate ?? "none"}) stays in history.`
              : `Scheduled to take effect on ${formatDay(`${startDate}T00:00:00Z`)}. The current rate (${currentRate ?? "none"}) stays active until then.`}
          </p>

          {leavesGap ? (
            <p className="text-xs text-[color:var(--custos-amber-700)]">
              No rate is effective after {formatDay(`${endDate}T00:00:00Z`)}. Add a follow-up rate
              to avoid a pricing gap.
            </p>
          ) : null}

          <div className="flex items-center justify-end gap-2">
            <Button
              type="button"
              variant="ghost"
              onClick={() => setFormOpen(false)}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={createMutation.isPending}>
              {createMutation.isPending ? "Adding…" : "Add rate"}
            </Button>
          </div>
        </form>
      ) : null}

      {rates.isLoading ? (
        <CardSkeleton />
      ) : rates.error ? (
        <ErrorState message={rates.error.message} onRetry={() => rates.refetch()} />
      ) : history.length === 0 ? (
        <EmptyState
          heading="No rates recorded for this resource."
          cta={
            canManage && !formOpen ? (
              <Button variant="outline" size="sm" onClick={openForm}>
                + Add rate
              </Button>
            ) : undefined
          }
        />
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
            {history.map((rate) => {
              const status = classifyRate(rate, history);
              return (
                <tr key={rate.id} className="border-t border-border/60">
                  <td className="py-2 pr-4 font-medium text-foreground">
                    <button
                      type="button"
                      className="underline-offset-4 hover:underline"
                      aria-label={`Rate ${rate.rate} details`}
                      onClick={() => setSelectedRateId(rate.id)}
                    >
                      {rate.rate}
                    </button>
                  </td>
                  <td className="py-2 pr-4 text-muted-foreground">{formatDay(rate.start_time)}</td>
                  <td className="py-2 pr-4 text-muted-foreground">{formatDay(rate.end_time)}</td>
                  <td className="py-2">
                    <StatusBadge variant={STATUS_VARIANT[status]} label={status} />
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      )}

      {selectedRate.data ? (
        <dl className="mt-4 grid gap-x-8 gap-y-2 rounded-lg border border-border bg-card p-4 text-sm sm:grid-cols-[max-content_1fr]">
          <dt className="text-muted-foreground">Rate ID</dt>
          <dd className="font-mono text-foreground">{selectedRate.data.id}</dd>
          <dt className="text-muted-foreground">Rate</dt>
          <dd className="text-foreground">{selectedRate.data.rate} SU/unit</dd>
          <dt className="text-muted-foreground">Starts</dt>
          <dd className="text-foreground">{formatDay(selectedRate.data.start_time)}</dd>
          <dt className="text-muted-foreground">Ends</dt>
          <dd className="text-foreground">{formatDay(selectedRate.data.end_time)}</dd>
        </dl>
      ) : selectedRate.error ? (
        <ErrorState message={selectedRate.error.message} onRetry={() => selectedRate.refetch()} />
      ) : null}
    </SideDrawer>
  );
}
