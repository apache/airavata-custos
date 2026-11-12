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
import type { App } from "../api";
import { useApps, useDisableApp, useEnableApp } from "../queries";
import { CreateAppDialog } from "./CreateAppDialog";
import { DeleteAppDialog } from "./DeleteAppDialog";
import { EditAppDialog } from "./EditAppDialog";
import { NewSecretDialog } from "./NewSecretDialog";

export function AppsView() {
  const ability = useAbility();
  const apps = useApps();
  const disable = useDisableApp();
  const enable = useEnableApp();
  const [creating, setCreating] = React.useState(false);
  const [editing, setEditing] = React.useState<App | null>(null);
  const [disabling, setDisabling] = React.useState<App | null>(null);
  const [newSecretFor, setNewSecretFor] = React.useState<App | null>(null);
  const [deleting, setDeleting] = React.useState<App | null>(null);

  if (ability.cannot("manage", "App")) {
    return <ErrorState message="Not permitted." />;
  }

  const columns: Array<DataTableColumn<App>> = [
    {
      key: "name",
      header: "Name",
      cell: (app) => (
        <div>
          <div className="font-medium">{app.name}</div>
          {app.description ? (
            <div className="text-xs text-muted-foreground">{app.description}</div>
          ) : null}
        </div>
      ),
    },
    {
      key: "key",
      header: "App key",
      cell: (app) => <span className="font-mono text-xs">{app.key}</span>,
    },
    {
      key: "privileges",
      header: "Privileges",
      cell: (app) =>
        app.privileges?.length ? (
          <div className="space-y-0.5 font-mono text-xs">
            {app.privileges.map((privilege) => (
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
      cell: (app) =>
        app.status === "ACTIVE" ? (
          <StatusBadge variant="active" />
        ) : (
          <StatusBadge variant="inactive" label="Disabled" />
        ),
    },
    {
      key: "hint",
      header: "Secret hint",
      cell: (app) => (
        <span className="font-mono text-xs">
          cak_{app.key}.{app.secret_hint}...
        </span>
      ),
    },
    {
      key: "manage",
      header: "",
      cell: (app) => (
        <div className="flex justify-end gap-2">
          <Button size="sm" variant="outline" onClick={() => setEditing(app)}>
            Edit
          </Button>
          <Button size="sm" variant="outline" onClick={() => setNewSecretFor(app)}>
            New secret
          </Button>
          {app.status === "ACTIVE" ? (
            <Button size="sm" variant="outline" onClick={() => setDisabling(app)}>
              Disable
            </Button>
          ) : (
            <>
              <Button
                size="sm"
                variant="outline"
                disabled={enable.isPending}
                onClick={() =>
                  enable.mutate(app.id, { onError: (error) => toast.error(error.message) })
                }
              >
                Enable
              </Button>
              <Button size="sm" variant="destructive" onClick={() => setDeleting(app)}>
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
          <h1 className="font-display text-[28px] font-bold leading-tight">Apps</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Trusted tools that call Custos with their own token.
          </p>
        </div>
        <Button onClick={() => setCreating(true)}>Create app</Button>
      </div>

      {apps.isLoading ? (
        <CardSkeleton />
      ) : apps.error ? (
        <ErrorState message={apps.error.message} onRetry={() => apps.refetch()} />
      ) : apps.data?.length === 0 ? (
        <EmptyState heading="No apps yet." description="Create one to give a tool its own token." />
      ) : (
        <DataTable columns={columns} rows={apps.data ?? []} rowKey={(app) => app.id} />
      )}

      <CreateAppDialog open={creating} onOpenChange={setCreating} />

      {editing ? <EditAppDialog app={editing} onClose={() => setEditing(null)} /> : null}

      {newSecretFor ? (
        <NewSecretDialog app={newSecretFor} onClose={() => setNewSecretFor(null)} />
      ) : null}

      {deleting ? <DeleteAppDialog app={deleting} onClose={() => setDeleting(null)} /> : null}

      <ConfirmDialog
        open={disabling !== null}
        onOpenChange={(open) => {
          if (!open) setDisabling(null);
        }}
        title={`Disable ${disabling?.name ?? "app"}?`}
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
