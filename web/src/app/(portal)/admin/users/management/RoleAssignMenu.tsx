"use client";

import { Pencil } from "lucide-react";
import { useState } from "react";
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
import { PermissionRW } from "@/shared/users-admin/PermissionRW";
import { rwStateFor } from "@/shared/users-admin/permissions";
import type { RoleRow } from "@/shared/users-admin/types";

function setsEqual(a: Set<string>, b: Set<string>) {
  if (a.size !== b.size) return false;
  for (const value of a) {
    if (!b.has(value)) return false;
  }
  return true;
}

export function RoleAssignMenu({
  roles,
  heldRoleIds,
  onToggleRole,
  triggerLabel,
}: {
  roles: RoleRow[];
  heldRoleIds: Set<string>;
  onToggleRole: (roleId: string) => void;
  triggerLabel: string;
}) {
  const [open, setOpen] = useState(false);
  const [draftIds, setDraftIds] = useState<Set<string>>(new Set());

  function handleOpenChange(next: boolean) {
    setOpen(next);
    if (next) setDraftIds(new Set(heldRoleIds));
  }

  function toggleDraft(roleId: string) {
    setDraftIds((prev) => {
      const next = new Set(prev);
      if (next.has(roleId)) next.delete(roleId);
      else next.add(roleId);
      return next;
    });
  }

  function handleSave() {
    for (const role of roles) {
      if (role.id && draftIds.has(role.id) !== heldRoleIds.has(role.id)) {
        onToggleRole(role.id);
      }
    }
    setOpen(false);
  }

  const hasChanges = !setsEqual(draftIds, heldRoleIds);

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger
        render={
          <button
            type="button"
            className="inline-flex items-center gap-1.5 text-sm font-medium text-muted-foreground hover:text-foreground"
          />
        }
      >
        <Pencil className="size-3.5" />
        Edit roles
      </DialogTrigger>
      <DialogContent className="gap-5 sm:max-w-2xl">
        <DialogHeader className="-mx-4 -mb-5 gap-1 border-b border-border px-4 pt-2 pb-4">
          <DialogTitle className="text-lg font-semibold">Manage roles</DialogTitle>
          <DialogDescription>{triggerLabel}</DialogDescription>
        </DialogHeader>

        <div className="-mx-4 max-h-[28rem] overflow-y-auto">
          <ul className="space-y-3 px-6 pt-3 pb-4">
            {roles.map((role) => {
              const assigned = role.id ? draftIds.has(role.id) : false;
              const rwPermissions = rwStateFor(role.permissions).filter((p) => p.read || p.write);

              return (
                <li key={role.id} className="overflow-hidden rounded-lg border border-border">
                  <div className="flex items-start justify-between gap-4 bg-muted/60 px-4 py-3">
                    <div>
                      <p className="font-heading text-base font-semibold text-foreground">
                        {role.name}
                      </p>
                      {role.description && (
                        <p className="mt-1 text-sm text-muted-foreground">{role.description}</p>
                      )}
                    </div>
                    <Button
                      type="button"
                      variant={assigned ? "secondary" : "default"}
                      size="sm"
                      className="shrink-0"
                      onClick={() => role.id && toggleDraft(role.id)}
                    >
                      {assigned ? "Unassign" : "Assign"}
                    </Button>
                  </div>

                  <div className="p-4">
                    {rwPermissions.length > 0 ? (
                      <ul className="space-y-2">
                        {rwPermissions.map((p) => (
                          <li
                            key={p.section}
                            className="flex items-center justify-between text-sm"
                          >
                            <span className="font-mono text-foreground">{p.section}</span>
                            <PermissionRW read={p.read} write={p.write} />
                          </li>
                        ))}
                      </ul>
                    ) : (
                      <p className="text-sm text-muted-foreground">No privileges granted.</p>
                    )}
                  </div>
                </li>
              );
            })}
          </ul>
        </div>

        <DialogFooter className="-mt-5">
          <Button variant="outline" onClick={() => setOpen(false)} type="button">
            Cancel
          </Button>
          <Button onClick={handleSave} type="button" disabled={!hasChanges}>
            Save
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
