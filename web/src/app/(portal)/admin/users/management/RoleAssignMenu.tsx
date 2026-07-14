"use client";

import { useRoleDetails } from "@/features/core/users/queries";
import type { Role } from "@/features/core/users/schemas";
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
import { Pencil } from "lucide-react";
import { useState } from "react";
import { PrivilegeList } from "./PrivilegeList";

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
  onSave,
  triggerLabel,
  isCurrentUser,
  isPending,
  error,
}: {
  roles: Role[];
  heldRoleIds: Set<string>;
  onSave: (roleIds: string[]) => Promise<boolean>;
  triggerLabel: string;
  isCurrentUser: boolean;
  isPending: boolean;
  error: string | null;
}) {
  const [open, setOpen] = useState(false);
  const [draftIds, setDraftIds] = useState<Set<string>>(new Set());
  const details = useRoleDetails(
    roles.flatMap((role) => (role.id ? [role.id] : [])),
    open,
  );
  const detailById = new Map(details.roles.map((role) => [role.id, role]));

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

  async function handleSave() {
    const removedIds = [...heldRoleIds].filter((roleId) => !draftIds.has(roleId));
    const removesOwnRoleManager =
      isCurrentUser &&
      removedIds.some((roleId) => detailById.get(roleId)?.privileges.includes("core:roles:manage"));
    const confirmationMessage = removesOwnRoleManager
      ? "This may remove your own ability to manage roles. Continue with these changes?"
      : isCurrentUser && removedIds.length > 0 && details.isError
        ? "Some role privileges are unavailable, so these changes may remove your own access. Continue?"
        : null;
    if (confirmationMessage && !window.confirm(confirmationMessage)) {
      return;
    }
    if (await onSave([...draftIds])) setOpen(false);
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
              const privileges = role.id ? (detailById.get(role.id)?.privileges ?? []) : [];

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
                      disabled={isPending}
                    >
                      {assigned ? "Unassign" : "Assign"}
                    </Button>
                  </div>

                  <div className="p-4">
                    {details.isLoading && privileges.length === 0 ? (
                      <p className="text-sm text-muted-foreground">Loading privileges…</p>
                    ) : details.isError && privileges.length === 0 ? (
                      <p className="text-sm text-muted-foreground">Privileges unavailable.</p>
                    ) : (
                      <PrivilegeList privileges={privileges} />
                    )}
                  </div>
                </li>
              );
            })}
          </ul>
        </div>

        {error ? (
          <p role="alert" className="text-sm text-[color:var(--custos-red-600)]">
            {error}
          </p>
        ) : null}
        <DialogFooter className="-mt-5">
          <Button
            variant="outline"
            onClick={() => setOpen(false)}
            type="button"
            disabled={isPending}
          >
            Cancel
          </Button>
          <Button
            onClick={() => void handleSave()}
            type="button"
            disabled={!hasChanges || isPending || details.isLoading}
          >
            {isPending ? "Saving…" : "Save"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
