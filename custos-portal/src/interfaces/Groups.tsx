export interface GroupMembership {
  id: string;
  name: string;
  created_time: string;
  last_modified_time: string;
  description: string;
  owner_id: string;
  num_members?: number;
  your_role?: string;
};

export interface Member {
  username: string;
  email: string;
  first_name: string;
  last_name: string;
  created_at: string;
  last_modified_at: string;
  membership_type: string;
}

export interface Group {
  id: string;
  name: string;
  realm_roles?: string[];
  client_roles?: string[];
  parent_id?: string;
  created_time: string;
  last_modified_time: string;
  attributes?: Attribute[];
  description: string;
  owner_id: string;
}

export interface Attribute {
  id: number;
  key: string;
  value: string[];
}