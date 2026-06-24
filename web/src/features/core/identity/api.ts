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

import { apiFetch } from "@/shared/api/client";
import { auth } from "@/shared/auth/auth";
import { privilegesResponseSchema } from "./schemas";
import type { CurrentUser, Privilege } from "./types";

export async function getCurrentUser(): Promise<CurrentUser | null> {
  const session = await auth();
  if (!session?.user) return null;
  return {
    id: session.user.id ?? session.user.email ?? "",
    email: session.user.email ?? "",
    name: session.user.name ?? session.user.email ?? "",
    privileges: session.privileges ?? [],
  };
}

export async function getPrivileges(): Promise<Privilege[]> {
  const raw = await apiFetch("/user/privileges");
  return privilegesResponseSchema.parse(raw);
}
