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
import { useCreateRole, usePrivilegeCatalog, useUpdateRole } from "@/features/core/roles/queries";
import type { RoleRow } from "@/features/core/roles/schemas";
import { Button } from "@/shared/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/shared/ui/dialog";
import { Input } from "@/shared/ui/input";
import { Label } from "@/shared/ui/label";
import { togglePermission } from "@/shared/users-admin/permissions";
import type { PermissionKey } from "@/shared/users-admin/permissions";
import { useUsersAdmin } from "@/shared/users-admin/UsersAdminContext";
import type { UserRow } from "@/shared/users-admin/types";
import { PermissionMatrixEditor } from "./PermissionMatrixEditor";

function fullNameFor(user: UserRow): string {
  const name = [user.first_name, user.last_name].filter(Boolean).join(" ");
  return name || (user.email ?? "Unknown user");
}

export function RoleFormDialog({
  role,
  triggerRender,
  triggerContent,
}: {
  // Omit for "create a new role"; pass an existing role to edit it in place.
  role?: RoleRow;
  triggerRender: React.ReactElement;
  triggerContent: React.ReactNode;
}) {
  const isEdit = Boolean(role);
  const catalogQuery = usePrivilegeCatalog();
  const createRole = useCreateRole();
  const updateRole = useUpdateRole();
  const { users } = useUsersAdmin();
  const [open, setOpen] = React.useState(false);
  const [name, setName] = React.useState("");
  const [description, setDescription] = React.useState("");
  const [permissions, setPermissions] = React.useState<PermissionKey[]>([]);
  const [userSearch, setUserSearch] = React.useState("");
  const [selectedUserIds, setSelectedUserIds] = React.useState<Set<string>>(new Set());

  function handleOpenChange(next: boolean) {
    setOpen(next);
    if (next) {
      setName(role?.name ?? "");
      setDescription(role?.description ?? "");
      setPermissions(role?.privileges ?? []);
      setUserSearch("");
      setSelectedUserIds(new Set(role?.holderIds ?? []));
    }
  }

  function toggleUserSelected(userId: string) {
    setSelectedUserIds((prev) => {
      const next = new Set(prev);
      if (next.has(userId)) next.delete(userId);
      else next.add(userId);
      return next;
    });
  }

  async function handleSubmit() {
    if (!name.trim()) return;
    const input = { name: name.trim(), description: description.trim(), privileges: permissions };
    const memberUserIds = Array.from(selectedUserIds);
    try {
      if (isEdit && role?.id) {
        await updateRole.mutateAsync({ role, input, memberUserIds });
        toast.success("Role updated");
      } else {
        await createRole.mutateAsync({ ...input, memberUserIds });
        toast.success("Role created");
      }
      setOpen(false);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Role update failed");
    }
  }

  const saving = createRole.isPending || updateRole.isPending;
  const catalog = catalogQuery.data ?? [];
  const needle = userSearch.trim().toLowerCase();
  const matchingUsers = needle
    ? users.filter((u) => `${fullNameFor(u)} ${u.email ?? ""}`.toLowerCase().includes(needle))
    : users;

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger render={triggerRender}>{triggerContent}</DialogTrigger>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{isEdit ? "Edit role" : "Create role"}</DialogTitle>
          <DialogDescription>
            {isEdit
              ? `Update what ${role?.name} can see and do.`
              : "Define a name and choose the permissions it grants."}
          </DialogDescription>
        </DialogHeader>

        <div className="max-h-[60vh] space-y-5 overflow-y-auto">
          <div className="space-y-3">
            <div className="space-y-1.5">
              <Label htmlFor="role-name">Name</Label>
              <Input
                id="role-name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g. Billing Reviewer"
                autoFocus
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="role-description">Description</Label>
              <Input
                id="role-description"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="What this role is for"
              />
            </div>
          </div>

          <div className="border-t border-border" />

          <PermissionMatrixEditor
            permissions={permissions}
            catalog={catalog}
            onTogglePermission={(key) => setPermissions((prev) => togglePermission(prev, key))}
          />

          <div className="border-t border-border" />

          <div className="space-y-2">
            <Label htmlFor="role-user-search">Assign to users (optional)</Label>
            <Input
              id="role-user-search"
              type="search"
              value={userSearch}
              onChange={(e) => setUserSearch(e.target.value)}
              placeholder="Search by username or email"
            />
            <ul className="max-h-40 space-y-1 overflow-y-auto rounded-md border p-2">
              {matchingUsers.length === 0 ? (
                <li className="px-1 py-1 text-sm text-muted-foreground">No users match.</li>
              ) : (
                matchingUsers.map((u) => {
                  const id = u.id ?? u.email ?? "";
                  return (
                    <li key={id}>
                      <label className="flex cursor-pointer items-center gap-2 rounded-sm px-1 py-1 text-sm hover:bg-muted">
                        <input
                          type="checkbox"
                          checked={selectedUserIds.has(id)}
                          onChange={() => toggleUserSelected(id)}
                          className="size-4 rounded border-input"
                        />
                        <span className="font-medium text-foreground">{fullNameFor(u)}</span>
                        <span className="text-xs text-muted-foreground">{u.email}</span>
                      </label>
                    </li>
                  );
                })
              )}
            </ul>
            {selectedUserIds.size > 0 ? (
              <p className="text-xs text-muted-foreground">
                {selectedUserIds.size} user{selectedUserIds.size === 1 ? "" : "s"} selected
              </p>
            ) : null}
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => setOpen(false)} type="button" disabled={saving}>
            Cancel
          </Button>
          <Button onClick={handleSubmit} disabled={!name.trim() || saving} type="button">
            {saving ? "Saving..." : isEdit ? "Save changes" : "Create role"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
