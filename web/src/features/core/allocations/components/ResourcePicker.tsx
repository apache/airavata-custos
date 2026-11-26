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

import { EffectiveRateCell } from "@/features/core/resources/components/ResourcesTab";
import { useResourceSummaries } from "@/features/core/resources/queries";
import type * as React from "react";

export type PickerOption = { id: string; label: string; extra?: React.ReactNode };

// The cluster's resources, each with its effective rate.
export function useResourceOptions(clusterId: string): PickerOption[] {
  return (useResourceSummaries().data ?? [])
    .filter((r) => r.compute_cluster_id === clusterId)
    .map((r) => ({ id: r.id, label: r.name, extra: <EffectiveRateCell resource={r} /> }));
}

export function ResourcePicker({
  options,
  value,
  onChange,
  empty = "No resources on this cluster.",
}: {
  options: PickerOption[];
  value: string[];
  onChange: (next: string[]) => void;
  empty?: string | undefined;
}) {
  if (options.length === 0) {
    return <p className="text-sm text-muted-foreground">{empty}</p>;
  }
  return (
    <ul className="divide-y rounded-md border">
      {options.map((o) => (
        <li key={o.id}>
          <label className="flex items-center gap-3 px-3 py-2">
            <input
              type="checkbox"
              checked={value.includes(o.id)}
              onChange={(e) =>
                onChange(e.target.checked ? [...value, o.id] : value.filter((id) => id !== o.id))
              }
            />
            <span className="flex-1 text-sm font-medium">{o.label}</span>
            {o.extra}
          </label>
        </li>
      ))}
    </ul>
  );
}
