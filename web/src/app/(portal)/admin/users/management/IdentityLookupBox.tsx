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

import { type IdentityLookup, lookupIdentity } from "@/features/core/users/queries";
import { errorCopy } from "@/shared/api/client";
import { queryUuid } from "@/shared/layout/nav";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/shared/ui/select";
import { idInput } from "@/shared/users-admin/idInput";
import { useMutation } from "@tanstack/react-query";
import * as React from "react";
import { IDENTITY_SOURCE_LABELS, identitySourceKey } from "./identities";

const MODES = { external: "Source ID", oidc: "OIDC subject", id: "Identity ID" };
type Mode = keyof typeof MODES;
const MODE_KEYS = Object.keys(MODES) as Mode[];

// Resolves an external identity to its user, which opens in the drawer even when off this page.
export function IdentityLookupBox({ onResolve }: { onResolve: (userId: string) => void }) {
  const [by, setBy] = React.useState<Mode>("external");
  const [source, setSource] = React.useState("");
  const [value, setValue] = React.useState("");
  const find = useMutation({
    mutationFn: async (lookup: IdentityLookup) => {
      const identity = await lookupIdentity(lookup);
      if (!identity) throw new Error("No identity matches that.");
      return identity.user_id;
    },
    onSuccess: onResolve,
    meta: { inlineError: true },
  });

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const term = value.trim();
    find.mutate(
      by === "external"
        ? { source: identitySourceKey(source.trim()), external_id: term }
        : by === "oidc"
          ? { oidc_sub: term }
          : { id: queryUuid(term) ?? term },
    );
  }

  return (
    <form className="flex flex-wrap items-center gap-2" onSubmit={handleSubmit}>
      <Select
        value={by}
        onValueChange={(next) => setBy(MODE_KEYS.find((key) => key === next) ?? "external")}
      >
        <SelectTrigger aria-label="Look up by" className="h-9 w-36 px-3">
          <SelectValue>{(mode: Mode) => MODES[mode]}</SelectValue>
        </SelectTrigger>
        <SelectContent>
          {MODE_KEYS.map((mode) => (
            <SelectItem key={mode} value={mode}>
              {MODES[mode]}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      {by === "external" ? (
        <>
          <Input
            aria-label="Identity source"
            placeholder="Source"
            list="lookup-identity-sources"
            value={source}
            onChange={(e) => setSource(e.target.value)}
            className="w-28"
            required
          />
          <datalist id="lookup-identity-sources">
            {Object.values(IDENTITY_SOURCE_LABELS).map((label) => (
              <option key={label} value={label} />
            ))}
          </datalist>
        </>
      ) : null}
      <Input
        aria-label={`Look up user by ${MODES[by]}`}
        placeholder={MODES[by]}
        value={value}
        onChange={(e) => setValue(e.target.value)}
        className="w-48"
        required
        {...(by === "id" ? idInput("identity") : {})}
      />
      <Button type="submit" variant="outline" pending={find.isPending}>
        Find user
      </Button>
      {find.error ? (
        <span role="alert" className="text-sm text-muted-foreground">
          {errorCopy(find.error)}
        </span>
      ) : null}
    </form>
  );
}
