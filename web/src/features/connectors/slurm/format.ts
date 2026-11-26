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

// Formats untyped conflict values as table-cell lines.

const MINUTES_PER_DAY = 1440;

export function formatMinutes(minutes: number | null | undefined): string {
  if (minutes == null) return "—";
  const days = Math.floor(minutes / MINUTES_PER_DAY);
  const hours = Math.floor((minutes % MINUTES_PER_DAY) / 60);
  const rest = minutes % 60;
  const clock = [hours, rest].map((n) => String(n).padStart(2, "0")).join(":");
  return `${days > 0 ? `${days}-` : ""}${clock}:00`;
}

type Limits = {
  max_wall_minutes?: number | null;
  max_jobs_per_user?: number | null;
  max_tres_per_user?: string | null;
};

// A QoS definition's limits, one per line; unset limits are left out.
export function limitLines(q: Limits): string[] {
  const lines: string[] = [];
  if (q.max_wall_minutes != null) lines.push(`MaxWall ${formatMinutes(q.max_wall_minutes)}`);
  if (q.max_jobs_per_user != null) lines.push(`MaxJobsPerUser ${q.max_jobs_per_user}`);
  if (q.max_tres_per_user) lines.push(`MaxTRESPerUser ${q.max_tres_per_user}`);
  return lines;
}

// The fields a conflict value may carry: a QoS definition, or a QoS list.
type ConflictRecord = Partial<
  Record<
    | "priority"
    | "max_wall_minutes"
    | "max_jobs_per_user"
    | "max_tres_per_user"
    | "flags"
    | "qos"
    | "default",
    unknown
  >
>;

// One line per item of a conflict value: a QoS definition, a QoS list, or a scalar.
export function conflictLines(value: unknown): string[] {
  if (value == null || value === "") return [];
  if (typeof value !== "object") return [String(value)];
  if (Array.isArray(value)) return value.map(String);
  const record = value as ConflictRecord;
  if (Array.isArray(record.qos)) {
    const lines = record.qos.map(String);
    return record.default ? [...lines, `default: ${record.default}`] : lines;
  }
  if ("priority" in record) {
    return [
      `Priority ${record.priority}`,
      ...limitLines(record as Limits),
      ...conflictLines(record.flags),
    ];
  }
  return [JSON.stringify(value)];
}

export function asRecord<T = ConflictRecord>(value: unknown): T | null {
  return value && typeof value === "object" ? (value as T) : null;
}

export type Joined<C, S> = { name: string; custos?: C | undefined; slurm?: S | undefined };

// One row per name on either side: Custos's in order, then the ones only Slurm has.
export function joinByName<C extends { name: string }, S extends { name: string }>(
  custos: C[],
  slurm: S[],
): Joined<C, S>[] {
  const rows = new Map<string, Joined<C, S>>(
    custos.map((c) => [c.name, { name: c.name, custos: c }]),
  );
  for (const s of slurm) rows.set(s.name, { ...rows.get(s.name), name: s.name, slurm: s });
  return [...rows.values()];
}
