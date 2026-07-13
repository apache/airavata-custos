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

function Cell({
  label,
  active,
  activeClass,
  onClick,
}: {
  label: string;
  active: boolean;
  activeClass: string;
  onClick?: () => void;
}) {
  const title = label === "read" ? "Read" : "Write";
  const className = cn(
    "inline-flex h-6 items-center justify-center rounded px-2 text-xs font-medium",
    active ? activeClass : "border border-border text-muted-foreground",
    onClick && "cursor-pointer transition-transform hover:scale-105",
  );
  if (!onClick) {
    return (
      <span title={title} className={className}>
        {label}
      </span>
    );
  }
  return (
    <button
      type="button"
      title={title}
      aria-pressed={active}
      aria-label={`${title}${active ? ", granted" : ", not granted"}`}
      onClick={onClick}
      className={className}
    >
      {label}
    </button>
  );
}

export function PermissionRW({
  read,
  write,
  onToggleRead,
  onToggleWrite,
}: {
  read: boolean;
  write: boolean;
  onToggleRead?: () => void;
  onToggleWrite?: () => void;
}) {
  return (
    <div className="flex items-center gap-1">
      <Cell
        label="read"
        active={read}
        activeClass="bg-[color:var(--tone-info-bg)] text-[color:var(--tone-info-fg)]"
        onClick={onToggleRead}
      />
      <Cell
        label="write"
        active={write}
        activeClass="bg-[color:var(--tone-ok-bg)] text-[color:var(--tone-ok-fg)]"
        onClick={onToggleWrite}
      />
    </div>
  );
}
