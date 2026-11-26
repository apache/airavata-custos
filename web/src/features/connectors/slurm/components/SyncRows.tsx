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

import type * as React from "react";
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import type { Conflict } from "../api";
import { SlurmActions } from "./SlurmActions";
import { SlurmValue } from "./SlurmValue";

// One setting: Custos's value, Slurm's when it differs, and the conflicts to resolve.
export type SyncRow = {
  label: string;
  custos: unknown;
  slurm?: unknown;
  conflicts: Conflict[];
  items?: React.ReactNode;
};

export function SyncRows({
  clusterId,
  syncUnknown,
  rows,
  empty,
}: {
  clusterId: string;
  syncUnknown: boolean;
  rows: SyncRow[];
  empty?: string;
}) {
  const columns: Array<DataTableColumn<SyncRow>> = [
    { key: "label", header: "Setting", cell: (r) => r.label },
    {
      key: "value",
      header: "Value",
      cell: (r) => <SlurmValue custos={r.custos} slurm={r.slurm} />,
    },
    {
      key: "actions",
      header: "Actions",
      interactive: true,
      cell: (r) => (
        <SlurmActions
          clusterId={clusterId}
          scope="allocation"
          syncUnknown={syncUnknown}
          conflicts={r.conflicts}
          items={r.items}
        />
      ),
    },
  ];
  return (
    <DataTable
      columns={columns}
      rows={rows}
      rowKey={(r) => r.label}
      {...(empty ? { empty } : {})}
    />
  );
}
