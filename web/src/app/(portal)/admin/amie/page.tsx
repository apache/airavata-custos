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

import { redirect } from "next/navigation";
import { auth } from "@/shared/auth/auth";
import { defineAbilitiesFor } from "@/shared/casl/abilities";
import { AMIE_TABS } from "./tabs";

// /admin/amie has no content of its own; bounce to the first tab the caller can read.
export default async function AmieIndex() {
  const ability = defineAbilitiesFor((await auth())?.privileges ?? []);
  const first = AMIE_TABS.find((tab) => ability.can("read", tab.subject));
  redirect(first?.href ?? "/admin/amie/packets");
}
