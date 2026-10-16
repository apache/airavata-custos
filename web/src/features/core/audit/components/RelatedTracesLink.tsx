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

import { focusHeadingOf } from "@/shared/ui/focus";
import Link from "next/link";
import * as React from "react";
import { entityKind } from "../utils";

// Set by the trace drawer: a link that leaves it tells it not to return focus to its raiser.
export const DrawerLeaveContext = React.createContext<() => void>(() => {});

export function RelatedTracesLink({
  entityId,
  entityType,
  className,
}: { entityId: string; entityType: string | undefined; className: string }) {
  const leave = React.useContext(DrawerLeaveContext);
  const href = `/admin/traces?q=${encodeURIComponent(entityId)}&window=30d`;
  return (
    <Link
      href={href}
      className={className}
      aria-label={
        entityType ? `Related traces for this ${entityKind(entityType).toLowerCase()}` : undefined
      }
      onClick={(e) => {
        if (e.button !== 0 || e.metaKey || e.ctrlKey || e.shiftKey || e.altKey) return;
        leave();
        focusHeadingOf(href);
      }}
    >
      Related traces
    </Link>
  );
}
