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

import Link from "next/link";
import * as React from "react";
import { toast } from "sonner";

const HEADING = "No portal access";
const BODY = "Your account has no privileges yet. Ask an administrator to grant you a role.";

// Shown in place of the portal content when the signed-in user holds no
// privileges. Announces the reason as a toast and keeps a durable inline
// message, so the state does not vanish with the toast.
export function NoAccessNotice() {
  React.useEffect(() => {
    toast.error(HEADING, { description: BODY });
  }, []);
  return (
    <div className="flex flex-col gap-2">
      <h1 className="text-2xl font-bold tracking-tight">{HEADING}</h1>
      <p className="text-sm text-muted-foreground">{BODY}</p>
      <Link href="/sign-in" className="mt-2 text-sm font-medium text-brand">
        Sign in as a different user
      </Link>
    </div>
  );
}
