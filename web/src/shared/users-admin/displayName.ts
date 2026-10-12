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

import type { User } from "@/generated/core/types.gen";

type Named = Pick<User, "first_name" | "middle_name" | "last_name" | "email">;

// A person's name as every view renders it: first, middle and last, else their email.
export const displayName = (user: Named): string =>
  [user.first_name, user.middle_name, user.last_name]
    .map((part) => part?.trim())
    .filter(Boolean)
    .join(" ") ||
  user.email ||
  "";

// The email shown beside a name, unless the name already is that email.
export const emailBeside = (name: string | undefined, email: string | undefined) =>
  email && email !== name ? email : undefined;

// Avatar initials of a display name: its first and last words, else its first two characters.
export function initials(name = ""): string {
  const parts = name.trim().split(/\s+/).filter(Boolean);
  const chars =
    parts.length >= 2
      ? `${parts[0]?.charAt(0)}${parts.at(-1)?.charAt(0)}`
      : name.trim().slice(0, 2);
  return chars.toUpperCase();
}
