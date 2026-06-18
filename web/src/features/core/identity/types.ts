import type { PrivilegeKey } from "./schemas";

export type { PrivilegeKey };
export type Privilege = PrivilegeKey;

export type CurrentUser = {
  id: string;
  email: string;
  name: string;
  privileges: Privilege[];
};
