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

import { Button } from "@/shared/ui/button";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { useRouter } from "next/navigation";
import { Breadcrumbs } from "./Breadcrumbs";
import { SidebarMenuButton } from "./Sidebar";
import { UserPill } from "./UserPill";
import { useNavHistory } from "./useNavHistory";

function NavArrows() {
  const router = useRouter();
  const { canGoBack, canGoForward } = useNavHistory();
  const className =
    "h-8 w-8 rounded-full text-muted-foreground hover:bg-muted/40 hover:text-foreground";
  return (
    <div className="flex items-center gap-1">
      <Button
        variant="ghost"
        size="icon"
        aria-label="Go back"
        disabled={!canGoBack}
        focusableWhenDisabled
        className={className}
        onClick={() => router.back()}
      >
        <ChevronLeft className="h-4 w-4" />
      </Button>
      <Button
        variant="ghost"
        size="icon"
        aria-label="Go forward"
        disabled={!canGoForward}
        focusableWhenDisabled
        className={className}
        onClick={() => router.forward()}
      >
        <ChevronRight className="h-4 w-4" />
      </Button>
    </div>
  );
}

export function Topbar({ breadcrumbs }: { breadcrumbs: boolean }) {
  return (
    <header className="sticky top-0 z-40 flex h-16 items-center justify-between gap-2 border-b border-border bg-background px-4 md:gap-4 md:px-8">
      <div className="flex min-w-0 flex-1 items-center gap-2 md:gap-4">
        <SidebarMenuButton />
        <NavArrows />
        {breadcrumbs ? <Breadcrumbs /> : null}
      </div>

      <div className="flex shrink-0 items-center gap-4">
        <UserPill />
      </div>
    </header>
  );
}
