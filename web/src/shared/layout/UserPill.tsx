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

import { useMe } from "@/features/core/identity/queries";
import { useSignOut } from "@/shared/auth/useSignOut";
import { Avatar, AvatarFallback } from "@/shared/ui/avatar";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLinkItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/shared/ui/dropdown-menu";
import { displayName, emailBeside, initials } from "@/shared/users-admin/displayName";
import { LogOut, Settings } from "lucide-react";
import { useSession } from "next-auth/react";

export function UserPill() {
  const { data: session, status } = useSession();
  const { signOut, isPending } = useSignOut();
  const loading = status === "loading";
  const user = session?.user;
  // The session keeps the IdP's name and email from sign-in; /me reflects the account's own.
  const me = useMe().data?.user;
  const name = (me && displayName(me)) || user?.name || user?.email || "Signed out";
  const address = me?.email || user?.email;
  const email = address ? emailBeside(name, address) : "no session";

  return (
    <div className="flex min-w-0 items-center gap-3">
      <div className="hidden min-w-0 flex-col items-end leading-tight lg:flex">
        <span title={name} className="max-w-[24ch] truncate text-sm font-semibold text-foreground">
          {loading ? "..." : name}
        </span>
        {email ? (
          <span title={email} className="max-w-[24ch] truncate text-xs text-muted-foreground">
            {email}
          </span>
        ) : null}
      </div>
      <DropdownMenu>
        <DropdownMenuTrigger
          render={(props) => (
            <button
              {...props}
              type="button"
              aria-label="Account menu"
              className="shrink-0 rounded-full outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
            >
              <Avatar className="h-10 w-10 border-2 border-brand">
                <AvatarFallback>{initials(name)}</AvatarFallback>
              </Avatar>
            </button>
          )}
        />
        <DropdownMenuContent align="end" className="w-64">
          <div className="px-2 py-1.5">
            <div title={name} className="truncate text-sm font-semibold text-foreground">
              {name}
            </div>
            {email ? (
              <div title={email} className="truncate text-xs text-muted-foreground">
                {email}
              </div>
            ) : null}
          </div>
          <DropdownMenuSeparator />
          <DropdownMenuLinkItem href="/settings" closeOnClick>
            <Settings className="mr-2 h-4 w-4" />
            Settings
          </DropdownMenuLinkItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem
            variant="destructive"
            disabled={isPending}
            onClick={() => void signOut()}
          >
            <LogOut className="mr-2 h-4 w-4" />
            {isPending ? "Signing out…" : "Sign out"}
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
}
