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

import type * as React from "react";

const VISIBLE_COUNT = 2;
const toggleClass =
  "text-xs font-medium text-muted-foreground underline-offset-2 hover:text-foreground hover:underline";

// A table cell of badges that shows the first two until expanded; `label` names the status lines.
export function BadgesCell<T>({
  items,
  label,
  isLoading,
  hasError,
  expanded,
  onToggleExpand,
  className,
  children,
}: {
  items: T[];
  label: string;
  isLoading: boolean;
  hasError: boolean;
  expanded: boolean;
  onToggleExpand: () => void;
  className: string;
  children: (item: T) => React.ReactNode;
}) {
  const visible = expanded ? items : items.slice(0, VISIBLE_COUNT);
  const hiddenCount = items.length - visible.length;
  const status = isLoading
    ? `Loading ${label.toLowerCase()}…`
    : hasError
      ? `${label} unavailable`
      : items.length === 0
        ? `No ${label.toLowerCase()}`
        : null;

  return (
    <div className={`flex flex-wrap items-center gap-1.5 ${className}`}>
      {status ? <span className="text-sm text-muted-foreground">{status}</span> : visible.map(children)}
      {hiddenCount > 0 || (expanded && items.length > VISIBLE_COUNT) ? (
        <button type="button" onClick={onToggleExpand} className={toggleClass}>
          {hiddenCount > 0 ? `+${hiddenCount}` : "Show less"}
        </button>
      ) : null}
    </div>
  );
}
