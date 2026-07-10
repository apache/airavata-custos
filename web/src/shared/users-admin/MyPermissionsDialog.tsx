"use client";

import { useCurrentUser } from "@/features/core/identity/queries";
import { Badge } from "@/shared/ui/badge";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/shared/ui/dialog";
import { PermissionRW } from "./PermissionRW";
import { permissionsFromRoles, rwStateFor } from "./permissions";
import { useUsersAdmin } from "./UsersAdminContext";

export function MyPermissionsDialog({
  open,
  onOpenChange,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const { user: currentUser } = useCurrentUser();
  const { users } = useUsersAdmin();
  const me = users.find((u) => u.email === currentUser?.email) ?? null;
  const rwPermissions = rwStateFor(me ? permissionsFromRoles(me.roles) : []);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>My Permissions</DialogTitle>
        </DialogHeader>

        <div className="space-y-5">
          <section>
            <h3 className="mb-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
              Roles
            </h3>
            <div className="flex flex-wrap gap-1.5">
              {!me || me.roles.length === 0 ? (
                <span className="text-sm text-muted-foreground">No roles</span>
              ) : (
                me.roles.map((role) => (
                  <Badge key={role.id} variant="outline">
                    {role.name}
                  </Badge>
                ))
              )}
            </div>
          </section>

          <section>
            <h3 className="mb-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
              Permissions
            </h3>
            <ul className="space-y-2">
              {rwPermissions.map((p) => (
                <li key={p.section} className="flex items-center justify-between text-sm">
                  <span className="font-mono text-foreground">{p.section}</span>
                  <PermissionRW read={p.read} write={p.write} />
                </li>
              ))}
            </ul>
          </section>
        </div>
      </DialogContent>
    </Dialog>
  );
}
