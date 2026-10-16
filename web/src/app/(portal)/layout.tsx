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

import { PATH_HEADER } from "@/middleware";
import { auth } from "@/shared/auth/auth";
import { PrivilegesLoaded } from "@/shared/casl/AbilityProvider";
import { NoAccessNotice } from "@/shared/layout/NoAccessNotice";
import { PortalLayout } from "@/shared/layout/PortalLayout";
import { headers } from "next/headers";
import { redirect } from "next/navigation";

const REFUSED = new Set(["user_disabled", "identity_not_linked"]);

export default async function PortalRoutesLayout({ children }: { children: React.ReactNode }) {
  const session = await auth();
  const path = (await headers()).get(PATH_HEADER) ?? "/";
  if (!session?.user) redirect(`/sign-in?callbackUrl=${encodeURIComponent(path)}`);
  // A refused sign-in ends the IdP session too, or SSO would sign straight back in; the sign-in
  // page then shows the reason.
  if (session.authError && REFUSED.has(session.authError)) {
    redirect(
      `/api/auth/end-session?callbackUrl=${encodeURIComponent(`/sign-in?error=${session.authError}`)}`,
    );
  }
  // A user with no privileges gets in, since their pages are limited by membership. A sign-in
  // whose Custos user could not be looked up sees the notice here, as a redirect would loop.
  const hasAccess = Boolean(session.custosUserId);
  return (
    <PortalLayout>
      {hasAccess ? <PrivilegesLoaded>{children}</PrivilegesLoaded> : <NoAccessNotice />}
    </PortalLayout>
  );
}
