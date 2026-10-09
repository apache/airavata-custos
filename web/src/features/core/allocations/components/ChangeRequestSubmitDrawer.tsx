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
import { useForm } from "react-hook-form";
import { z } from "zod";
import type { AllocationStatus } from "@/generated/core/types.gen";
import { zAllocationStatus } from "@/generated/core/zod.gen";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { Label } from "@/shared/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/shared/ui/select";
import { SideDrawer } from "@/shared/ui/SideDrawer";
import { toastOnSuccess } from "@/shared/ui/sonner";
import { useSubmitChangeRequest } from "../queries";

const CHANGE_TYPE_LABELS: Record<string, string> = {
  INCREASE_CREDITS: "Increase credits (SUs)",
  EXTEND_END_DATE: "Extend end date",
  CHANGE_STATUS: "Change status",
  OTHER: "Other",
};

const REASON_MIN = 20;
const reasonField = z.string().min(REASON_MIN, `Reason must be at least ${REASON_MIN} characters`);

const formSchema = z.discriminatedUnion("requested_change_type", [
  z.object({
    requested_change_type: z.literal("INCREASE_CREDITS"),
    requested_amount: z.number().int().positive({
      message: "Additional SUs must be a positive integer",
    }),
    reason: reasonField,
  }),
  z.object({
    requested_change_type: z.literal("EXTEND_END_DATE"),
    requested_end_date: z.string().min(1, "Requested end date is required"),
    reason: reasonField,
  }),
  z.object({
    requested_change_type: z.literal("CHANGE_STATUS"),
    requested_status: zAllocationStatus,
    reason: reasonField,
  }),
  z.object({
    requested_change_type: z.literal("OTHER"),
    reason: reasonField,
  }),
]);

type FormValues = z.infer<typeof formSchema>;
type ChangeType = FormValues["requested_change_type"];

// NaN leaves the number input empty until the user types.
const EMPTY: FormValues = {
  requested_change_type: "INCREASE_CREDITS",
  requested_amount: Number.NaN,
  reason: "",
};

export type ChangeRequestSubmitDrawerProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  allocationId: string;
  requesterId: string;
  currentSuAmount: number;
  currentStatus?: AllocationStatus;
};

export function ChangeRequestSubmitDrawer({
  open,
  onOpenChange,
  allocationId,
  requesterId,
  currentSuAmount,
  currentStatus,
}: ChangeRequestSubmitDrawerProps) {
  const submitMutation = useSubmitChangeRequest();
  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: EMPTY,
    mode: "onBlur",
  });

  const changeType = form.watch("requested_change_type");
  // A cancelled draft does not survive into the next open.
  const setOpen = (next: boolean) => {
    if (!next) form.reset(EMPTY);
    onOpenChange(next);
  };

  const onSubmit = form.handleSubmit((values) => {
    const requestedSuAmount =
      values.requested_change_type === "INCREASE_CREDITS"
        ? currentSuAmount + values.requested_amount
        : currentSuAmount;
    const reasonWithMeta =
      values.requested_change_type === "EXTEND_END_DATE"
        ? `[EXTEND_END_DATE → ${values.requested_end_date}] ${values.reason}`
        : values.requested_change_type === "OTHER"
          ? `[OTHER] ${values.reason}`
          : values.reason;

    submitMutation.mutate(
      {
        body: {
          compute_allocation_id: allocationId,
          requested_su_amount: requestedSuAmount,
          requested_status:
            values.requested_change_type === "CHANGE_STATUS" ? values.requested_status : currentStatus,
          reason: reasonWithMeta,
          requester_id: requesterId,
        },
      },
      toastOnSuccess("Change request submitted", () => setOpen(false)),
    );
  });

  function onChangeTypeChange(next: ChangeType) {
    const reason = form.getValues("reason");
    if (next === "INCREASE_CREDITS") {
      form.reset({ ...EMPTY, reason });
    } else if (next === "EXTEND_END_DATE") {
      form.reset({
        requested_change_type: "EXTEND_END_DATE",
        reason,
        requested_end_date: "",
      });
    } else if (next === "CHANGE_STATUS") {
      form.reset({ requested_change_type: "CHANGE_STATUS", reason, requested_status: "INACTIVE" });
    } else {
      form.reset({ requested_change_type: "OTHER", reason });
    }
  }

  const errorOf = (name: "requested_amount" | "requested_end_date" | "reason") =>
    form.getFieldState(name, form.formState).error?.message;

  return (
    <SideDrawer
      open={open}
      onOpenChange={setOpen}
      title="Submit change request"
      description="Request extra SUs, a new end date, a status change, or describe another change."
    >
      <form onSubmit={onSubmit} className="space-y-4" aria-label="Submit change request form">
        <div className="space-y-2">
          <Label htmlFor="cr-type">Change type</Label>
          <Select value={changeType} onValueChange={(v) => v && onChangeTypeChange(v)}>
            <SelectTrigger id="cr-type" aria-label="Change type">
              <SelectValue>{(value: string) => CHANGE_TYPE_LABELS[value] ?? value}</SelectValue>
            </SelectTrigger>
            <SelectContent>
              {Object.entries(CHANGE_TYPE_LABELS).map(([value, label]) => (
                <SelectItem key={value} value={value}>
                  {label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {changeType === "INCREASE_CREDITS" ? (
          <div className="space-y-2">
            <Label htmlFor="cr-amount">Additional SUs requested</Label>
            <Input
              id="cr-amount"
              type="number"
              min={1}
              placeholder="e.g. 10000"
              {...form.register("requested_amount", { valueAsNumber: true })}
            />
            {errorOf("requested_amount") ? (
              <p className="text-xs text-destructive">{errorOf("requested_amount")}</p>
            ) : null}
          </div>
        ) : null}

        {changeType === "EXTEND_END_DATE" ? (
          <div className="space-y-2">
            <Label htmlFor="cr-end">Requested new end date</Label>
            <Input id="cr-end" type="date" {...form.register("requested_end_date")} />
            {errorOf("requested_end_date") ? (
              <p className="text-xs text-destructive">{errorOf("requested_end_date")}</p>
            ) : null}
          </div>
        ) : null}

        {changeType === "CHANGE_STATUS" ? (
          <div className="space-y-2">
            <Label htmlFor="cr-status">Requested status</Label>
            <select
              id="cr-status"
              className="h-9 w-full rounded-md border bg-background px-3 text-sm"
              {...form.register("requested_status")}
            >
              {zAllocationStatus.options.map((s) => (
                <option key={s} value={s}>
                  {s}
                </option>
              ))}
            </select>
          </div>
        ) : null}

        <div className="space-y-2">
          <Label htmlFor="cr-reason">
            Reason <span className="text-muted-foreground">(min {REASON_MIN} chars)</span>
          </Label>
          <textarea
            id="cr-reason"
            rows={5}
            className="w-full rounded-lg border border-input bg-transparent px-3 py-2 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50"
            placeholder="Why is this change needed?"
            {...form.register("reason")}
          />
          {errorOf("reason") ? (
            <p className="text-xs text-destructive">{errorOf("reason")}</p>
          ) : null}
        </div>

        <div className="flex items-center justify-end gap-2 pt-2">
          <Button type="button" variant="ghost" onClick={() => setOpen(false)}>
            Cancel
          </Button>
          <Button type="submit" disabled={submitMutation.isPending}>
            {submitMutation.isPending ? "Submitting…" : "Submit request"}
          </Button>
        </div>
      </form>
    </SideDrawer>
  );
}
