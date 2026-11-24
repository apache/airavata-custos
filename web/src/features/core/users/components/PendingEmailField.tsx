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

import { useAbility } from "@/shared/casl/AbilityProvider";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { useUpdatePendingUserEmail } from "../queries";

// First sign-in matches a pending user by email, so an admin may correct it until then.
export function PendingEmailField({ userId, email }: { userId: string; email: string }) {
  const canEdit = useAbility().can("manage", "User");
  const update = useUpdatePendingUserEmail();
  if (!canEdit) return null;
  return (
    <>
      <dt className="text-muted-foreground">Email</dt>
      <dd>
        <form
          className="flex items-center gap-2"
          onSubmit={(event) => {
            event.preventDefault();
            const next = new FormData(event.currentTarget).get("email");
            update.mutate({ userId, email: String(next).trim() });
          }}
        >
          <Input
            name="email"
            type="email"
            aria-label="Email"
            className="h-8"
            defaultValue={email}
          />
          <Button type="submit" variant="outline" size="sm" disabled={update.isPending}>
            Save
          </Button>
        </form>
        {update.error ? (
          <p role="alert" className="mt-1 text-xs text-destructive">
            {update.error.message}
          </p>
        ) : null}
      </dd>
    </>
  );
}
