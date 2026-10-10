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

import * as React from "react";
import { usePrivilegeCatalog } from "@/features/core/roles/queries";
import { UserName } from "@/features/core/users/components/UserPicker";
import { usePrivilegeHolders } from "@/features/core/users/queries";
import type { UserPrivilege } from "@/generated/core/types.gen";
import { formatDate } from "@/shared/format";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/shared/ui/select";

// Direct grants only; holders through a role show on the role cards above.
export function PrivilegeHoldersCard() {
  const catalog = usePrivilegeCatalog(true).data ?? [];
  const [selected, setSelected] = React.useState("");
  const key = catalog.find((k) => k === selected);
  const holders = usePrivilegeHolders(key);

  return (
    <section className="space-y-3 rounded-md border bg-card p-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h2 className="text-sm font-semibold">Direct privilege holders</h2>
        <Select value={selected} onValueChange={(value) => setSelected(value ?? "")}>
          <SelectTrigger aria-label="Privilege" className="h-9 w-64 px-3">
            <SelectValue placeholder="Pick a privilege" />
          </SelectTrigger>
          <SelectContent>
            {catalog.map((k) => (
              <SelectItem key={k} value={k}>
                {k}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
      {!key ? null : holders.isLoading ? (
        <p className="text-sm text-muted-foreground">Loading holders…</p>
      ) : holders.isError ? (
        <p className="text-sm text-muted-foreground">Could not load holders.</p>
      ) : (holders.data ?? []).length === 0 ? (
        <p className="text-sm text-muted-foreground">No user holds {key} directly.</p>
      ) : (
        <ul className="space-y-1.5 text-sm">
          {(holders.data ?? []).map((grant) => (
            <HolderRow key={grant.id} grant={grant} />
          ))}
        </ul>
      )}
    </section>
  );
}

function HolderRow({ grant }: { grant: UserPrivilege }) {
  return (
    <li className="flex items-center justify-between gap-2">
      <span className="font-medium text-foreground">
        <UserName id={grant.user_id} />
      </span>
      <span className="truncate text-xs text-muted-foreground">
        {[grant.reason, formatDate(grant.granted_at)].filter(Boolean).join(" · ")}
      </span>
    </li>
  );
}
