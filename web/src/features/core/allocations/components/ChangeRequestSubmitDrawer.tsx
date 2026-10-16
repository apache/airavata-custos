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

import { useCurrentUser } from "@/features/core/identity/queries";
import type { ComputeAllocation, ComputeAllocationChangeRequest } from "@/generated/core/types.gen";
import { zAllocationStatus } from "@/generated/core/zod.gen";
import { formatSU } from "@/shared/format";
import { ErrorState } from "@/shared/ui/ErrorState";
import { firstField, statusLabel, useFreshOnOpen } from "@/shared/ui/FormDialog";
import { CardSkeleton } from "@/shared/ui/Loading";
import { SideDrawer } from "@/shared/ui/SideDrawer";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { Label } from "@/shared/ui/label";
import { toastError, toastOnSuccess } from "@/shared/ui/sonner";
import { zodResolver } from "@hookform/resolvers/zod";
import * as React from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import {
  alreadyDecided,
  useChangeRequest,
  useRefreshChangeRequests,
  useRefreshDecidedChangeRequest,
  useSubmitChangeRequest,
  useUpdateChangeRequest,
} from "../queries";

const REASON_MIN = 20;

// Approval records the requested SU total and status as the allocation's new values, so both are
// always requested.
const formSchema = z.object({
  su_amount: z
    .string()
    .trim()
    .regex(/^\d+(\.\d+)?$/, "Requested SUs must be a number.")
    .refine(
      (v) => Number(v) <= Number.MAX_SAFE_INTEGER,
      `Requested SUs must be at most ${formatSU(Number.MAX_SAFE_INTEGER)}.`,
    ),
  status: zAllocationStatus,
  reason: z.string().trim().min(REASON_MIN, `Reason must be at least ${REASON_MIN} characters.`),
});

// The server keeps a request's SUs when it receives 0, so an edit asks for more.
const editSchema = formSchema.extend({
  su_amount: formSchema.shape.su_amount.refine(
    (v) => Number(v) > 0,
    "Requested SUs must be above 0.",
  ),
});

type FormValues = z.infer<typeof formSchema>;

export type ChangeRequestSubmitDrawerProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  allocation: ComputeAllocation;
  // A PENDING request to edit instead of filing a new one.
  request?: ComputeAllocationChangeRequest;
};

const valuesOf = (
  allocation: ComputeAllocation,
  request?: ComputeAllocationChangeRequest,
): FormValues => ({
  su_amount: request ? String(request.requested_su_amount) : "",
  status: request?.requested_status ?? allocation.status,
  reason: request?.reason ?? "",
});

export function ChangeRequestSubmitDrawer({
  open,
  onOpenChange,
  allocation,
  request,
}: ChangeRequestSubmitDrawerProps) {
  const submitMutation = useSubmitChangeRequest();
  const editMutation = useUpdateChangeRequest();
  const mutation = request ? editMutation : submitMutation;
  const { user } = useCurrentUser();
  const form = useForm<FormValues>({
    resolver: zodResolver(request ? editSchema : formSchema),
    defaultValues: valuesOf(allocation, request),
    reValidateMode: "onChange",
  });
  // An edit prefills from the request as the server holds it when the drawer opens.
  const record = useChangeRequest(request?.id);
  const refreshRequests = useRefreshChangeRequests();
  const refreshDecided = useRefreshDecidedChangeRequest(request?.id ?? "");
  const fresh = useFreshOnOpen(request ? record : undefined, open, { onGone: refreshRequests });
  // The value the form opened with; a later background refetch does not reset a draft.
  const [opened, setOpened] = React.useState<ComputeAllocationChangeRequest>();
  const formRef = React.useRef<HTMLFormElement>(null);

  // biome-ignore lint/correctness/useExhaustiveDependencies: resets once per opening
  React.useEffect(() => {
    if (!open || !fresh.ready) return;
    const current = request ? (record.data ?? request) : undefined;
    setOpened(current);
    form.reset(valuesOf(allocation, current));
    firstField(formRef.current)?.focus();
  }, [open, fresh.ready]);

  const onSubmit = form.handleSubmit(async (values) => {
    const saved = valuesOf(allocation, opened);
    if (
      opened &&
      values.su_amount === saved.su_amount &&
      values.status === saved.status &&
      values.reason === saved.reason.trim()
    ) {
      return onOpenChange(false);
    }
    const change = {
      requested_su_amount: Number(values.su_amount),
      requested_status: values.status,
      reason: values.reason,
    };
    if (opened) {
      // A decision made since the drawer opened stands; the edit body never carries one.
      const { data: current } = await record.refetch();
      if (!current) return;
      if (current.change_status !== "PENDING") {
        refreshDecided();
        return toastError(alreadyDecided(current.change_status));
      }
    }
    await toastOnSuccess(
      opened
        ? editMutation.mutateAsync({
            path: { id: opened.id },
            body: {
              compute_allocation_id: opened.compute_allocation_id,
              requester_id: opened.requester_id,
              timestamp: opened.timestamp,
              ...change,
            },
          })
        : submitMutation.mutateAsync({
            body: { compute_allocation_id: allocation.id, requester_id: user?.id, ...change },
          }),
      request ? "Change request updated" : "Change request submitted",
      () => onOpenChange(false),
    );
  });

  const errorOf = (name: keyof FormValues) =>
    form.getFieldState(name, form.formState).error?.message;

  return (
    <SideDrawer
      open={open}
      onOpenChange={onOpenChange}
      title={request ? "Edit change request" : "Submit change request"}
      initialFocus={() => firstField(formRef.current) ?? true}
      description={`Request a new SU total for this allocation, which started with ${formatSU(allocation.initial_su_amount)} SUs, or a new status.`}
    >
      <form
        ref={formRef}
        onSubmit={onSubmit}
        className="space-y-4"
        aria-label={request ? "Edit change request form" : "Submit change request form"}
      >
        {fresh.gone ? (
          <ErrorState heading="This change request no longer exists" />
        ) : fresh.error ? (
          <ErrorState onRetry={() => record.refetch()} />
        ) : !fresh.ready ? (
          <CardSkeleton />
        ) : (
          <>
            <div className="space-y-2">
              <Label htmlFor="cr-amount">Requested SUs</Label>
              <Input
                id="cr-amount"
                type="number"
                min={0}
                max={Number.MAX_SAFE_INTEGER}
                step="any"
                placeholder="The allocation's new SU total"
                {...form.register("su_amount")}
              />
              {errorOf("su_amount") ? (
                <p className="text-xs text-destructive-text">{errorOf("su_amount")}</p>
              ) : null}
            </div>

            <div className="space-y-2">
              <Label htmlFor="cr-status">Requested status</Label>
              <select
                id="cr-status"
                className="h-9 w-full rounded-md border bg-background px-3 text-sm"
                {...form.register("status")}
              >
                {zAllocationStatus.options.map((s) => (
                  <option key={s} value={s}>
                    {statusLabel(s)}
                  </option>
                ))}
              </select>
            </div>

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
                <p className="text-xs text-destructive-text">{errorOf("reason")}</p>
              ) : null}
            </div>
          </>
        )}
        <div className="flex items-center justify-end gap-2 pt-2">
          <Button type="button" variant="ghost" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button
            type="submit"
            pending={mutation.isPending || form.formState.isSubmitting}
            disabled={!fresh.ready || !user?.id}
            focusableWhenDisabled
          >
            {mutation.isPending ? "Saving…" : request ? "Save changes" : "Submit request"}
          </Button>
        </div>
      </form>
    </SideDrawer>
  );
}
