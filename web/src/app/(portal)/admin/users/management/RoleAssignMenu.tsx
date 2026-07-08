"use client";

import { Plus } from "lucide-react";
import { Button } from "@/shared/ui/button";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuLabel,
  DropdownMenuTrigger,
} from "@/shared/ui/dropdown-menu";
import type { RoleRow } from "../types";

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
  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        render={
          <Button variant="ghost" size="icon-xs" className="size-5" aria-label={triggerLabel} />
        }
      >
        <Plus />
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="min-w-44">
        <DropdownMenuGroup>
          <DropdownMenuLabel>Assign roles</DropdownMenuLabel>
          {roles.map((role) => (
            <DropdownMenuCheckboxItem
              key={role.id}
              checked={role.id ? heldRoleIds.has(role.id) : false}
              onCheckedChange={() => role.id && onToggleRole(role.id)}
            >
              {role.name}
            </DropdownMenuCheckboxItem>
          ))}
        </DropdownMenuGroup>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
