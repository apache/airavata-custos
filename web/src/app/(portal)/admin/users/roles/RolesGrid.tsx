"use client";

import { useUsersAdmin } from "@/shared/users-admin/UsersAdminContext";
import { RoleCard } from "./RoleCard";

export function RolesGrid() {
  const { roles, users } = useUsersAdmin();

  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
      {roles.map((role) => (
        <RoleCard
          key={role.id}
          role={role}
          memberCount={users.filter((u) => u.roles.some((r) => r.id === role.id)).length}
        />
      ))}
    </div>
  );
}
