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

import type { SystemAccount } from "../api";
import { SlurmValue } from "./SlurmValue";
import { useSlurmCluster } from "../queries";
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import { ErrorState } from "@/shared/ui/ErrorState";
import { CardSkeleton } from "@/shared/ui/Loading";

const columns: Array<DataTableColumn<SystemAccount>> = [
  { key: "name", header: "Account", cell: (a) => <b>{a.name}</b> },
  {
    key: "users",
    header: "Users",
    cell: (a) =>
      (a.users ?? []).map((u) => (
        <div key={u.name} title={`Default account: ${u.default_account}`}>
          {u.name}{" "}
          <span className="text-muted-foreground">
            {u.qos?.length ? u.qos.join(", ") : "account's"}
          </span>
        </div>
      )),
  },
  { key: "qos", header: "QoS", cell: (a) => <SlurmValue custos={a.qos ?? []} /> },
  { key: "parent", header: "Parent", cell: (a) => a.parent || "—" },
];

export function SystemAccountsTab({ clusterId }: { clusterId: string }) {
  const accounts = useSlurmCluster(clusterId, (v) => v.system_accounts ?? []);
  if (accounts.isLoading) return <CardSkeleton />;
  if (accounts.error) {
    return <ErrorState message={accounts.error.message} onRetry={() => accounts.refetch()} />;
  }
  return (
    <div className="space-y-3">
      <p className="text-sm text-muted-foreground">
        SLURM accounts without a Custos allocation. Read-only.
      </p>
      <DataTable
        columns={columns}
        rows={accounts.data ?? []}
        rowKey={(a) => a.name}
        empty="No system accounts."
      />
    </div>
  );
}
