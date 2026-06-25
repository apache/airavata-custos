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
import type { AllocationMembership, AllocationStatus } from "../schemas";

const ROLES = ["PI", "CO_PI", "ALLOCATION_MANAGER", "MEMBER"] as const;
type MemberRole = (typeof ROLES)[number];

const ROLE_LABELS: Record<MemberRole, string> = {
  PI: "PI",
  CO_PI: "Co-PI",
  ALLOCATION_MANAGER: "Allocation Manager",
  MEMBER: "Member",
};

export type MemberEditPayload = {
  role: MemberRole;
  membership_status: AllocationStatus;
  start_time?: string;
  end_time?: string;
};

export type MemberEditDialogProps = {
  member: AllocationMembership | null;
  onClose: () => void;
  onSubmit: (payload: MemberEditPayload) => void;
  isPending: boolean;
};

export function MemberEditDialog({ member, onClose, onSubmit, isPending }: MemberEditDialogProps) {
  const [role, setRole] = React.useState<MemberRole>("MEMBER");
  const [status, setStatus] = React.useState<AllocationStatus>("ACTIVE");
  const [endTime, setEndTime] = React.useState("");

  React.useEffect(() => {
    if (member) {
      setRole((member.role as MemberRole | undefined) ?? "MEMBER");
      setStatus(member.membership_status);
      setEndTime(member.end_time.slice(0, 10));
    }
  }, [member]);

  return (
    <Dialog open={Boolean(member)} onOpenChange={(open) => !open && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Edit member</DialogTitle>
          <DialogDescription>{member?.display_name ?? member?.user_id}</DialogDescription>
        </DialogHeader>
        <form
          className="space-y-4"
          onSubmit={(e) => {
            e.preventDefault();
            onSubmit({
              role,
              membership_status: status,
              ...(endTime ? { end_time: new Date(endTime).toISOString() } : {}),
            });
          }}
        >
          <div className="space-y-2">
            <Label htmlFor="member-role">Role</Label>
            <select
              id="member-role"
              value={role}
              onChange={(e) => setRole(e.target.value as MemberRole)}
              className="h-9 w-full rounded-md border bg-background px-3 text-sm"
            >
              {ROLES.map((r) => (
                <option key={r} value={r}>
                  {ROLE_LABELS[r]}
                </option>
              ))}
            </select>
          </div>
          <div className="space-y-2">
            <Label htmlFor="member-status">Membership status</Label>
            <select
              id="member-status"
              value={status}
              onChange={(e) => setStatus(e.target.value as AllocationStatus)}
              className="h-9 w-full rounded-md border bg-background px-3 text-sm"
            >
              <option value="ACTIVE">ACTIVE</option>
              <option value="INACTIVE">INACTIVE</option>
              <option value="DELETED">DELETED</option>
            </select>
          </div>
          <div className="space-y-2">
            <Label htmlFor="member-end">End date</Label>
            <Input
              id="member-end"
              type="date"
              value={endTime}
              onChange={(e) => setEndTime(e.target.value)}
            />
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={onClose}>
              Cancel
            </Button>
            <Button type="submit" disabled={isPending}>
              {isPending ? "Saving…" : "Save"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
