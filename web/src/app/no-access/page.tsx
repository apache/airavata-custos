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

import Link from "next/link";

export const metadata = { title: "No access · Custos Portal" };

export default function NoAccessPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-muted px-6 py-12">
      <div className="w-full max-w-md rounded-2xl border border-border bg-card p-8 shadow-sm">
        <h1 className="text-xl font-semibold tracking-tight">No portal access</h1>
        <p className="mt-2 text-sm text-muted-foreground">
          Your account has no privileges yet. Ask an administrator to grant you a role.
        </p>
        <Link href="/sign-in" className="mt-6 inline-block text-sm font-medium text-brand">
          Sign in as a different user
        </Link>
      </div>
    </div>
  );
}
