export interface User {
  username: string;
  email: string;
  first_name: string;
  last_name: string;
  created_at: string;
  client_roles: string[];
  realm_roles: string[];
  last_modified_at: string;
  attributes: Attribute[];
}

export interface Attribute {
  key: string;
  values: string[];
}

export interface RegisterRequest {
  username: string;
  email: string;
  firstName: string;
  lastName: string;
  status: string;
  attributes: Attribute[];
  clientRoles: string[];
  realmRoles: string[];
  lastModifiedAt: number;
  type: string;
}
