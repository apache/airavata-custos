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
import { Button } from "@/shared/ui/button";
import { ConfirmDialog } from "@/shared/ui/ConfirmDialog";
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { CardSkeleton } from "@/shared/ui/Loading";
import { StatusBadge } from "@/shared/ui/StatusBadge";
import type { AppAuthorization } from "../api";
import { useAppAuthorizations, useAuthorizeApp, useRemoveAppAuthorization } from "../queries";

export function MyApps() {
  const apps = useAppAuthorizations();
  const authorize = useAuthorizeApp();
  const remove = useRemoveAppAuthorization();
  const [removing, setRemoving] = React.useState<AppAuthorization | null>(null);

  const columns: Array<DataTableColumn<AppAuthorization>> = [
    {
      key: "name",
      header: "App",
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
      key: "status",
      header: "Status",
      cell: (app) =>
        app.authorized ? (
          <StatusBadge variant="active" label="Allowed" />
        ) : (
          <StatusBadge variant="inactive" label="Not allowed" />
        ),
    },
    {
      key: "manage",
      header: "",
      cell: (app) => (
        <div className="flex justify-end">
          {app.authorized ? (
            <Button size="sm" variant="outline" onClick={() => setRemoving(app)}>
              Remove
            </Button>
          ) : (
            <Button
              size="sm"
              disabled={authorize.isPending}
              onClick={() =>
                authorize.mutate(app.id, { onError: (error) => toast.error(error.message) })
              }
            >
              Allow
            </Button>
          )}
        </div>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="font-display text-[28px] font-bold leading-tight">Apps</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Allow an app to generate SSH certificates for your account.
        </p>
      </div>

      {apps.isLoading ? (
        <CardSkeleton />
      ) : apps.error ? (
        <ErrorState message={apps.error.message} onRetry={() => apps.refetch()} />
      ) : apps.data?.length === 0 ? (
        <EmptyState
          heading="No apps need your permission."
          description="Apps that need it appear here."
        />
      ) : (
        <DataTable columns={columns} rows={apps.data ?? []} rowKey={(app) => app.id} />
      )}

      <ConfirmDialog
        open={removing !== null}
        onOpenChange={(open) => {
          if (!open) setRemoving(null);
        }}
        title={`Remove ${removing?.name ?? "app"}?`}
        confirmLabel="Remove"
        pending={remove.isPending}
        error={remove.error}
        onConfirm={() => {
          if (removing) remove.mutate(removing.id, { onSuccess: () => setRemoving(null) });
        }}
      >
        It can no longer generate certificates for you. Certificates it already generated are not
        revoked.
      </ConfirmDialog>
    </div>
  );
}
