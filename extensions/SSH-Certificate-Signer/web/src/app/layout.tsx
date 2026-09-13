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

import type { Metadata } from "next";
import { redirect } from "next/navigation";
import type { ReactNode } from "react";
import { serverEnv } from "@/lib/env";
import { getSharedSession } from "@/shared/auth/session";
import { Providers } from "@/shared/providers/Providers";
import "./globals.css";

export const metadata: Metadata = {
  title: "SSH Certificate Signer",
  description: "Administer SSH certificates issued by Apache Custos",
};

export default async function RootLayout({ children }: { children: ReactNode }) {
  const session = await getSharedSession();
  if (!session) {
    redirect(`${serverEnv.CUSTOS_PORTAL_BASE_URL}/sign-in`);
  }
  const canRead = session.privileges.includes("signer:certificates:read");

  return (
    <html lang="en" className="light h-full antialiased">
      <body className="min-h-full bg-background text-foreground">
        <Providers privileges={session.privileges}>
          <div className="min-h-screen">
            <header className="border-b border-border bg-card">
              <div className="mx-auto flex max-w-7xl items-center justify-between px-6 py-4">
                <div>
                  <p className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                    Custos extension
                  </p>
                  <p className="font-display text-lg font-bold">SSH Certificate Signer</p>
                </div>
                <div className="flex items-center gap-4 text-sm">
                  <span className="hidden text-muted-foreground sm:inline">
                    {session.name ?? session.email ?? "Signed in"}
                  </span>
                  <a
                    className="font-semibold text-brand hover:underline"
                    href={serverEnv.CUSTOS_PORTAL_BASE_URL}
                  >
                    Back to Custos
                  </a>
                </div>
              </div>
            </header>
            <main id="main-content" className="mx-auto max-w-7xl px-6 py-8">
              {canRead ? children : <Forbidden />}
            </main>
          </div>
        </Providers>
      </body>
    </html>
  );
}

function Forbidden() {
  return (
    <section className="rounded-lg border border-border bg-card p-8 text-center">
      <h1 className="font-display text-2xl font-bold">Access denied</h1>
      <p className="mt-2 text-sm text-muted-foreground">
        You are not permitted to administer SSH certificates.
      </p>
    </section>
  );
}
