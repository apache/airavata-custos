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

// Permission sections mirror the backend's privilege scopes (domain:resource,
// e.g. "core:allocations"), so each section can be granted read/write
// independently. Displayed verbatim as the effective-privilege key.
export const PERMISSION_SECTIONS = [
  "amie:packets",
  "amie:replies",
  "amie:unmapped",
  "core:allocations",
  "core:clusters",
  "core:organizations",
  "core:projects",
  "core:traces",
  "core:users",
  "temp-account:accounts",
] as const;
export type PermissionSection = (typeof PERMISSION_SECTIONS)[number];
export type PermissionKey = `${PermissionSection}:read` | `${PermissionSection}:write`;

export function rwStateFor(permissions: PermissionKey[]) {
  const held = new Set(permissions);
  return PERMISSION_SECTIONS.map((section) => ({
    section,
    read: held.has(`${section}:read` as PermissionKey),
    write: held.has(`${section}:write` as PermissionKey),
  }));
}

// Write implies read (you can't write what you can't read), so granting
// write also grants read, and revoking read also revokes write.
export function togglePermission(permissions: PermissionKey[], key: PermissionKey): PermissionKey[] {
  const held = new Set(permissions);
  const isRead = key.endsWith(":read");
  const pairedKey = (
    isRead ? key.replace(/:read$/, ":write") : key.replace(/:write$/, ":read")
  ) as PermissionKey;

  if (held.has(key)) {
    held.delete(key);
    if (isRead) held.delete(pairedKey);
  } else {
    held.add(key);
    if (!isRead) held.add(pairedKey);
  }
  return Array.from(held);
}

// A user's effective permissions are the union of everything granted by
// each role they hold.
export function permissionsFromRoles(roles: Array<{ permissions: PermissionKey[] }>): PermissionKey[] {
  return Array.from(new Set(roles.flatMap((r) => r.permissions)));
}
