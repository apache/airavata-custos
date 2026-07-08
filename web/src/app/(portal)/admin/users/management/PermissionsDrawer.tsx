import { Badge } from "@/shared/ui/badge";
import { SideDrawer } from "@/shared/ui/SideDrawer";
import { PermissionRW } from "../PermissionRW";
import { permissionsFromRoles, rwStateFor } from "../permissions";
import type { UserRow } from "../types";
import { useUsersAdmin } from "../UsersAdminContext";
import { identitySourceIcon, identitySourceLabel } from "./identities";
import { RoleAssignMenu } from "./RoleAssignMenu";

export function PermissionsDrawer({
  user,
  onClose,
}: {
  user: UserRow | null;
  onClose: () => void;
}) {
  const { roles, toggleUserRole } = useUsersAdmin();
  const rwPermissions = rwStateFor(user ? permissionsFromRoles(user.roles) : []);
  const heldRoleIds = new Set(user?.roles.map((r) => r.id ?? "") ?? []);

  return (
    <SideDrawer
      open={user !== null}
      onOpenChange={(open) => {
        if (!open) onClose();
      }}
      title={user ? [user.first_name, user.last_name].filter(Boolean).join(" ") : undefined}
      description={user?.email}
      width="sm"
      modal={false}
      disablePointerDismissal
    >
      {user && (
        <div className="space-y-5">
          <section>
            <h3 className="mb-3 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
              Roles
            </h3>
            <div className="flex flex-wrap items-center gap-1.5">
              {user.roles.length === 0 ? (
                <span className="text-sm text-muted-foreground">No roles</span>
              ) : (
                user.roles.map((role) => (
                  <Badge key={role.id} variant="outline">
                    {role.name}
                  </Badge>
                ))
              )}
              <RoleAssignMenu
                roles={roles}
                heldRoleIds={heldRoleIds}
                onToggleRole={(roleId) => user.id && toggleUserRole(user.id, roleId)}
                triggerLabel={`Assign a role to ${user.email ?? "this user"}`}
              />
            </div>
          </section>

          <div className="border-t border-border" />

          <section>
            <h3 className="mb-3 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
              External Identities
            </h3>
            <div className="flex flex-wrap items-center gap-1.5">
              {user.identities.length === 0 ? (
                <span className="text-sm text-muted-foreground">No identities</span>
              ) : (
                user.identities.map((identity) => {
                  const Icon = identitySourceIcon(identity.source);
                  return (
                    <Badge key={identity.id} variant="outline">
                      <Icon data-icon="inline-start" />
                      {identitySourceLabel(identity.source)}
                    </Badge>
                  );
                })
              )}
            </div>
          </section>

          <div className="border-t border-border" />

          <section>
            <h3 className="mb-3 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
              Permissions
            </h3>
            <ul className="space-y-2">
              {rwPermissions.map((p) => (
                <li key={p.section} className="flex items-center justify-between text-sm">
                  <span className="text-foreground">{p.label}</span>
                  <PermissionRW read={p.read} write={p.write} />
                </li>
              ))}
            </ul>
          </section>
        </div>
      )}
    </SideDrawer>
  );
}
