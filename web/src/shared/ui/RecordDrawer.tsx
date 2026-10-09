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

import type { UseQueryResult } from "@tanstack/react-query";
import type * as React from "react";
import { Button } from "@/shared/ui/button";
import { ErrorState } from "@/shared/ui/ErrorState";
import { CardSkeleton } from "@/shared/ui/Loading";
import { SideDrawer, type SideDrawerWidth } from "@/shared/ui/SideDrawer";
import { confirmToast } from "@/shared/ui/sonner";

// A side drawer over one record fetched by id: a label/value list, extra children, and
// an optional confirmed delete. The drawer is open while the caller holds an id.
export function RecordDrawer<T>({
  title,
  id,
  onClose,
  query,
  fields,
  children,
  remove,
  width,
}: {
  title: string;
  id: string | undefined;
  onClose: () => void;
  query: UseQueryResult<T>;
  fields: (record: T) => Record<string, React.ReactNode>;
  children?: (record: T) => React.ReactNode;
  remove?: { label: string; confirm: string; isPending: boolean; onConfirm: () => void };
  width?: SideDrawerWidth;
}) {
  const record = query.data;
  return (
    <SideDrawer
      open={Boolean(id)}
      onOpenChange={(open) => (open ? null : onClose())}
      title={title}
      width={width}
    >
      {query.isLoading ? (
        <CardSkeleton />
      ) : query.error ? (
        <ErrorState message={query.error.message} onRetry={() => query.refetch()} />
      ) : id && record ? (
        <div className="space-y-6">
          <FieldList fields={fields(record)} />
          {children?.(record)}
          {remove ? (
            <Button
              variant="destructive"
              size="sm"
              disabled={remove.isPending}
              onClick={() => confirmToast(remove.confirm, "Delete", remove.onConfirm)}
            >
              {remove.label}
            </Button>
          ) : null}
        </div>
      ) : null}
    </SideDrawer>
  );
}

export const FieldList = ({ fields }: { fields: Record<string, React.ReactNode> }) => (
  <dl className="grid grid-cols-[max-content_1fr] gap-x-4 gap-y-2 text-sm">
    {Object.entries(fields).map(([label, value]) => (
      <div key={label} className="contents">
        <dt className="text-muted-foreground">{label}</dt>
        <dd>{value}</dd>
      </div>
    ))}
  </dl>
);

export const Mono = ({ children }: { children?: React.ReactNode }) => (
  <span className="font-mono text-xs">{children}</span>
);
