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

import { cn } from "@/lib/utils";
import * as React from "react";

export function Spinner({ className }: { className?: string }) {
  return (
    <span
      role="status"
      aria-label="Loading"
      className={cn(
        "inline-block size-5 animate-spin rounded-full border-2 border-muted border-t-[color:var(--custos-blue-500)]",
        className,
      )}
    />
  );
}

export function CenteredSpinner({ label }: { label?: string }) {
  return (
    <div className="flex flex-col items-center justify-center gap-2 py-12 text-sm text-muted-foreground">
      <Spinner />
      {label ? <span>{label}</span> : null}
    </div>
  );
}

export function SkeletonRow({ columns = 4, className }: { columns?: number; className?: string }) {
  return (
    <div className={cn("flex animate-pulse gap-4", className)}>
      {Array.from({ length: columns }).map((_, i) => (
        <div key={`skeleton-col-${i}`} className="h-3 flex-1 rounded bg-muted" />
      ))}
    </div>
  );
}

export function TableSkeleton({ rows = 6, columns = 5 }: { rows?: number; columns?: number }) {
  return (
    <div className="space-y-3 rounded-md border bg-card p-4">
      {Array.from({ length: rows }).map((_, i) => (
        <SkeletonRow key={`skeleton-row-${i}`} columns={columns} />
      ))}
    </div>
  );
}

export function CardSkeleton({ className }: { className?: string }) {
  return (
    <div className={cn("animate-pulse space-y-3 rounded-md border bg-card p-5", className)}>
      <div className="h-4 w-32 rounded bg-muted" />
      <div className="h-8 w-24 rounded bg-muted" />
      <div className="h-3 w-48 rounded bg-muted" />
    </div>
  );
}
