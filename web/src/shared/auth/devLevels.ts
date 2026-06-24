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

import type { Privilege } from "@/features/core/identity/types";

export type DevLevel = "viewer" | "manager" | "admin";

// Dev-mode privilege bundles. Levels are coarser than the spec roles so the
// dropdown stays simple; admin covers the full PrivilegeKey enum.
export const DEV_LEVEL_PRIVILEGES: Record<DevLevel, Privilege[]> = {
  viewer: ["hpc:read"],
  manager: ["hpc:read", "hpc:write", "amie:read"],
  admin: [
    "amie:read",
    "amie:write",
    "hpc:read",
    "hpc:write",
    "signer:read",
    "signer:write",
    "privileges:grant",
    "roles:manage",
  ],
};

export const DEV_LEVEL_NAMES: Record<DevLevel, string> = {
  viewer: "Dev Viewer",
  manager: "Dev Manager",
  admin: "Dev Admin",
};
