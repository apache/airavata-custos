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

import { useOrganization } from "@/features/core/organizations/queries";
import { ROUTE_LABELS } from "@/shared/layout/nav";
import { ErrorState } from "@/shared/ui/ErrorState";
import { CenteredSpinner } from "@/shared/ui/Loading";
import { useMe, useMyAccess, useMyIdentities } from "../queries";
import { AccessCard } from "./AccessCard";
import { AppearanceCard } from "./AppearanceCard";
import { IdentitiesCard } from "./IdentitiesCard";
import { ProfileCard } from "./ProfileCard";

export function SettingsPage() {
  const me = useMe();
  // The backend user id from /me is authoritative for sub-resource reads;
  // session.user.id may be an email fallback.
  const user = me.data?.user;
  const userId = user?.id;
  // /users/{id}/user-identities is admin-gated and /me carries no identities.
  const showsIdentities = me.data?.privileges?.includes("core:users:read") ?? false;
  const identities = useMyIdentities(showsIdentities ? userId : undefined);
  const canReadOrganization = me.data?.privileges?.includes("core:organizations:read") ?? false;
  const organization = useOrganization(canReadOrganization ? user?.organization_id : undefined);
  const access = useMyAccess(userId, me.data?.privileges ?? [], me.data?.roles ?? []);

  return (
    <div className="mx-auto w-full max-w-[1080px] px-6 py-8">
      <header className="mb-7">
        <h1 className="font-heading text-2xl font-bold tracking-tight">
          {ROUTE_LABELS["/settings"]}
        </h1>
        <p className="text-muted-foreground">
          {showsIdentities
            ? "Your profile, linked identities, access, and appearance."
            : "Your profile, access, and appearance."}
        </p>
      </header>

      {me.isPending ? (
        <CenteredSpinner label="Loading your settings" />
      ) : me.isError || !user ? (
        <ErrorState message="Could not load your profile." onRetry={() => me.refetch()} />
      ) : (
        <div className="space-y-6">
          <ProfileCard
            user={user}
            organizationName={organization.data?.name}
            showsIdentities={Boolean(identities.data)}
          />
          {identities.data ? <IdentitiesCard identities={identities.data} /> : null}
          {access.data ? <AccessCard access={access.data} /> : null}
          <AppearanceCard />
        </div>
      )}
    </div>
  );
}
