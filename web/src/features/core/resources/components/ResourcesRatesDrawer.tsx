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

import { dayAfter } from "@/features/core/allocations/components/AllocationStatusLabel";
import type { ComputeAllocationResourceRate } from "@/generated/core/types.gen";
import { isGone } from "@/shared/api/useOnGone";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { formatDay, formatRate } from "@/shared/format";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { CardSkeleton } from "@/shared/ui/Loading";
import { SideDrawer } from "@/shared/ui/SideDrawer";
import { StatusBadge, type StatusBadgeVariant } from "@/shared/ui/StatusBadge";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { Label } from "@/shared/ui/label";
import { toastOnSuccess } from "@/shared/ui/sonner";
import { zodResolver } from "@hookform/resolvers/zod";
import * as React from "react";
import { type DefaultValues, useForm } from "react-hook-form";
import { z } from "zod";
import {
  type RateStatus,
  effectiveRate,
  rateStatus,
  useCreateResourceRate,
  useEffectiveRate,
  useResource,
  useResourceRate,
  useResourceRates,
} from "../queries";

const createRateSchema = z
  .object({
    rate: z
      .number({
        error: (issue) => (issue.input === undefined ? "Rate is required." : "Enter a valid rate."),
      })
      .min(0, "Rate must be zero or greater."),
    start_date: z.string().min(1, "Start date is required."),
    end_date: z.string().min(1, "End date is required."),
  })
  .refine((d) => d.end_date > d.start_date, {
    path: ["end_date"],
    message: "End date must be after the start date.",
    // Also while another field fails, so every error shows at once.
    when: ({ value }) => {
      const { start_date, end_date } = value as { start_date?: string; end_date?: string };
      return Boolean(start_date && end_date);
    },
  });
type CreateRateForm = z.infer<typeof createRateSchema>;

const STATUS_VARIANT: Record<RateStatus["kind"], StatusBadgeVariant> = {
  ACTIVE: "active",
  SCHEDULED: "pending",
  NEVER: "inactive",
  OVERRIDDEN: "inactive",
  EXPIRED: "expired",
};

function statusLabel(status: RateStatus): string {
  if (status.kind === "NEVER") return "Never takes effect";
  if (status.kind === "SCHEDULED") return `Scheduled from ${formatDay(status.from)}`;
  if (status.kind !== "OVERRIDDEN") return status.kind === "ACTIVE" ? "Active" : "Expired";
  return status.until ? `Overridden until ${formatDay(status.until)}` : "Overridden";
}

// resumesOn: the draft's end, when the current rate is in force again there.
function draftCopy(
  status: RateStatus,
  currentRate: number | undefined,
  resumesOn: string | undefined,
) {
  const keeps = (where: string) =>
    currentRate === undefined ? "" : ` The current rate (${formatRate(currentRate)}) ${where}.`;
  switch (status.kind) {
    case "ACTIVE":
      return `This becomes the active rate immediately.${keeps(
        resumesOn ? `comes back into force on ${formatDay(resumesOn)}` : "stays in history",
      )}`;
    case "SCHEDULED":
      return `Scheduled to take effect on ${formatDay(status.from)}.${keeps(
        resumesOn
          ? `stays active until then and resumes on ${formatDay(resumesOn)}`
          : "stays active until then",
      )}`;
    case "OVERRIDDEN":
      return status.until
        ? `An existing rate stays in force; this rate takes effect on ${formatDay(status.until)}.`
        : "An existing rate stays in force for this rate's whole window, so it never takes effect.";
    case "NEVER":
      return "Other rates stay in force for this rate's whole window, so it never takes effect.";
    case "EXPIRED":
      return "This window has already ended, so the rate never takes effect.";
  }
}

function todayUtc(): string {
  return new Date().toISOString().slice(0, 10);
}

// A start of today submits as now, so the draft is judged as the server will see it.
const startTime = (day: string) =>
  day === todayUtc() ? new Date().toISOString() : `${day}T00:00:00Z`;

