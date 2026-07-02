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

import { Avatar, AvatarFallback } from "@/shared/ui/avatar";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/shared/ui/dropdown-menu";
import { LogOut } from "lucide-react";
import { useSession } from "next-auth/react";
import { useSignOut } from "@/shared/auth/useSignOut";

export function UserPill() {
  const { data: session, status } = useSession();
  const { signOut, isPending } = useSignOut();
  const loading = status === "loading";
  const user = session?.user;
  const name = user?.name ?? user?.email ?? "Signed out";
  const email = user?.email ?? "no session";
  const initial = (user?.email ?? name).slice(0, 1).toUpperCase();

  return (
    <div className="flex items-center gap-3">
      <div className="flex flex-col items-end leading-tight">
        <span className="text-sm font-semibold text-foreground">{loading ? "..." : name}</span>
        <span className="text-xs text-muted-foreground">{email}</span>
      </div>
      <DropdownMenu>
        <DropdownMenuTrigger
          render={(props) => (
            <button
              {...props}
              type="button"
              aria-label="Account menu"
              className="rounded-full outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
            >
              <Avatar className="h-10 w-10 border-2 border-brand">
                <AvatarFallback>{initial}</AvatarFallback>
              </Avatar>
            </button>
          )}
        />
        <DropdownMenuContent align="end" className="w-40">
          <DropdownMenuItem
            disabled={isPending}
            onClick={() => void signOut({ callbackUrl: "/sign-in" })}
          >
            <LogOut className="mr-2 h-4 w-4" />
            {isPending ? "Signing out…" : "Sign out"}
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
}
