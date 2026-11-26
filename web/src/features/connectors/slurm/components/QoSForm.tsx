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

// A QoS definition form: creates without `qos`, edits with it, or, with `compare`,
// resolves a conflict by picking each field from Custos or Slurm.

import * as React from "react";
import type { ComputeClusterQoS, ComputeClusterQoSInput } from "@/generated/core/zod.gen";
import { Button } from "@/shared/ui/button";
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/shared/ui/dialog";
import { Input } from "@/shared/ui/input";
import { Label } from "@/shared/ui/label";
import { asRecord } from "../format";

// A QoS definition's editable fields, in form order.
const QOS_FIELDS = [
  { key: "priority", label: "Priority" },
  { key: "max_wall_minutes", label: "MaxWall (minutes)" },
  { key: "max_jobs_per_user", label: "MaxJobsPerUser" },
  { key: "max_tres_per_user", label: "MaxTRESPerUser" },
  { key: "flags", label: "Flags" },
] as const;

type QoSFieldKey = (typeof QOS_FIELDS)[number]["key"];

const text = (value: unknown): string => (value == null ? "" : String(value));

// A QoS definition as form text; unset fields are blank.
function qosText(q: Partial<ComputeClusterQoS> | null | undefined): Record<QoSFieldKey, string> {
  return {
    priority: text(q?.priority),
    max_wall_minutes: text(q?.max_wall_minutes),
    max_jobs_per_user: text(q?.max_jobs_per_user),
    max_tres_per_user: text(q?.max_tres_per_user),
    flags: q?.flags?.join(", ") ?? "",
  };
}

const optionalNumber = (value: string): number | undefined =>
  value.trim() === "" ? undefined : Number(value);

// A QoS definition body from form text; blank limits are unset.
function qosBody(get: (key: QoSFieldKey) => string): ComputeClusterQoSInput {
  return {
    priority: Number(get("priority")),
    max_wall_minutes: optionalNumber(get("max_wall_minutes")),
    max_jobs_per_user: optionalNumber(get("max_jobs_per_user")),
    max_tres_per_user: get("max_tres_per_user").trim(),
    flags: get("flags")
      .split(",")
      .map((s) => s.trim())
      .filter(Boolean),
  };
}

type QoSFormProps = {
  title: string;
  qos?: ComputeClusterQoS | undefined;
  compare?: { custos: unknown; slurm: unknown };
  pending: boolean;
  error?: string | null | undefined;
  onCancel: () => void;
  onSubmit: (body: ComputeClusterQoSInput) => void;
};

const SIDES = [
  { side: "custos", label: "Custos" },
  { side: "slurm", label: "SLURM" },
] as const;

export function QoSForm({ title, qos, compare, pending, error, onCancel, onSubmit }: QoSFormProps) {
  const sides = compare && {
    custos: qosText(asRecord<ComputeClusterQoS>(compare.custos)),
    slurm: qosText(asRecord<ComputeClusterQoS>(compare.slurm)),
  };
  const [values, setValues] = React.useState<Record<string, string>>(() => ({
    name: qos?.name ?? "",
    ...(sides?.custos ?? qosText(qos)),
  }));
  const [isDefault, setIsDefault] = React.useState(qos?.is_default ?? false);
  const get = (key: string) => values[key] ?? "";
  const set = (key: string, value: string) => setValues((v) => ({ ...v, [key]: value }));
  const creating = !compare && !qos;
  const canSubmit =
    (!creating || get("name").trim() !== "") && get("priority").trim() !== "" && !pending;

  const rows: Array<{ key: string; label: string }> = [
    ...(creating ? [{ key: "name", label: "Name" }] : []),
    ...QOS_FIELDS,
  ];

  return (
    <Dialog open onOpenChange={(next) => !next && !pending && onCancel()}>
      <DialogContent className={compare ? "sm:max-w-2xl" : undefined}>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
        </DialogHeader>
        <form
          className="space-y-4"
          onSubmit={(e) => {
            e.preventDefault();
            if (!canSubmit) return;
            const body = qosBody(get);
            onSubmit(compare ? body : { name: get("name").trim(), ...body, is_default: isDefault });
          }}
        >
          <div
            className={`grid items-center gap-2 text-sm ${compare ? "grid-cols-[8rem_1fr_1fr_1fr]" : "grid-cols-[8rem_1fr]"}`}
          >
            {sides ? (
              <>
                <span />
                {SIDES.map((s) => (
                  <span key={s.side} className="text-xs font-medium text-muted-foreground">
                    {s.label}
                  </span>
                ))}
                <span className="text-xs font-medium text-muted-foreground">Result</span>
              </>
            ) : null}
            {rows.map((f) => (
              <React.Fragment key={f.key}>
                <Label htmlFor={`qos-${f.key}`}>{f.label}</Label>
                {sides
                  ? SIDES.map(({ side, label }) => {
                      const value = sides[side][f.key as QoSFieldKey];
                      return (
                        <label key={side} className="flex items-center gap-1.5">
                          <input
                            type="radio"
                            name={f.key}
                            aria-label={`${f.label} ${label}`}
                            checked={get(f.key) === value}
                            onChange={() => set(f.key, value)}
                          />
                          {value || "—"}
                        </label>
                      );
                    })
                  : null}
                <Input
                  id={`qos-${f.key}`}
                  value={get(f.key)}
                  onChange={(e) => set(f.key, e.target.value)}
                />
              </React.Fragment>
            ))}
          </div>
          {compare ? null : (
            <label className="flex items-center gap-2 text-sm">
              <input
                type="checkbox"
                checked={isDefault}
                disabled={qos?.is_default ?? false}
                onChange={(e) => setIsDefault(e.target.checked)}
              />
              Default QoS
            </label>
          )}
          {error ? (
            <p role="alert" className="text-sm text-destructive">
              {error}
            </p>
          ) : null}
          <DialogFooter>
            <Button type="button" variant="outline" onClick={onCancel} disabled={pending}>
              Cancel
            </Button>
            <Button type="submit" disabled={!canSubmit}>
              {creating ? "Create" : "Save"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
