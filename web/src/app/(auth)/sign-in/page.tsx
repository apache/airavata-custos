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

import { redirect } from "next/navigation";

type SearchParams = Promise<{ error?: string; callbackUrl?: string }>;

// There is no custom sign-in screen: the landing page is the single entry
// point. On an auth error we send the signed-out user back to the landing,
// which shows the reason as a banner; otherwise straight to the IdP.
export default async function SignInPage(props: { searchParams: SearchParams }) {
  const { error, callbackUrl } = await props.searchParams;
  if (error) {
    redirect(`/?authError=${encodeURIComponent(error)}`);
  }
  return (
    <div className="flex min-h-screen items-center justify-center bg-muted px-6 py-12">
      <div className="w-full max-w-md rounded-2xl border border-border bg-card p-8 shadow-sm">
        <div className="mb-6 flex flex-col items-start gap-2">
          <img src="/custos-logo.svg" alt="Custos" className="h-8 w-auto" />
          <h1 className="text-xl font-semibold tracking-tight">Sign in</h1>
          <p className="text-sm text-muted-foreground">Sign in with your Custos account.</p>
        </div>
        <Suspense fallback={null}>
          <SignInForm />
        </Suspense>
      </div>
    </div>
  );
}
