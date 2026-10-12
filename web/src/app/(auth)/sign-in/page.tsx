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

import { auth } from "@/shared/auth/auth";
import { sameOriginUrl } from "@/shared/auth/sameOrigin";
import { headers } from "next/headers";
import { redirect } from "next/navigation";

type SearchParams = Promise<{ error?: string; callbackUrl?: string }>;

// There is no custom sign-in screen: the landing page is the single entry
// point. On an auth error we send the signed-out user back to the landing,
// which shows the reason as a banner; a signed-in Custos user goes straight to the callback, and
// anyone else to the IdP.
export default async function SignInPage(props: { searchParams: SearchParams }) {
  const { error, callbackUrl } = await props.searchParams;
  // A sign-in whose Custos user could not be resolved goes back to the IdP, or the notice loops.
  if (!error && (await auth())?.custosUserId) {
    const h = await headers();
    const origin = `${h.get("x-forwarded-proto") ?? "http"}://${h.get("host")}`;
    const { pathname, search, hash } = sameOriginUrl(callbackUrl, origin);
    redirect(pathname + search + hash);
  }
  const suffix = callbackUrl ? `callbackUrl=${encodeURIComponent(callbackUrl)}` : "";
  if (error) {
    redirect(`/?authError=${encodeURIComponent(error)}${suffix && `&${suffix}`}`);
  }
  redirect(`/api/auth/federated-sign-in${suffix && `?${suffix}`}`);
}
