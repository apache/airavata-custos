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

import { useAbility } from "@/shared/casl/AbilityProvider";
import { Button } from "@/shared/ui/button";
import { ErrorState } from "@/shared/ui/ErrorState";
import { UsersNav } from "../UsersNav";
import { RoleFormDialog } from "./RoleFormDialog";
import { RolesGrid } from "./RolesGrid";

export function RoleManagementView() {
  const ability = useAbility();
  const canManageRoles = ability.can("manage", "Role");

  if (!canManageRoles) {
    return (
      <div className="space-y-6">
        <UsersNav />
        <ErrorState
          heading="Not permitted"
          message="You do not have permission to manage roles."
        />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <UsersNav
        rightSlot={<RoleFormDialog triggerRender={<Button />} triggerContent="Create role" />}
      />
      <RolesGrid />
    </div>
  );
}
