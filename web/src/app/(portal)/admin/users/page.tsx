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

import { auth } from "@/shared/auth/auth";
import { defineAbilitiesFor } from "@/shared/casl/abilities";
import { USERS_TABS, isNavItemVisible } from "@/shared/layout/nav";
import { redirect } from "next/navigation";

// /admin/users has no content of its own; bounce to the first tab the caller can open.
export default async function AdminUsersIndex() {
  const ability = defineAbilitiesFor((await auth())?.privileges ?? []);
  const tab = USERS_TABS.find((t) => isNavItemVisible(t, ability));
  redirect((tab ?? USERS_TABS[0]).href);
}
