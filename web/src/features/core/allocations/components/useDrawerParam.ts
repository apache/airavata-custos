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

import {
  setSearchParam,
  useCanonicalParams,
  useShallowSearchParams,
} from "@/shared/hooks/useShallowSearchParams";
import { queryUuid } from "@/shared/layout/nav";

// Each detail tab's drawer and the tab that renders it.
const DRAWER_TABS = { membership: "members", usage: "usage", diff: "history" } as const;

// The open record of a tab's drawer lives in ?<key>=, so a reload or a shared link reopens it.
export function useDrawerParam(key: keyof typeof DRAWER_TABS) {
  const params = useShallowSearchParams();
  return [queryUuid(params.get(key)), (id?: string) => setSearchParam(params, key, id)] as const;
}

// Keeps the shown tab's drawer param normalised and drops every other tab's.
export function useDrawerParams(tab: string | null) {
  const params = useShallowSearchParams();
  useCanonicalParams(
    Object.fromEntries(
      Object.entries(DRAWER_TABS).map(([key, owner]) => [
        key,
        owner === tab ? queryUuid(params.get(key)) : undefined,
      ]),
    ),
  );
}
