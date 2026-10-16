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

import { usePathname, useSearchParams } from "next/navigation";
import * as React from "react";
import { toast } from "sonner";

const HEADING = "No portal access";
const BODY =
  "Your account could not be looked up; the server may be unreachable. Sign in again to retry.";
// Ending the session first makes the IdP ask again instead of returning the same sign-in; the new
// sign-in comes back to the page that showed the notice.
const resignIn = (here: string) =>
  `/api/auth/end-session?callbackUrl=${encodeURIComponent(`/sign-in?callbackUrl=${encodeURIComponent(here)}`)}`;

// Shown when /me could not be reached at sign-in (a refused sign-in is redirected by the portal
// layout). The message also stays on the page, since a toast alone disappears.
export function NoAccessNotice() {
  const search = useSearchParams().toString();
  const here = `${usePathname()}${search && `?${search}`}`;
  React.useEffect(() => {
    toast.error(HEADING, { description: BODY });
  }, []);
  return (
    <div className="flex flex-col gap-2">
      <h1 className="text-2xl font-bold tracking-tight">{HEADING}</h1>
      <p className="text-sm text-muted-foreground">{BODY}</p>
      <a href={resignIn(here)} className="mt-2 text-sm font-medium text-link">
        Sign in again
      </a>
    </div>
  );
}
