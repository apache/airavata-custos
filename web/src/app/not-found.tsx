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
import { brand } from "@/shared/brand";
import { PortalLayout } from "@/shared/layout/PortalLayout";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ViewParamsDropped } from "@/shared/ui/ViewParamsDropped";
import { buttonVariants } from "@/shared/ui/button";
import { Link } from "@/shared/ui/link";
import type { Metadata } from "next";

export const metadata: Metadata = { title: `Page not found · ${brand.name}` };

// Unknown routes keep the portal chrome for a signed-in Custos user, without breadcrumbs
// built from a path that has no pages.
export default async function NotFound() {
  const session = await auth();
  const notice = (
    <>
      <ViewParamsDropped />
      <EmptyState
        as="h1"
        heading="Page not found"
        description="This page doesn't exist or has moved."
        cta={
          <Link href={session?.custosUserId ? "/allocations" : "/"} className={buttonVariants()}>
            Back to home
          </Link>
        }
      />
    </>
  );
  return session?.custosUserId ? (
    <PortalLayout breadcrumbs={false}>{notice}</PortalLayout>
  ) : (
    <main>{notice}</main>
  );
}
