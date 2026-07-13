// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

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
