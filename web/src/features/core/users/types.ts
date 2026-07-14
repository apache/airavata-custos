// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements. See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership. The ASF licenses this file
// to you under the Apache License, Version 2.0.

import type { Role, User, UserIdentity } from "./schemas";

export type UserListParams = {
  limit?: number;
  offset?: number;
};

export type RoleWithPrivileges = Role & {
  privileges: string[];
};

export type UserManagementRow = User & {
  roles: Role[];
  identities: UserIdentity[];
  rolesLoading: boolean;
  identitiesLoading: boolean;
  rolesError: boolean;
  identitiesError: boolean;
};

export type UpdateUserRolesInput = {
  userId: string;
  currentRoleIds: string[];
  desiredRoleIds: string[];
};
