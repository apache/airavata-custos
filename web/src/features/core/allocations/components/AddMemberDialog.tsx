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
} from "@/shared/ui/dialog";
import { Input } from "@/shared/ui/input";
import { Label } from "@/shared/ui/label";
import type { CreateMembershipPayload } from "../schemas";

const ROLES = ["PI", "CO_PI", "ALLOCATION_MANAGER", "MEMBER"] as const;
type Role = (typeof ROLES)[number];

const ROLE_LABELS: Record<Role, string> = {
  PI: "PI",
  CO_PI: "Co-PI",
  ALLOCATION_MANAGER: "Allocation Manager",
  MEMBER: "Member",
};

export type AddMemberDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  allocationId: string;
  defaultEndTime: string;
  onSubmit: (payload: CreateMembershipPayload) => void;
  isPending: boolean;
};

export function AddMemberDialog({
  open,
  onOpenChange,
  allocationId,
  defaultEndTime,
  onSubmit,
  isPending,
}: AddMemberDialogProps) {
  const [userId, setUserId] = React.useState("");
  const [role, setRole] = React.useState<Role>("MEMBER");

  React.useEffect(() => {
    if (!open) {
      setUserId("");
      setRole("MEMBER");
    }
  }, [open]);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Add member</DialogTitle>
          <DialogDescription>
            Add a user to this allocation. They will see it under their allocations.
          </DialogDescription>
        </DialogHeader>
        <form
          className="space-y-4"
          onSubmit={(e) => {
            e.preventDefault();
            if (!userId.trim()) return;
            onSubmit({
              compute_allocation_id: allocationId,
              user_id: userId.trim(),
              start_time: new Date().toISOString(),
              end_time: defaultEndTime,
              membership_status: "ACTIVE",
              role,
            });
          }}
        >
          <div className="space-y-2">
            <Label htmlFor="add-member-user">User ID</Label>
            <Input
              id="add-member-user"
              value={userId}
              onChange={(e) => setUserId(e.target.value)}
              placeholder="user-123"
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="add-member-role">Role</Label>
            <select
              id="add-member-role"
              value={role}
              onChange={(e) => setRole(e.target.value as Role)}
              className="h-9 w-full rounded-md border bg-background px-3 text-sm"
            >
              {ROLES.map((r) => (
                <option key={r} value={r}>
                  {ROLE_LABELS[r]}
                </option>
              ))}
            </select>
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={isPending || !userId.trim()}>
              {isPending ? "Adding…" : "Add member"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
