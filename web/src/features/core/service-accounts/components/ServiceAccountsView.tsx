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
import { toast } from "sonner";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { Button } from "@/shared/ui/button";
import { ConfirmDialog } from "@/shared/ui/ConfirmDialog";
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { CardSkeleton } from "@/shared/ui/Loading";
import { StatusBadge } from "@/shared/ui/StatusBadge";
import type { ServiceAccount } from "../api";
import { useDisableServiceAccount, useEnableServiceAccount, useServiceAccounts } from "../queries";
import { CreateServiceAccountDialog } from "./CreateServiceAccountDialog";
import { DeleteServiceAccountDialog } from "./DeleteServiceAccountDialog";
import { EditServiceAccountDialog } from "./EditServiceAccountDialog";
import { NewSecretDialog } from "./NewSecretDialog";

export function ServiceAccountsView() {
  const ability = useAbility();
  const accounts = useServiceAccounts();
  const disable = useDisableServiceAccount();
  const enable = useEnableServiceAccount();
  const [creating, setCreating] = React.useState(false);
  const [editing, setEditing] = React.useState<ServiceAccount | null>(null);
  const [disabling, setDisabling] = React.useState<ServiceAccount | null>(null);
  const [newSecretFor, setNewSecretFor] = React.useState<ServiceAccount | null>(null);
  const [deleting, setDeleting] = React.useState<ServiceAccount | null>(null);

  if (ability.cannot("manage", "ServiceAccount")) {
    return <ErrorState message="Not permitted." />;
  }

  const columns: Array<DataTableColumn<ServiceAccount>> = [
    {
      key: "name",
      header: "Name",
      cell: (account) => (
        <div>
          <div className="font-medium">{account.name}</div>
          {account.description ? (
            <div className="text-xs text-muted-foreground">{account.description}</div>
          ) : null}
        </div>
      ),
    },
    {
      key: "key",
      header: "Key",
      cell: (account) => <span className="font-mono text-xs">{account.key}</span>,
    },
    {
      key: "privileges",
      header: "Privileges",
      cell: (account) =>
        account.privileges?.length ? (
          <div className="space-y-0.5 font-mono text-xs">
            {account.privileges.map((privilege) => (
              <div key={privilege}>{privilege}</div>
            ))}
          </div>
        ) : (
          <span className="text-muted-foreground">None</span>
        ),
    },
    {
      key: "status",
      header: "Status",
      cell: (account) =>
        account.status === "ACTIVE" ? (
          <StatusBadge variant="active" />
        ) : (
          <StatusBadge variant="inactive" label="Disabled" />
        ),
    },
    {
      key: "hint",
      header: "Secret hint",
      cell: (account) => (
        <span className="font-mono text-xs">
          csa_{account.key}.{account.secret_hint}...
        </span>
      ),
    },
    {
      key: "manage",
      header: "",
      cell: (account) => (
        <div className="flex justify-end gap-2">
          <Button size="sm" variant="outline" onClick={() => setEditing(account)}>
            Edit
          </Button>
          <Button size="sm" variant="outline" onClick={() => setNewSecretFor(account)}>
            New secret
          </Button>
          {account.status === "ACTIVE" ? (
            <Button size="sm" variant="outline" onClick={() => setDisabling(account)}>
              Disable
            </Button>
          ) : (
            <>
              <Button
                size="sm"
                variant="outline"
                disabled={enable.isPending}
                onClick={() =>
                  enable.mutate(account.id, { onError: (error) => toast.error(error.message) })
                }
              >
                Enable
              </Button>
              <Button size="sm" variant="destructive" onClick={() => setDeleting(account)}>
                Delete
              </Button>
            </>
          )}
        </div>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 className="font-display text-[28px] font-bold leading-tight">Service accounts</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Trusted tools that call Custos with their own token.
          </p>
        </div>
        <Button onClick={() => setCreating(true)}>Create service account</Button>
      </div>

      {accounts.isLoading ? (
        <CardSkeleton />
      ) : accounts.error ? (
        <ErrorState message={accounts.error.message} onRetry={() => accounts.refetch()} />
      ) : accounts.data?.length === 0 ? (
        <EmptyState
          heading="No service accounts yet."
          description="Create one to give a tool its own token."
        />
      ) : (
        <DataTable columns={columns} rows={accounts.data ?? []} rowKey={(account) => account.id} />
      )}

      <CreateServiceAccountDialog open={creating} onOpenChange={setCreating} />

      {editing ? (
        <EditServiceAccountDialog account={editing} onClose={() => setEditing(null)} />
      ) : null}

      {newSecretFor ? (
        <NewSecretDialog account={newSecretFor} onClose={() => setNewSecretFor(null)} />
      ) : null}

      {deleting ? (
        <DeleteServiceAccountDialog account={deleting} onClose={() => setDeleting(null)} />
      ) : null}

      <ConfirmDialog
        open={disabling !== null}
        onOpenChange={(open) => {
          if (!open) setDisabling(null);
        }}
        title={`Disable ${disabling?.name ?? "service account"}?`}
        confirmLabel="Disable"
        pending={disable.isPending}
        error={disable.error}
        onConfirm={() => {
          if (disabling) disable.mutate(disabling.id, { onSuccess: () => setDisabling(null) });
        }}
      >
        Its token stops working at once. You can enable it again, or delete it.
      </ConfirmDialog>
    </div>
  );
}
