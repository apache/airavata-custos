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
import type { ComputeAllocation } from "@/features/core/allocations/schemas";
import type { CreateUserPayload } from "@/features/core/users/schemas";
import { cn } from "@/lib/utils";
import { Button } from "@/shared/ui/button";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/shared/ui/dialog";
import { Input } from "@/shared/ui/input";
import { Label } from "@/shared/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/shared/ui/select";
import { ShieldCheck } from "lucide-react";

const NO_ALLOCATION = "none";

export type AddUserDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (payload: CreateUserPayload) => void;
  isPending: boolean;
  error?: string | null;
  canCreateAdmins: boolean;
  allocations: ComputeAllocation[];
};

export function AddUserDialog({
  open,
  onOpenChange,
  onSubmit,
  isPending,
  error,
  canCreateAdmins,
  allocations,
}: AddUserDialogProps) {
  const [userType, setUserType] = React.useState<"researcher" | "admin">("researcher");
  const [email, setEmail] = React.useState("");
  const [firstName, setFirstName] = React.useState("");
  const [lastName, setLastName] = React.useState("");
  const [username, setUsername] = React.useState("");
  const [allocationId, setAllocationId] = React.useState(NO_ALLOCATION);
  const [clusterAdmin, setClusterAdmin] = React.useState(false);

  React.useEffect(() => {
    if (!open) {
      setUserType("researcher");
      setEmail("");
      setFirstName("");
      setLastName("");
      setUsername("");
      setAllocationId(NO_ALLOCATION);
      setClusterAdmin(false);
    }
  }, [open]);

  const isAdmin = userType === "admin";

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (isPending) return;
    const payload: CreateUserPayload = {
      email: email.trim(),
      first_name: firstName.trim(),
      last_name: lastName.trim(),
    };
    if (isAdmin) {
      payload.portal_admin = true;
      if (clusterAdmin) payload.cluster_admin = true;
    } else {
      if (username.trim()) payload.username = username.trim();
      if (allocationId !== NO_ALLOCATION) {
        payload.allocation_id = allocationId;
      }
    }
    onSubmit(payload);
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>Add user</DialogTitle>
        </DialogHeader>
        <form className="space-y-4" onSubmit={handleSubmit}>
          <div className="space-y-2">
            <Label>User type</Label>
            <div className="grid grid-cols-2 gap-2" role="radiogroup" aria-label="User type">
              <TypeOption
                name="Researcher"
                description="Runs jobs under an allocation"
                selected={userType === "researcher"}
                onSelect={() => setUserType("researcher")}
              />
              <TypeOption
                name="Admin"
                description="Manages the portal, the clusters, or both"
                selected={userType === "admin"}
                onSelect={() => setUserType("admin")}
                disabled={!canCreateAdmins}
                disabledReason="Only super admins can create admins"
              />
            </div>
          </div>
          <div className="space-y-2">
            <Label htmlFor="add-user-email">Email</Label>
            <Input
              id="add-user-email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="user@university.edu"
              autoComplete="off"
              required
            />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-2">
              <Label htmlFor="add-user-first">First name</Label>
              <Input
                id="add-user-first"
                value={firstName}
                onChange={(e) => setFirstName(e.target.value)}
                autoComplete="off"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="add-user-last">Last name</Label>
              <Input
                id="add-user-last"
                value={lastName}
                onChange={(e) => setLastName(e.target.value)}
                autoComplete="off"
              />
            </div>
          </div>
          {!isAdmin ? (
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <Label htmlFor="add-user-username">Username</Label>
                <span className="text-xs text-muted-foreground">Optional</span>
              </div>
              <Input
                id="add-user-username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="Leave empty to auto-generate"
                autoComplete="off"
                spellCheck={false}
              />
              <p className="text-xs text-muted-foreground">
                Generated from the name when left empty, the same way ingested accounts get
                theirs.
              </p>
            </div>
          ) : null}
          {!isAdmin ? (
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <Label>Allocation</Label>
                <span className="text-xs text-muted-foreground">Optional</span>
              </div>
              <Select
                value={allocationId}
                onValueChange={(value) => setAllocationId(value ?? NO_ALLOCATION)}
              >
                <SelectTrigger aria-label="Allocation" className="w-full">
                  <SelectValue>
                    {(value: string) =>
                      value === NO_ALLOCATION
                        ? "No allocation, assign later"
                        : (allocations.find((a) => a.id === value)?.name ?? value)
                    }
                  </SelectValue>
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value={NO_ALLOCATION}>No allocation, assign later</SelectItem>
                  {allocations.map((allocation) => (
                    <SelectItem key={allocation.id} value={allocation.id}>
                      {allocation.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <p className="text-xs text-muted-foreground">
                Can be assigned later, but needed before they can run jobs.
              </p>
            </div>
          ) : (
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <Label>Admin access</Label>
              </div>
              <div className="rounded-md border">
                <label className="flex cursor-pointer items-start gap-2.5 p-3">
                  <input
                    type="checkbox"
                    checked
                    readOnly
                    className="mt-0.5 size-4 rounded border-input accent-[color:var(--brand)]"
                  />
                  <span>
                    <span className="block text-sm font-medium">Portal admin</span>
                    <span className="block text-xs text-muted-foreground">
                      Manage users, allocations, and settings in this portal.
                    </span>
                  </span>
                </label>
                <label className="flex cursor-pointer items-start gap-2.5 border-t p-3">
                  <input
                    type="checkbox"
                    checked={clusterAdmin}
                    onChange={(e) => setClusterAdmin(e.target.checked)}
                    className="mt-0.5 size-4 rounded border-input accent-[color:var(--brand)]"
                  />
                  <span>
                    <span className="block text-sm font-medium">Cluster admin</span>
                    <span className="block text-xs text-muted-foreground">
                      Admin access on the cluster itself, with a login account.
                    </span>
                  </span>
                </label>
              </div>
              <div className="flex items-start gap-2.5 rounded-md border border-[color:var(--tone-warn-fg)]/30 bg-[color:var(--tone-warn-bg)] p-3 text-xs">
                <ShieldCheck className="mt-0.5 size-4 shrink-0 text-[color:var(--tone-warn-fg)]" />
                <span>
                  <span className="block text-sm font-medium">
                    {clusterAdmin
                      ? "This user will be a portal and cluster admin"
                      : "This user will be a portal admin"}
                  </span>
                  <span className="text-muted-foreground">
                    {clusterAdmin
                      ? "Admin access to the portal, and sudo on the cluster."
                      : "Admin access to the portal."}
                  </span>
                </span>
              </div>
            </div>
          )}
          {error ? <p className="text-sm text-[color:var(--custos-red-700)]">{error}</p> : null}
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" variant="brand" disabled={isPending}>
              {isPending ? "Adding…" : isAdmin ? "Add admin user" : "Add user"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

function TypeOption({
  name,
  description,
  selected,
  onSelect,
  disabled,
  disabledReason,
}: {
  name: string;
  description: string;
  selected: boolean;
  onSelect: () => void;
  disabled?: boolean;
  disabledReason?: string;
}) {
  return (
    <label
      title={disabled ? disabledReason : undefined}
      className={cn(
        "rounded-md border p-3 transition-colors",
        selected ? "border-[color:var(--brand)] bg-[color:var(--brand-tint)]/40" : "hover:bg-muted",
        disabled ? "cursor-not-allowed opacity-50 hover:bg-transparent" : "cursor-pointer",
      )}
    >
      <input
        type="radio"
        name="add-user-type"
        checked={selected}
        onChange={onSelect}
        disabled={disabled}
        className="sr-only"
      />
      <span className="block text-sm font-medium">{name}</span>
      <span className="block text-xs text-muted-foreground">{description}</span>
    </label>
  );
}
