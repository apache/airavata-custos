import type { Role, User, UserIdentity } from "@/generated/core/types.gen";
import type { PermissionKey } from "./permissions";

export type RoleRow = Role & { permissions: PermissionKey[] };

// Shaped to match GET /users/{id} (User), joined with the roles resolved
// from GET /users/{id}/roles + GET /roles, and the linked accounts from
// GET /users/{id}/user-identities — a stand-in for a future "list users"
// endpoint that returns this view directly. Effective permissions are
// derived from the union of the user's roles rather than tracked per-user.
export type UserRow = User & {
  roles: RoleRow[];
  identities: UserIdentity[];
};
