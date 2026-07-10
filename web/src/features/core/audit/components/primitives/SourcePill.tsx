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

export type SourcePillProps = {
  source: string;
  size?: "sm" | "md";
  className?: string;
};

const SOURCE_STYLES: Record<string, string> = {
  amie: "bg-[color:var(--tone-info-bg)] text-[color:var(--tone-info-fg)]",
  comanage: "bg-[color:var(--tone-accent-bg)] text-[color:var(--tone-accent-fg)]",
  slurm: "bg-[color:var(--tone-warn-bg)] text-[color:var(--tone-warn-fg)]",
  http: "bg-muted text-muted-foreground",
  core: "bg-muted text-muted-foreground",
};

// Unknown sources fall through to neutral muted — never crashes when the
// backend ships a new connector before the portal updates.
export function SourcePill({ source, size = "md", className }: SourcePillProps) {
  const key = source.toLowerCase();
  const styles = SOURCE_STYLES[key] ?? "bg-muted text-muted-foreground";
  return (
    <span
      className={cn(
        "inline-flex items-center rounded-md font-semibold tracking-tight whitespace-nowrap leading-none",
        size === "sm" ? "h-5 px-1.5 text-[11px]" : "h-5 px-2 text-[11.5px]",
        styles,
        className,
      )}
    >
      {key}
    </span>
  );
}
