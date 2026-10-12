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

import { zAllocationStatus } from "@/generated/core/zod.gen";
import { listOf } from "@/shared/api/listOf";
import { isGone, useOnGone } from "@/shared/api/useOnGone";
import { normalizeQueryId } from "@/shared/layout/nav";
import type { UseQueryResult } from "@tanstack/react-query";
import * as React from "react";
import { ErrorState } from "./ErrorState";
import { CardSkeleton } from "./Loading";
import { Button } from "./button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "./dialog";
import { Input } from "./input";
import { Label } from "./label";

// The query an edit form prefills from; see useFreshOnOpen.
type Refetchable = Pick<UseQueryResult, "refetch" | "dataUpdatedAt" | "errorUpdatedAt" | "error"> &
  Partial<Pick<UseQueryResult, "isEnabled">>;

// An edit form prefills from its record as the server holds it now: opening refetches the record
// and the form waits for that read, so edited() merges onto the fresh value. A record the
// read finds gone (a 404, or `missing` from a refetched list) calls onGone once, so its parent list
// drops the stale row.
export function useFreshOnOpen(
  query: Refetchable | undefined,
  open: boolean,
  { missing = false, onGone }: { missing?: boolean; onGone?: () => void } = {},
) {
  const [since, setSince] = React.useState<number>();
  // Waits for the record query to be enabled; one gated on the opening turns enabled with it.
  const enabled = open && query !== undefined && query.isEnabled !== false;
  // biome-ignore lint/correctness/useExhaustiveDependencies: refetches once per opening
  React.useEffect(() => {
    if (!enabled) return setSince(undefined);
    setSince(Date.now());
    // Its observer takes the enabled options in its own effect, after this child effect; a fetch
    // that enabling starts serves as the refetch.
    queueMicrotask(() => void query?.refetch({ cancelRefetch: false }));
  }, [enabled]);
  const error = query && since !== undefined && query.errorUpdatedAt >= since ? query.error : null;
  const ready = !query || (since !== undefined && query.dataUpdatedAt >= since);
  const gone = open && (isGone(error) || (ready && missing));
  useOnGone(gone, onGone);
  return { ready, error, gone };
}

const FIELD =
  "input:not([type=hidden]):not(:disabled), select:not(:disabled), textarea:not(:disabled), [role=combobox]:not([data-disabled])";
// A form's first enabled field, where an opening dialog or drawer puts focus.
export const firstField = (root: ParentNode | null | undefined) =>
  root?.querySelector<HTMLElement>(FIELD) ?? null;

type FormDialogProps = {
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  trigger?: React.ReactElement;
  title: string;
  description?: string;
  submitLabel: string;
  isPending: boolean;
  onSubmit: (form: FormData, close: () => void) => void;
  // An edit form's record: the fields wait for its refetch on open (useFreshOnOpen).
  record?: Refetchable;
  // Shown in place of the fields when the fresh read finds the record gone: a 404, or `missing`
  // once a refetched list lacks it; onGone then refreshes the list that showed it.
  gone?: string;
  missing?: boolean;
  onGone?: () => void;
  // Holds Save back, e.g. until a value differs.
  submitDisabled?: boolean;
  children: React.ReactNode;
};

