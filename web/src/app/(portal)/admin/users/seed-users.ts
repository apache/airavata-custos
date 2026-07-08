import type { UserIdentity } from "@/generated/core/types.gen";
import { ROLE_AUDITOR, ROLE_OPERATOR, ROLE_SUPER_ADMIN } from "./roles-catalog";
import type { UserRow } from "./types";

function identity(
  user_id: string,
  source: string,
  external_id: string,
  email?: string,
): UserIdentity {
  return {
    id: `${user_id}-${source}`,
    user_id,
    source,
    external_id,
    email,
    created_at: "2026-01-15T09:00:00Z",
  };
}

export const INITIAL_USERS: UserRow[] = [
  {
    id: "u1",
    first_name: "Dev",
    last_name: "Admin",
    email: "admin@custos.local",
    status: "ACTIVE",
    roles: [ROLE_SUPER_ADMIN, ROLE_OPERATOR, ROLE_AUDITOR],
    identities: [
      identity("u1", "access", "dev-admin-access-001", "admin@access-ci.org"),
      identity("u1", "cilogon", "cilogon-uid-1001", "admin@cilogon.org"),
      identity("u1", "orcid", "0000-0001-1111-0001"),
      identity("u1", "nairr", "nairr-uid-1001"),
    ],
  },
  {
    id: "u2",
    first_name: "Rachel",
    last_name: "Gao",
    email: "rgao@access-ci.org",
    status: "ACTIVE",
    roles: [ROLE_OPERATOR],
    identities: [
      identity("u2", "access", "rgao-access-001", "rgao@access-ci.org"),
      identity("u2", "orcid", "0000-0002-1825-0097"),
    ],
  },
  {
    id: "u3",
    first_name: "James",
    last_name: "Okonkwo",
    email: "jokonkwo@university.edu",
    status: "ACTIVE",
    roles: [ROLE_AUDITOR],
    identities: [identity("u3", "cilogon", "cilogon-uid-3003", "jokonkwo@cilogon.org")],
  },
  {
    id: "u4",
    first_name: "Priya",
    last_name: "Sharma",
    email: "psharma@hpc-lab.org",
    status: "SUSPENDED",
    roles: [ROLE_OPERATOR, ROLE_AUDITOR],
    identities: [
      identity("u4", "access", "psharma-access-004", "psharma@hpc-lab.org"),
      identity("u4", "cilogon", "cilogon-uid-4004"),
      identity("u4", "nairr", "nairr-uid-4004"),
    ],
  },
  {
    id: "u5",
    first_name: "Daniel",
    last_name: "Wu",
    email: "dwu@custos-hpc.io",
    status: "INACTIVE",
    identities: [],
    roles: [],
  },
];
