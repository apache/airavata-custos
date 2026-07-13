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

import {
  Activity,
  Building2,
  ClipboardList,
  FolderKanban,
  HardDrive,
  type LucideIcon,
  Server,
  UserCog,
} from "lucide-react";

export type AbilityCheck = { action: string; subject: string };

export type NavGroup = "allocations" | "admin";

export type NavItem = {
  href: string;
  label: string;
  icon: LucideIcon;
  group: NavGroup;
  ability?: AbilityCheck;
};

export const NAV_GROUP_LABELS: Record<NavGroup, string> = {
  allocations: "My allocations",
  admin: "Site administration",
};

export const portalNav: NavItem[] = [
  {
    href: "/allocations",
    label: "Allocations",
    icon: Server,
    group: "allocations",
    ability: { action: "read", subject: "Allocation" },
  },
  {
    href: "/projects",
    label: "Projects",
    icon: FolderKanban,
    group: "allocations",
    ability: { action: "read", subject: "Project" },
  },
  {
    href: "/admin/users",
    label: "Users & Permissions",
    icon: UserCog,
    group: "admin",
    ability: { action: "read", subject: "User" },
  },
  {
    href: "/admin/traces",
    label: "Tracing",
    icon: Activity,
    group: "admin",
    ability: { action: "read", subject: "Trace" },
  },
  {
    href: "/admin/amie",
    label: "AMIE",
    icon: ClipboardList,
    group: "admin",
    ability: { action: "read", subject: "AMIE" },
  },
  {
    href: "/admin/organizations",
    label: "Organizations",
    icon: Building2,
    group: "admin",
    ability: { action: "read", subject: "Organization" },
  },
  {
    href: "/admin/resources",
    label: "Resources",
    icon: HardDrive,
    group: "admin",
    ability: { action: "read", subject: "Cluster" },
  },
];
