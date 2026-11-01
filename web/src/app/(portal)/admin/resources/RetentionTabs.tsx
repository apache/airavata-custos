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

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { CornerDownRight, HardDrive, Pencil, Server, User } from "lucide-react";
import Link from "next/link";
import * as React from "react";
import { toast } from "sonner";
import { z } from "zod";
import {
  type ComputeAllocation,
  type ComputeClusterUser,
  zComputeAllocation,
  zComputeClusterUser,
} from "@/generated/core/zod.gen";
import { apiFetch } from "@/shared/api/client";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { Button } from "@/shared/ui/button";
import { DataTable, type DataTableColumn, type DataTableProps } from "@/shared/ui/DataTable";
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/shared/ui/dialog";
import { ErrorState } from "@/shared/ui/ErrorState";
import { Input } from "@/shared/ui/input";
import { TableSkeleton } from "@/shared/ui/Loading";
import { Label } from "@/shared/ui/label";
import { UsageBar } from "@/shared/ui/UsageBar";

const folderSchema = z.object({
  id: z.string(),
  cluster_id: z.string(),
  path: z.string(),
  owner_kind: z.enum(["cluster_user", "allocation"]),
  owner_name: z.string(),
  unix_owner: z.string(),
  allocation_id: z.string().nullable(),
  allocation_name: z.string().nullable(),
  start_time: z.string().nullable(),
  end_time: z.string().nullable(),
  quota_until: z.string().nullable(),
  status: z.enum(["active", "pending_deletion"]),
  quota: z.object({ hard_limit: z.number(), used_capacity: z.number() }).nullable(),
});
type Folder = z.infer<typeof folderSchema> & { inherited: boolean };

// Marks a folder without a quota inside a folder with one, whose quota it counts against.
const nest = (folders: Array<z.infer<typeof folderSchema>>): Folder[] =>
  folders.map((f) => ({
    ...f,
    inherited:
      !f.quota &&
      folders.some(
        (g) => g.quota && g.cluster_id === f.cluster_id && f.path.startsWith(`${g.path}/`),
      ),
  }));

// Rows an admin may remove.
const DISPOSABLE = "bg-[color:var(--tone-warn-bg)]/40";

const day = (iso: string) => new Date(iso).toLocaleDateString();

const mono = (text: string) => <span className="font-mono text-xs">{text}</span>;

const inherited = <span className="text-xs text-muted-foreground">Inherited</span>;

const bytes = (n: number) => {
  const i = n > 0 ? Math.min(5, Math.floor(Math.log2(n) / 10)) : 0;
  return `${(n / 1024 ** i).toFixed(i ? 1 : 0)} ${["B", "KiB", "MiB", "GiB", "TiB", "PiB"][i]}`;
};

type Target = { path: string; done: string; method?: "POST" | "DELETE" };

// An admin's row button that deletes target's path, or posts to it, toasts the outcome and refreshes the list.
function useDeleteColumn<T>(
  list: string,
  variant: "destructive" | "outline",
  label: string,
  target: (row: T) => Target | false,
  [action, subject] = ["manage", "Cluster"],
): DataTableColumn<T> {
  const client = useQueryClient();
  const canManage = useAbility().can(action, subject);
  const del = useMutation({
    mutationFn: ({ path, method = "DELETE" }: Target) => apiFetch(path, { method }),
    onSuccess: (_, { done }) => {
      toast.success(done);
      return client.invalidateQueries({ queryKey: [list] });
    },
    onError: (err) => toast.error(err.message),
  });
  return {
    key: `actions-${label}`,
    header: "",
    width: "6rem",
    align: "right",
    interactive: true,
    cell: (row) => {
      const t = canManage && target(row);
      return t ? (
        <Button size="xs" variant={variant} disabled={del.isPending} onClick={() => del.mutate(t)}>
          {label}
        </Button>
      ) : null;
    },
  };
}

function useRetireColumns<T extends { id: string }>(
  list: string,
  subject: string,
  base: string,
  live: (row: T) => boolean,
  name: (row: T) => string,
) {
  return [
    useDeleteColumn<T>(
      list,
      "outline",
      "Deactivate",
      (row) =>
        live(row) && {
          path: `${base}/${row.id}/deactivate`,
          done: `${name(row)} deactivated`,
          method: "POST",
        },
      ["deactivate", subject],
    ),
    useDeleteColumn<T>(
      list,
      "destructive",
      "Delete",
      (row) => !live(row) && { path: `${base}/${row.id}`, done: `${name(row)} deleted` },
      ["delete", subject],
    ),
  ];
}

function Listing<T>({
  list,
  path,
  parse,
  keep = () => true,
  ...table
}: Omit<DataTableProps<T>, "rows"> & {
  list: string;
  path: string;
  parse: (data: unknown) => T[];
  keep?: (row: T) => boolean;
}) {
  const query = useQuery({
    queryKey: [list, path],
    queryFn: async () => parse(await apiFetch(path)),
  });
  if (query.isLoading) return <TableSkeleton rows={6} columns={table.columns.length} />;
  if (query.error)
    return <ErrorState message={query.error.message} onRetry={() => void query.refetch()} />;
  return <DataTable {...table} rows={(query.data ?? []).filter(keep)} />;
}

