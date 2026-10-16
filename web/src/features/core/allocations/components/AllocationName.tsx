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

import { cn } from "@/lib/utils";
import { isGone } from "@/shared/api/useOnGone";
import { Link } from "@/shared/ui/link";
import { useAllocation } from "../queries";

// Links an allocation by name; a known name skips the lookup, and the raw id is never shown.
export function AllocationName({
  id,
  name,
  className,
}: {
  id: string | undefined;
  name?: string;
  className?: string;
}) {
  const query = useAllocation(name ? undefined : id);
  const label = name || query.data?.name || (isGone(query.error) ? "Unknown allocation" : "—");
  return (
    <Link
      href={`/allocations/${id}`}
      className={cn("[overflow-wrap:anywhere]", className ?? "hover:underline")}
    >
      {label}
    </Link>
  );
}
