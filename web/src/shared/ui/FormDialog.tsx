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

import * as React from "react";
import { zAllocationStatus } from "@/generated/core/zod.gen";
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

// Uncontrolled form dialog: fields carry `name` and `defaultValue`, and submit hands back FormData.
// Given a `trigger` element it owns its open state; otherwise the caller controls `open`.
type FormDialogProps = {
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  trigger?: React.ReactElement;
  title: string;
  description?: string;
  submitLabel: string;
  isPending: boolean;
  onSubmit: (form: FormData, close: () => void) => void;
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
  children,
}: FormDialogProps) {
  const [ownOpen, setOwnOpen] = React.useState(false);
  const setOpen = onOpenChange ?? setOwnOpen;
  return (
    <Dialog open={open ?? ownOpen} onOpenChange={setOpen}>
      {trigger ? <DialogTrigger render={trigger} /> : null}
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          {description ? <DialogDescription>{description}</DialogDescription> : null}
        </DialogHeader>
        <form
          className="space-y-4"
          onSubmit={(e) => {
            e.preventDefault();
            onSubmit(new FormData(e.currentTarget), () => setOpen(false));
          }}
        >
          {children}
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => setOpen(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={isPending}>
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
    <div className="space-y-2">
      <Label htmlFor={id}>{label}</Label>
      <Input id={id} {...input} />
    </div>
  );
}

export function SelectField({
  label,
  name,
  options,
  defaultValue,
}: {
  label: string;
  name: string;
  options: Array<{ value: string; label: string }>;
  defaultValue?: string;
}) {
  const id = `field-${name}`;
  return (
    <div className="space-y-2">
      <Label htmlFor={id}>{label}</Label>
      <select
        id={id}
        name={name}
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
    </div>
  );
}

// Select options for records that carry an id and a display name.
export const named = (rows: Array<{ id?: string; name?: string }> = []) =>
  rows.map((r) => ({ value: r.id ?? "", label: r.name ?? "" }));

export const text = (form: FormData, key: string) => String(form.get(key) ?? "").trim();
export const num = (form: FormData, key: string) => Number(form.get(key) ?? 0);
// Date inputs hold a local calendar day; the API takes RFC 3339 timestamps.
export const isoDate = (form: FormData, key: string) =>
  new Date(`${text(form, key)}T00:00:00`).toISOString();
// The local calendar day of a timestamp, the inverse of isoDate (sv-SE formats as YYYY-MM-DD).
export const dateInput = (iso?: string) => (iso ? new Date(iso).toLocaleDateString("sv-SE") : "");

export const STATUSES = zAllocationStatus.options;
export const STATUS_OPTIONS = STATUSES.map((s) => ({ value: s, label: s }));
// "Active" for ACTIVE; "all" is a filter's catch-all.
export const statusLabel = (s: string) =>
  s === "all" ? "All statuses" : s.charAt(0) + s.slice(1).toLowerCase();
export const status = (form: FormData, key: string) =>
  STATUSES.find((s) => s === text(form, key));