const GIB = 1024 ** 3;

// Sets a temporary hard limit on a folder's own quota; the change and its history are audited under cluster:path.
function QuotaDialog({ folder, onClose }: { folder: Folder; onClose: () => void }) {
  const client = useQueryClient();
  const [gib, setGib] = React.useState(String((folder.quota?.hard_limit ?? 0) / GIB));
  const [until, setUntil] = React.useState("");
  const [reason, setReason] = React.useState("");
  const latest = new Date(Date.now() + 90 * 864e5).toISOString().slice(0, 10);
  const save = useMutation({
    mutationFn: () =>
      apiFetch(`/connectors/vast/folders/${folder.id}/quota`, {
        method: "PUT",
        body: {
          hard_limit: Math.round(Number(gib) * GIB),
          until: new Date(`${until}T23:59:59`).toISOString(),
          reason,
        },
      }),
    onSuccess: () => {
      toast.success(`${folder.path} quota set until ${until}`);
      onClose();
      return client.invalidateQueries({ queryKey: ["vast-folders"] });
    },
    onError: (err) => toast.error(err.message),
  });
  return (
    <Dialog open onOpenChange={(open) => !open && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Quota for {folder.path}</DialogTitle>
        </DialogHeader>
        <form
          className="space-y-4"
          onSubmit={(e) => {
            e.preventDefault();
            save.mutate();
          }}
        >
          <div className="space-y-2">
            <Label htmlFor="quota-gib">Hard limit (GiB)</Label>
            <Input
              id="quota-gib"
              type="number"
              min="1"
              step="any"
              value={gib}
              onChange={(e) => setGib(e.target.value)}
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="quota-until">Until, then back to the configured limit</Label>
            <Input
              id="quota-until"
              type="date"
              max={latest}
              value={until}
              onChange={(e) => setUntil(e.target.value)}
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="quota-reason">Reason</Label>
            <Input id="quota-reason" value={reason} onChange={(e) => setReason(e.target.value)} />
          </div>
          <DialogFooter>
            <Link
              href={`/admin/traces?q=${encodeURIComponent(`${folder.cluster_id}:${folder.path}`)}&source=vast`}
              className="mr-auto self-center text-sm hover:underline"
            >
              History
            </Link>
            <Button type="button" variant="outline" onClick={onClose}>
              Cancel
            </Button>
            <Button type="submit" disabled={save.isPending}>
              Save
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

// Folders Custos created on VAST. A deleted owner's folder stays here until an admin deletes it.
export function StorageFoldersTab({ clusterId }: { clusterId: string | undefined }) {
  const canManage = useAbility().can("manage", "Cluster");
  const [editing, setEditing] = React.useState<Folder | null>(null);
  const del = useDeleteColumn<Folder>(
    "vast-folders",
    "destructive",
    "Delete",
    (row) =>
      row.status === "pending_deletion" && {
        path: `/connectors/vast/folders/${row.id}`,
        done: `${row.path} deleted`,
      },
  );
  const columns: Array<DataTableColumn<Folder>> = [
    {
      key: "path",
      header: "Folder",
      cell: (row) => {
        const Icon = row.inherited ? CornerDownRight : HardDrive;
        return (
          <span
            className="flex items-center gap-2"
            title={row.status === "active" ? "Active" : "Owner deleted"}
          >
            <Icon
              aria-hidden
              className={`size-3.5 shrink-0 ${row.status === "active" ? "text-[color:var(--tone-ok-fg)]" : "text-muted-foreground/40"}`}
            />
            {mono(row.path)}
          </span>
        );
      },
    },
    {
      key: "owner",
      header: "Owner",
      width: "16rem",
      interactive: true,
      cell: (row) => (
        <span
          className="flex items-center gap-3 whitespace-nowrap"
          title={row.unix_owner.replace(/^0:/, "root:")}
        >
          {row.allocation_id ? (
            <Link
              href={`/allocations/${row.allocation_id}`}
              className="flex items-center gap-1 hover:underline"
            >
              <Server aria-hidden className="size-3.5 text-muted-foreground" />
              {row.allocation_name}
            </Link>
          ) : null}
          {row.owner_kind === "cluster_user" ? (
            <span className="flex items-center gap-1">
              <User aria-hidden className="size-3.5 text-muted-foreground" />
              {mono(row.owner_name)}
            </span>
          ) : null}
        </span>
      ),
    },
    {
      key: "validity",
      header: "Validity",
      width: "13rem",
      cell: (row) =>
        row.start_time && row.end_time ? (
          <span className="whitespace-nowrap">
            {day(row.start_time)} – {day(row.end_time)}
          </span>
        ) : (
          "—"
        ),
    },
    {
      key: "quota",
      header: "Quota",
      width: "7rem",
      interactive: true,
      cell: (row) =>
        row.quota ? (
          <span className="flex items-center gap-1 whitespace-nowrap">
            {bytes(row.quota.hard_limit)}
            {row.quota_until && new Date(row.quota_until) > new Date() ? (
              <span className="text-xs text-muted-foreground">until {day(row.quota_until)}</span>
            ) : null}
            {canManage ? (
              <button type="button" aria-label="Edit quota" onClick={() => setEditing(row)}>
                <Pencil
                  aria-hidden
                  className="size-3 text-muted-foreground hover:text-foreground"
                />
              </button>
            ) : null}
          </span>
        ) : row.inherited ? (
          inherited
        ) : (
          "—"
        ),
    },
    {
      key: "used",
      header: "Used",
      width: "12rem",
      cell: (row) =>
        row.quota ? (
          <UsageBar
            size="sm"
            value={row.quota.used_capacity}
            max={row.quota.hard_limit}
            label={bytes(row.quota.used_capacity)}
            className="min-w-32"
          />
        ) : row.inherited ? (
          inherited
        ) : null,
    },
    del,
  ];
  return (
    <>
      <Listing
        list="vast-folders"
        path="/connectors/vast/folders"
        parse={(data) => nest(z.array(folderSchema).parse(data))}
        columns={columns}
        rowKey={(row) => row.id}
        rowClassName={(row) => (row.status === "active" ? undefined : DISPOSABLE)}
        keep={(row) => row.cluster_id === clusterId}
        empty="No folders recorded."
      />
      {editing ? <QuotaDialog folder={editing} onClose={() => setEditing(null)} /> : null}
    </>
  );
}

const statusIcon = (Icon: typeof User, active: boolean, text: React.ReactNode) => (
  <span className="flex items-center gap-2" title={active ? "Active" : "Inactive"}>
    <Icon
      aria-hidden
      className={`size-3.5 shrink-0 ${active ? "text-[color:var(--tone-ok-fg)]" : "text-muted-foreground/40"}`}
    />
    {text}
  </span>
);

// A cluster's accounts. A deactivated one stays, its access blocked and its username held, until an admin deletes it here.
export function UsersTab({ clusterId }: { clusterId: string | undefined }) {
  const retire = useRetireColumns<ComputeClusterUser>(
    "cluster-users",
    "ComputeClusterUser",
    "/compute-cluster-users",
    (row) => row.status === "ACTIVE" && row.approval_status === "APPROVED",
    (row) => row.local_username,
  );
  const columns: Array<DataTableColumn<ComputeClusterUser>> = [
    {
      key: "username",
      header: "Username",
      cell: (row) => statusIcon(User, row.status === "ACTIVE", mono(row.local_username)),
    },
    { key: "access", header: "Access", cell: (row) => row.access_level },
    { key: "approval", header: "Approval", cell: (row) => row.approval_status },
    {
      key: "provisioned",
      header: "Provisioned",
      cell: (row) => (row.provisioned_at ? day(row.provisioned_at) : "—"),
    },
    ...retire,
  ];
  return clusterId ? (
    <Listing
      list="cluster-users"
      path={`/compute-clusters/${clusterId}/users`}
      parse={(data) => z.array(zComputeClusterUser).parse(data ?? [])}
      columns={columns}
      rowKey={(row) => row.id}
      rowClassName={(row) => (row.status === "ACTIVE" ? undefined : DISPOSABLE)}
      empty="No accounts on this cluster."
    />
  ) : null;
}

export function AllocationsTab({ clusterId }: { clusterId: string | undefined }) {
  const retire = useRetireColumns<ComputeAllocation>(
    "cluster-allocations",
    "ComputeAllocation",
    "/compute-allocations",
    (row) => row.status === "ACTIVE",
    (row) => row.name,
  );
  const columns: Array<DataTableColumn<ComputeAllocation>> = [
    {
      key: "name",
      header: "Allocation",
      interactive: true,
      cell: (row) =>
        statusIcon(
          Server,
          row.status === "ACTIVE",
          <Link href={`/allocations/${row.id}`} className="hover:underline">
            {row.name}
          </Link>,
        ),
    },
    {
      key: "group",
      header: "Group",
      cell: (row) => (row.posix_group ? mono(row.posix_group) : "—"),
    },
    {
      key: "validity",
      header: "Validity",
      width: "13rem",
      cell: (row) => (
        <span className="whitespace-nowrap">
          {day(row.start_time)} – {day(row.end_time)}
        </span>
      ),
    },
    ...retire,
  ];
  return clusterId ? (
    <Listing
      list="cluster-allocations"
      path={`/compute-clusters/${clusterId}/allocations`}
      parse={(data) => z.array(zComputeAllocation).parse(data ?? [])}
      columns={columns}
      rowKey={(row) => row.id}
      rowClassName={(row) => (row.status === "ACTIVE" ? undefined : DISPOSABLE)}
      empty="No allocations on this cluster."
    />
  ) : null;
}
