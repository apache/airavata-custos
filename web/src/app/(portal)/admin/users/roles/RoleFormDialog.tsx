"use client";

import * as React from "react";
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
import { togglePermission } from "../permissions";
import type { PermissionKey } from "../permissions";
import type { RoleRow, UserRow } from "../types";
import { useUsersAdmin } from "../UsersAdminContext";
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
  const { users, addRole, updateRole } = useUsersAdmin();
  const isEdit = Boolean(role);
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
      setPermissions(role?.permissions ?? []);
      setUserSearch("");
      setSelectedUserIds(
        new Set(
          role ? users.filter((u) => u.roles.some((r) => r.id === role.id)).map((u) => u.id ?? "") : [],
        ),
      );
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

  function handleSubmit() {
    if (!name.trim()) return;
    const input = { name: name.trim(), description: description.trim(), permissions };
    if (isEdit && role?.id) {
      updateRole(role.id, input, Array.from(selectedUserIds));
    } else {
      addRole(input, Array.from(selectedUserIds));
    }
    setOpen(false);
  }

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
              ? `Update what ${role?.name} can see and do, and who holds it.`
              : "Define a name, choose the permissions it grants, and optionally assign it to users right away."}
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
          <Button variant="outline" onClick={() => setOpen(false)} type="button">
            Cancel
          </Button>
          <Button onClick={handleSubmit} disabled={!name.trim()} type="button">
            {isEdit ? "Save changes" : "Create role"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
