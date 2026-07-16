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
import { EmptyState } from "@/shared/ui/EmptyState";

export const metadata: Metadata = {
  title: "Access Requests — Admin",
};

// Placeholder: the queue lands in the next phase on this branch.
export default function AdminAccessRequestsPage() {
  return (
    <div className="flex flex-col gap-6">
      <h1 className="font-display text-[28px] font-bold leading-tight">Access Requests</h1>
      <EmptyState
        heading="Review queue coming in this branch"
        description="Pending trial-access requests will be listed here for approval."
      />
    </div>
  );
}