// No rate, so the number input stays empty until the user types.
function defaultForm(): DefaultValues<CreateRateForm> {
  const start = todayUtc();
  const end = new Date(start);
  end.setUTCFullYear(end.getUTCFullYear() + 1);
  return {
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
  const resource = useResource(resourceId ?? undefined);
  const name = resourceName ?? resource.data?.name;
  return (
    <SideDrawer
      open={Boolean(resourceId)}
      onOpenChange={onOpenChange}
      title={name ? `Rates: ${name}` : "Rates"}
      width="lg"
      recordKey={resourceId ?? undefined}
    >
      {resource.isLoading ? (
        <CardSkeleton />
      ) : isGone(resource.error) ? (
        <ErrorState heading="This resource no longer exists" as="h3" />
      ) : resource.error ? (
        <ErrorState error={resource.error} onRetry={() => resource.refetch()} as="h3" />
      ) : resourceId ? (
        // Keyed, so the selection and the add-rate form never carry over to another resource.
        <RatesPanel key={resourceId} resourceId={resourceId} />
      ) : null}
    </SideDrawer>
  );
}

function RatesPanel({ resourceId }: { resourceId: string }) {
  const ability = useAbility();
  const canManage = ability.can("write", "Allocation");

  const rates = useResourceRates(resourceId);
  const effective = useEffectiveRate(resourceId);
  const createMutation = useCreateResourceRate(resourceId);

  const history = React.useMemo(
    () =>
      [...(rates.data ?? [])].sort((a, b) => Date.parse(b.start_time) - Date.parse(a.start_time)),
    [rates.data],
  );

  const [selectedRateId, setSelectedRateId] = React.useState<string>();
  const selectedRate = useResourceRate(selectedRateId);
  const [formOpen, setFormOpen] = React.useState(false);
  const form = useForm<CreateRateForm>({
    resolver: zodResolver(createRateSchema),
    defaultValues: defaultForm(),
    reValidateMode: "onChange",
  });

  // A number input reads an unparsable entry (1e309) as blank; only its validity tells them apart.
  const rateInput = React.useRef<HTMLInputElement | null>(null);
  const rateField = form.register("rate", {
    setValueAs: (v: string) =>
      v ? Number(v) : rateInput.current?.validity.badInput ? Number.NaN : undefined,
  });

  const addRate = React.useRef<HTMLButtonElement>(null);
  const returnFocus = React.useRef(false);
  function openForm() {
    form.reset(defaultForm());
    setFormOpen(true);
  }
  function closeForm() {
    returnFocus.current = true;
    setFormOpen(false);
  }
  // "+ Add rate" unmounts while the form is open, so focus moves to the form and back by hand.
  // biome-ignore lint/correctness/useExhaustiveDependencies: follows the form opening and closing
  React.useEffect(() => {
    if (formOpen) form.setFocus("rate");
    else if (returnFocus.current) addRate.current?.focus();
    returnFocus.current = false;
  }, [formOpen]);

  const startDate = form.watch("start_date");
  const endDate = form.watch("end_date");
  const draft = {
    id: "",
    rate: 0,
    compute_allocation_resource_id: resourceId,
    start_time: startTime(startDate),
    end_time: `${endDate}T00:00:00Z`,
  };
  const draftStatus = rateStatus(draft, [...history, draft]);
  // A scheduled preview names it only when it is the rate in force just before the draft takes over.
  const currentRate =
    effective.data &&
    (draftStatus.kind !== "SCHEDULED" ||
      effectiveRate(history, Date.parse(draftStatus.from) - 1)?.id === effective.data.id)
      ? effective.data.rate
      : undefined;
  const draftEnd = Date.parse(draft.end_time);
  const afterDraft = effectiveRate(history, draftEnd);
  const resumesOn =
    effective.data && afterDraft?.id === effective.data.id ? draft.end_time : undefined;
  // The preview and gap warnings wait for a window the form would accept.
  const validWindow = Boolean(startDate) && endDate > startDate;
  // Gaps count only from now on; one entirely in the past changes nothing.
  const now = Date.now();
  const nextStart = history
    .map((r) => r.start_time)
    .filter((t) => Date.parse(t) > draftEnd)
    .sort((a, b) => Date.parse(a) - Date.parse(b))[0];
  // Backend forces an end date, so warn when no existing rate is in force from it.
  const leavesGap =
    validWindow &&
    draftEnd > now &&
    !afterDraft &&
    (nextStart === undefined || Date.parse(nextStart) > now);
  // The latest end among rates starting before the draft (now when there is none); earlier
  // than its start is a gap.
  const priorEnd = Math.max(
    ...history
      .filter((r) => Date.parse(r.start_time) < Date.parse(draft.start_time))
      .map((r) => Date.parse(r.end_time)),
  );
  const gapFrom = Math.max(priorEnd, now);
  const gapBefore = validWindow && gapFrom < Date.parse(draft.start_time) ? gapFrom : undefined;

  const onSubmit = form.handleSubmit((values) => {
    toastOnSuccess(
      createMutation.mutateAsync({
        body: {
          compute_allocation_resource_id: resourceId,
          rate: values.rate,
          start_time: startTime(values.start_date),
          end_time: `${values.end_date}T00:00:00Z`,
        },
      }),
      "Rate added",
      closeForm,
    );
  });

  const errors = form.formState.errors;

  return (
    <>
      {effective.data ? (
        <div className="mb-6 space-y-2 rounded-lg border border-border bg-card p-4">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
              Effective rate
            </span>
            <StatusBadge variant="active" label="Active" />
          </div>
          <p className="font-display text-2xl font-bold text-foreground">
            {formatRate(effective.data.rate)}
          </p>
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
            <Button
              ref={addRate}
              variant="outline"
              size="sm"
              className="shrink-0"
              onClick={openForm}
            >
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
              step="any"
              placeholder="e.g. 0.05"
              {...rateField}
              ref={(el) => {
                rateField.ref(el);
                rateInput.current = el;
              }}
            />
            <p className="min-h-4 text-xs text-destructive-text">{errors.rate?.message}</p>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="rate-start">Effective date</Label>
              <Input
                id="rate-start"
                type="date"
                {...form.register("start_date", { deps: "end_date" })}
              />
              <p className="min-h-4 text-xs text-destructive-text">{errors.start_date?.message}</p>
            </div>
            <div className="space-y-2">
              <Label htmlFor="rate-end">Ends</Label>
              <Input
                id="rate-end"
                type="date"
                min={dayAfter(startDate)}
                {...form.register("end_date")}
              />
              <p className="min-h-4 text-xs text-destructive-text">{errors.end_date?.message}</p>
            </div>
          </div>

          {validWindow ? (
            <p className="text-xs text-muted-foreground">
              {draftCopy(draftStatus, currentRate, resumesOn)}
            </p>
          ) : null}

          {gapBefore !== undefined ? (
            <p className="text-xs text-[color:var(--tone-warn-fg)]">
              No rate is effective from {formatDay(new Date(gapBefore).toISOString())} until{" "}
              {formatDay(draft.start_time)}.
            </p>
          ) : null}

          {leavesGap ? (
            <p className="text-xs text-[color:var(--tone-warn-fg)]">
              {nextStart
                ? `No rate is effective from ${formatDay(draft.end_time)} until ${formatDay(nextStart)}.`
                : `No rate is effective after ${formatDay(draft.end_time)}. Add a follow-up rate to avoid a pricing gap.`}
            </p>
          ) : null}

          <div className="flex items-center justify-end gap-2">
            <Button type="button" variant="ghost" onClick={closeForm}>
              Cancel
            </Button>
            <Button type="submit" pending={createMutation.isPending}>
              {createMutation.isPending ? "Adding…" : "Add rate"}
            </Button>
          </div>
        </form>
      ) : null}

      {rates.isLoading ? (
        <CardSkeleton />
      ) : rates.error ? (
        <ErrorState error={rates.error} onRetry={() => rates.refetch()} />
      ) : history.length === 0 ? (
        <EmptyState
          heading="No rates recorded for this resource"
          cta={
            canManage && !formOpen ? (
              <Button ref={addRate} variant="outline" size="sm" onClick={openForm}>
                + Add rate
              </Button>
            ) : undefined
          }
        />
      ) : (
        <table className="w-full text-left text-sm">
          <caption className="sr-only">Rates</caption>
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
              const status = rateStatus(rate, history);
              return (
                <tr key={rate.id} className="border-t border-border/60">
                  <td className="py-2 pr-4 font-medium text-foreground">
                    <button
                      type="button"
                      className="underline-offset-4 hover:underline"
                      aria-label={`Details of rate ${formatRate(rate.rate)} from ${formatDay(rate.start_time)}`}
                      onClick={() => setSelectedRateId(rate.id)}
                    >
                      {formatRate(rate.rate)}
                    </button>
                  </td>
                  <td className="py-2 pr-4 text-muted-foreground">{formatDay(rate.start_time)}</td>
                  <td className="py-2 pr-4 text-muted-foreground">{formatDay(rate.end_time)}</td>
                  <td className="py-2">
                    <StatusBadge
                      variant={STATUS_VARIANT[status.kind]}
                      label={statusLabel(status)}
                    />
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
          <dd className="text-foreground">{formatRate(selectedRate.data.rate)} SU/unit</dd>
          <dt className="text-muted-foreground">Starts</dt>
          <dd className="text-foreground">{formatDay(selectedRate.data.start_time)}</dd>
          <dt className="text-muted-foreground">Ends</dt>
          <dd className="text-foreground">{formatDay(selectedRate.data.end_time)}</dd>
        </dl>
      ) : selectedRate.error ? (
        <ErrorState error={selectedRate.error} onRetry={() => selectedRate.refetch()} />
      ) : null}
    </>
  );
}
