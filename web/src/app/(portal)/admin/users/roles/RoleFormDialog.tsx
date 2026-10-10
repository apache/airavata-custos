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
import {
  RoleSaveError,
  type RoleRow,
  useCreateRole,
  useDeleteRole,
  usePrivilegeCatalog,
  useUpdateRole,
} from "@/features/core/roles/queries";
import { UserPicker } from "@/features/core/users/components/UserPicker";
import type { PrivilegeKey } from "@/generated/core/types.gen";
import { useAbility } from "@/shared/casl/AbilityProvider";
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
import { confirmToast, toastOnSuccess } from "@/shared/ui/sonner";
import { toggleId, togglePermission } from "@/shared/users-admin/permissions";
import { PermissionMatrixEditor } from "./PermissionMatrixEditor";

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
  const ability = useAbility();
  const canGrant = ability.can("write", "PrivilegeGrant");
  const canReadUsers = ability.can("read", "User");
  const createRole = useCreateRole();
  const updateRole = useUpdateRole();
  const deleteRole = useDeleteRole();
  const [open, setOpen] = React.useState(false);
  // Becomes the created role when a create saves only partially.
  const [base, setBase] = React.useState(role);
  const catalogQuery = usePrivilegeCatalog(open && canGrant);
  const [name, setName] = React.useState("");
  const [description, setDescription] = React.useState("");
  const [permissions, setPermissions] = React.useState<PrivilegeKey[]>([]);
  const [selectedUserIds, setSelectedUserIds] = React.useState<Set<string>>(new Set());

  function handleOpenChange(next: boolean) {
    setOpen(next);
    if (next) {
      setBase(role);
      setName(role?.name ?? "");
      setDescription(role?.description ?? "");
      setPermissions(role?.privileges ?? []);
      setSelectedUserIds(new Set(role?.holderIds ?? []));
    }
  }

  function handleSubmit() {
    const input = {
      name: name.trim(),
      description: description.trim(),
      privileges: permissions,
      memberUserIds: Array.from(selectedUserIds),
    };
    const callbacks = (message: string) => ({
      ...toastOnSuccess(message, () => setOpen(false)),
      onError: (err: Error) => {
        if (err instanceof RoleSaveError) setBase(err.role);
      },
    });
    if (base) updateRole.mutate({ role: base, input }, callbacks("Role updated"));
    else createRole.mutate(input, callbacks("Role created"));
  }

  const isEdit = Boolean(base);
  const saving = createRole.isPending || updateRole.isPending;
  const catalog = catalogQuery.data ?? [];

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger render={triggerRender}>{triggerContent}</DialogTrigger>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{isEdit ? "Edit role" : "Create role"}</DialogTitle>
          <DialogDescription>
            {isEdit
              ? `Update what ${base?.name} can see and do.`
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
                disabled={base?.is_system}
                title={base?.is_system ? "System roles cannot be renamed" : undefined}
                autoFocus
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="role-description">Description</Label>
              <Input
                id="role-description"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder={
                  base?.description
                    ? "Leave blank to keep the current description"
                    : "What this role is for"
                }
              />
            </div>
          </div>

          <div className="border-t border-border" />

          {catalogQuery.isSuccess ? (
            <PermissionMatrixEditor
              permissions={permissions}
              catalog={catalog}
              onTogglePermission={(key) =>
                setPermissions((prev) => togglePermission(prev, key, catalog))
              }
            />
          ) : (
            <div className="space-y-2">
              <PermissionMatrixEditor permissions={permissions} editable={false} />
              <p className="text-xs text-muted-foreground">
                {catalogQuery.isError
                  ? `Could not load the privilege catalog: ${catalogQuery.error.message}`
                  : canGrant
                    ? "Loading the privilege catalog…"
                    : "Changing privileges needs core:privileges:grant, which lists the catalog."}
              </p>
            </div>
          )}

          {canReadUsers ? (
            <>
              <div className="border-t border-border" />
              <UserPicker
                id="role-user-search"
                label="Assign to users (optional)"
                enabled={open}
                selected={selectedUserIds}
                onToggle={(userId) => setSelectedUserIds((prev) => toggleId(prev, userId))}
              />
            </>
          ) : null}
        </div>

        <DialogFooter>
          {base && !base.is_system ? (
            <Button
              variant="destructive"
              className="mr-auto"
              type="button"
              disabled={saving || deleteRole.isPending}
              onClick={() =>
                confirmToast(
                  `Delete role "${base.name}"? Holders lose its privileges.`,
                  "Delete",
                  () =>
                    deleteRole.mutate(
                      base.id,
                      toastOnSuccess("Role deleted", () => setOpen(false)),
                    ),
                )
              }
            >
              Delete role
            </Button>
          ) : null}
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
