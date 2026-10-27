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

import { ApiError } from "@/shared/api/client";
import { SideDrawer } from "@/shared/ui/SideDrawer";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { Label } from "@/shared/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/shared/ui/select";
import * as React from "react";
import { toast } from "sonner";
import { z } from "zod";
import { useSubmitChangeRequest } from "../queries";

const CHANGE_TYPE_LABELS: Record<string, string> = {
  INCREASE_CREDITS: "Increase credits (SUs)",
  EXTEND_END_DATE: "Extend end date",
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
    requested_change_type: z.literal("OTHER"),
    reason: reasonField,
  }),
]);

type ChangeType = z.infer<typeof formSchema>["requested_change_type"];
// Every field any variant has; the schema keeps the ones the chosen type needs.
type Draft = { requested_amount: string; requested_end_date: string; reason: string };
const emptyDraft: Draft = { requested_amount: "", requested_end_date: "", reason: "" };

export type ChangeRequestSubmitDrawerProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  allocationId: string;
  requesterId: string;
  currentSuAmount: number;
};

export function ChangeRequestSubmitDrawer({
  open,
  onOpenChange,
  allocationId,
  requesterId,
  currentSuAmount,
}: ChangeRequestSubmitDrawerProps) {
  const submitMutation = useSubmitChangeRequest();
  const [changeType, setChangeType] = React.useState<ChangeType>("INCREASE_CREDITS");
  const [draft, setDraft] = React.useState(emptyDraft);
  const [errors, setErrors] = React.useState<Partial<Record<keyof Draft, string | undefined>>>({});
  const field = (key: keyof Draft) => ({
    value: draft[key],
    onChange: (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) =>
      setDraft((d) => ({ ...d, [key]: e.target.value })),
  });

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    const parsed = formSchema.safeParse({
      ...draft,
      requested_change_type: changeType,
      requested_amount: Number.parseFloat(draft.requested_amount),
    });
    if (!parsed.success) {
      const fields: Partial<Record<keyof Draft, string[]>> = z.flattenError(
        parsed.error,
      ).fieldErrors;
      setErrors({
        requested_amount: fields.requested_amount?.[0],
        requested_end_date: fields.requested_end_date?.[0],
        reason: fields.reason?.[0],
      });
      return;
    }
    setErrors({});
    const values = parsed.data;
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

    try {
      await submitMutation.mutateAsync({
        compute_allocation_id: allocationId,
        requested_su_amount: requestedSuAmount,
        requested_status: "ACTIVE",
        reason: reasonWithMeta,
        requester_id: requesterId,
      });
      toast.success("Change request submitted");
      setChangeType("INCREASE_CREDITS");
      setDraft(emptyDraft);
      onOpenChange(false);
    } catch (err) {
      const msg =
        err instanceof ApiError ? `Failed (${err.status}): ${err.message}` : "Failed to submit";
      toast.error(msg);
    }
  }

  function onChangeTypeChange(next: ChangeType) {
    setChangeType(next);
    setDraft((d) => ({ ...emptyDraft, reason: d.reason }));
    setErrors({});
  }

  return (
    <SideDrawer
      open={open}
      onOpenChange={onOpenChange}
      title="Submit change request"
      description="Request extra SUs, a new end date, or describe another change."
    >
      <form onSubmit={onSubmit} className="space-y-4" aria-label="Submit change request form">
        <div className="space-y-2">
          <Label htmlFor="cr-type">Change type</Label>
          <Select value={changeType} onValueChange={(v) => v !== null && onChangeTypeChange(v)}>
            <SelectTrigger id="cr-type" aria-label="Change type">
              <SelectValue>{(value: string) => CHANGE_TYPE_LABELS[value] ?? value}</SelectValue>
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="INCREASE_CREDITS">Increase credits (SUs)</SelectItem>
              <SelectItem value="EXTEND_END_DATE">Extend end date</SelectItem>
              <SelectItem value="OTHER">Other</SelectItem>
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
              {...field("requested_amount")}
            />
            {errors.requested_amount ? (
              <p className="text-xs text-destructive">{errors.requested_amount}</p>
            ) : null}
          </div>
        ) : null}

        {changeType === "EXTEND_END_DATE" ? (
          <div className="space-y-2">
            <Label htmlFor="cr-end">Requested new end date</Label>
            <Input id="cr-end" type="date" {...field("requested_end_date")} />
            {errors.requested_end_date ? (
              <p className="text-xs text-destructive">{errors.requested_end_date}</p>
            ) : null}
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
            {...field("reason")}
          />
          {errors.reason ? <p className="text-xs text-destructive">{errors.reason}</p> : null}
        </div>

        <div className="flex items-center justify-end gap-2 pt-2">
          <Button type="button" variant="ghost" onClick={() => onOpenChange(false)}>
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