export function FormDialog({
  open,
  onOpenChange,
  trigger,
  title,
  description,
  submitLabel,
  isPending,
  onSubmit,
  record,
  gone = "This record no longer exists",
  missing,
  onGone,
  submitDisabled = false,
  children,
}: FormDialogProps) {
  const [ownOpen, setOwnOpen] = React.useState(false);
  const fresh = useFreshOnOpen(record, open ?? ownOpen, { missing, onGone });
  const setOpen = onOpenChange ?? setOwnOpen;
  const popup = React.useRef<HTMLDivElement>(null);
  const form = React.useRef<HTMLFormElement>(null);
  // The fields of an edit form mount once its fresh record is ready; until then the dialog itself
  // holds focus, so a stray Enter does not press Cancel.
  const initialFocus = () => firstField(form.current) ?? popup.current;
  // biome-ignore lint/correctness/useExhaustiveDependencies: focuses once per ready record
  React.useEffect(() => {
    if (record && fresh.ready) initialFocus()?.focus();
  }, [fresh.ready]);
  return (
    <Dialog open={open ?? ownOpen} onOpenChange={setOpen}>
      {trigger ? <DialogTrigger render={trigger} /> : null}
      <DialogContent ref={popup} initialFocus={initialFocus}>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          {description ? <DialogDescription>{description}</DialogDescription> : null}
        </DialogHeader>
        <form
          ref={form}
          className="space-y-4"
          onSubmit={(e) => {
            e.preventDefault();
            onSubmit(new FormData(e.currentTarget), () => setOpen(false));
          }}
        >
          {fresh.gone ? (
            <ErrorState heading={gone} as="h3" />
          ) : fresh.ready ? (
            children
          ) : fresh.error ? (
            <ErrorState onRetry={() => record?.refetch()} as="h3" />
          ) : (
            <CardSkeleton />
          )}
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => setOpen(false)}>
              Cancel
            </Button>
            <Button
              type="submit"
              pending={isPending}
              disabled={!fresh.ready || fresh.gone || submitDisabled}
              focusableWhenDisabled
            >
              {isPending ? "Saving…" : submitLabel}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

export function Field({
  label,
  ...input
}: { label: string; name: string } & React.ComponentProps<typeof Input>) {
  const id = `field-${input.name}`;
  return (
    <div className="min-w-0 space-y-2">
      <Label htmlFor={id}>{label}</Label>
      <Input
        id={id}
        title={input.name === "name" ? "Enter a name, not only spaces." : undefined}
        {...input}
      />
    </div>
  );
}

export function SelectField({
  label,
  name,
  options,
  defaultValue,
  description,
}: {
  label: string;
  name: string;
  options: Array<{ value: string; label: string }>;
  defaultValue?: string;
  description?: string;
}) {
  const id = `field-${name}`;
  return (
    <div className="space-y-2">
      <Label htmlFor={id}>{label}</Label>
      <select
        id={id}
        name={name}
        aria-describedby={description ? `${id}-description` : undefined}
        defaultValue={defaultValue}
        required
        className="h-9 w-full rounded-md border bg-background px-3 text-sm"
      >
        {options.map((o) => (
          <option key={o.value} value={o.value}>
            {o.label}
          </option>
        ))}
      </select>
      {description ? (
        <p id={`${id}-description`} className="text-xs text-muted-foreground">
          {description}
        </p>
      ) : null}
    </div>
  );
}

// Select options for records that carry an id and a display name.
export const named = (rows: Array<{ id: string; name: string }> | null | undefined) =>
  listOf(rows).map((r) => ({ value: r.id, label: r.name }));

export const text = (form: FormData, key: string) => String(form.get(key) ?? "").trim();
// A pasted id, normalised like a ?param id so an uppercase UUID still matches.
export const idText = (form: FormData, key: string) => normalizeQueryId(text(form, key)) ?? "";
// The full record a PUT sends, which replaces the stored one: the fresh record with the form's
// fields over it, or null when no field differs (a blank field matching an absent one).
export const edited = <R extends object, T extends Record<string, unknown>>(
  current: R,
  next: T,
): (R & T) | null =>
  Object.entries(next).some(([k, v]) => v !== ((current as Record<string, unknown>)[k] ?? ""))
    ? { ...current, ...next }
    : null;
export const num = (form: FormData, key: string) => Number(form.get(key) ?? 0);
// Date inputs hold a calendar day, which the API stores at UTC midnight.
export const isoDate = (form: FormData, key: string) => `${text(form, key)}T00:00:00Z`;
// The UTC calendar day of a timestamp, the inverse of isoDate; the API may render it in another
// offset.
export const dateInput = (iso?: string) => (iso ? new Date(iso).toISOString().slice(0, 10) : "");

export const STATUSES = zAllocationStatus.options;
// "Active" for ACTIVE; "all" is a filter's catch-all.
export const statusLabel = (s: string) =>
  s === "all" ? "All statuses" : s.charAt(0) + s.slice(1).toLowerCase();
export const STATUS_OPTIONS = STATUSES.map((s) => ({ value: s, label: statusLabel(s) }));
export const status = (form: FormData, key: string) => STATUSES.find((s) => s === text(form, key));
